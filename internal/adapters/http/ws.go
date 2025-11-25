package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hidden-life/secrum-server/internal/app/messages"
	"github.com/hidden-life/secrum-server/internal/presence"
	"github.com/hidden-life/secrum-server/internal/real_time"
	"go.uber.org/zap"
)

type WSCodec interface {
	Encode(any) ([]byte, error)
	Decode([]byte, any) error
}

type wsAckPayload struct {
	Delivered []string `json:"delivered,omitempty"`
	Read      []string `json:"read,omitempty"`
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type JSONCodec struct{}

func (JSONCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func RegisterWSRoutes(r chi.Router, log *zap.Logger, hub *real_time.DeliveryHub, msgSvc *messages.Service, presenceSvc *presence.Service) {
	codec := JSONCodec{} // now JSON, but in future it will be easy to change

	r.Group(func(r chi.Router) {
		// use auth middleware
		r.Get("/ws", wsHandler(log, hub, msgSvc, codec, presenceSvc))
	})
}

func wsHandler(
	log *zap.Logger,
	hub *real_time.DeliveryHub,
	msgSvc *messages.Service,
	codec WSCodec,
	presenceSvc *presence.Service,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		deviceID := DeviceIDFromContext(r.Context())
		if userID == "" || deviceID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		devUUID, err := uuid.Parse(deviceID)
		if err != nil {
			asError(w, http.StatusBadRequest, "invalid device id")
			return
		}

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Warn("websocket upgrade failed", zap.Error(err))
			return
		}
		defer conn.Close()

		log.Info("websocket connected",
			zap.String("user_id", userID),
			zap.String("device_id", deviceID),
		)

		// Channel of output messages from hub -> current WS
		outCh := make(chan []byte, 32)

		// Register connection into the HUB
		userUUID, _ := uuid.Parse(userID)
		hub.Register(devUUID, userUUID, outCh)
		defer hub.Unregister(devUUID)
		defer close(outCh)

		// presence: mark user/device online
		if err := presenceSvc.SetOnline(r.Context(), userID); err != nil {
			log.Warn("failed to mark presence online", zap.Error(err))
		}

		// send status over the hub
		status := real_time.EventStatus{
			UserID: userID,
			Status: "online",
		}
		raw, err := real_time.MarshalEvent("status", status)

		if err == nil {
			_ = hub.Broadcast(r.Context(), raw)
		}

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// First sync HTTP -> WS
		if err := sendInitialSync(ctx, conn, codec, msgSvc, deviceID); err != nil {
			log.Warn("failed to send initial sync", zap.Error(err))
		}

		// Socket writer
		go wsWriter(ctx, log, conn, codec, outCh, presenceSvc, userID)

		// Socket reader
		wsReader(ctx, log, conn, codec, msgSvc, userID, deviceID, hub)

		defer func() {
			if err := presenceSvc.SetOffline(context.Background(), userID); err != nil {
				log.Warn("failed to mark presence offline", zap.Error(err))
			}

			status := real_time.EventStatus{
				UserID: userID,
				Status: "offline",
			}

			raw, err := real_time.MarshalEvent("status", status)
			if err == nil {
				_ = hub.Broadcast(context.Background(), raw)
			}
		}()
	}
}

func sendInitialSync(
	ctx context.Context,
	conn *websocket.Conn,
	codec WSCodec,
	msgSvc *messages.Service,
	deviceID string,
) error {
	msgs, err := msgSvc.FetchPending(ctx, deviceID, 500)
	if err != nil {
		return err
	}

	env := real_time.OutEnvelope{
		Type: "sync",
		Data: msgs,
	}

	payload, err := codec.Encode(env)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, payload)
}

func wsWriter(
	ctx context.Context,
	log *zap.Logger,
	conn *websocket.Conn,
	codec WSCodec,
	outCh <-chan []byte,
	presenceSvc *presence.Service,
	userID string,
) {
	ticker := time.NewTicker(30 * time.Second) // heartbeat
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case raw, ok := <-outCh:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
				log.Warn("failed to write ws message", zap.Error(err))
				return
			}

		case <-ticker.C:
			if err := presenceSvc.Refresh(ctx, userID); err != nil {
				log.Warn("failed to refresh presence", zap.Error(err))
			}

			env := real_time.OutEnvelope{
				Type: "pong",
				Data: time.Now().UTC().Format(time.RFC3339Nano),
			}
			payload, err := codec.Encode(env)
			if err != nil {
				log.Warn("failed to encode ws pong", zap.Error(err))
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Warn("failed to write ws pong", zap.Error(err))
				return
			}
		}
	}
}

func wsReader(
	ctx context.Context,
	log *zap.Logger,
	conn *websocket.Conn,
	codec WSCodec,
	msgSvc *messages.Service,
	userID, deviceID string,
	hub *real_time.DeliveryHub,
) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Info("ws read closed", zap.Error(err))
			return
		}

		var env real_time.InEnvelope
		if err := codec.Decode(data, &env); err != nil {
			log.Warn("failed to decode ws incoming message", zap.Error(err))
			continue
		}

		switch env.Type {
		case "ping":
			out := real_time.OutEnvelope{
				Type: "pong",
				Data: time.Now().UTC().Format(time.RFC3339Nano),
			}
			payload, err := codec.Encode(out)
			if err != nil {
				log.Warn("failed to encode ws pong", zap.Error(err))
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Warn("failed to write ws pong", zap.Error(err))
				return
			}

		case "ack", "ack_delivered", "ack_read":
			var payload wsAckPayload
			if len(env.Data) > 0 {
				if err := json.Unmarshal(env.Data, &payload); err != nil {
					log.Warn("failed to decode ack payload", zap.Error(err))
					continue
				}
			}

			req := messages.AckRequest{
				Delivered: payload.Delivered,
				Read:      payload.Read,
			}

			if err := msgSvc.AckDelivered(ctx, deviceID, req); err != nil {
				log.Warn("failed to process ack", zap.Error(err))
			}

		case "typing":
			var req real_time.EventTyping
			if len(env.Data) > 0 {
				if err := json.Unmarshal(env.Data, &req); err != nil {
					log.Warn("failed to decode typing payload", zap.Error(err))
					continue
				}
			}

			req.UserID = userID // who's typing

			if !req.IsGroup { // p2p typing
				peerUUID, err := uuid.Parse(req.PeerID)
				if err != nil {
					continue
				}

				raw, err := real_time.MarshalEvent("typing", req)
				if err == nil {
					_ = hub.PushToUser(ctx, peerUUID, raw)
				}
			} else { // group typing
				gid, err := uuid.Parse(req.PeerID)
				if err != nil {
					continue
				}

				raw, err := real_time.MarshalEvent("typing", req)
				if err == nil {
					_ = hub.PushToGroup(ctx, gid, raw)
				}
			}

		default:
			log.Debug("unknown ws message type", zap.String("type", env.Type))
		}
	}
}
