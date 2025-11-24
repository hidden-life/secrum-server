package real_time

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeviceConnection struct {
	DeviceID uuid.UUID
	UserID   uuid.UUID
	Ch       chan []byte
}

type DeliveryHub struct {
	log *zap.Logger
	mtx sync.RWMutex

	// deviceID -> connection
	devices map[uuid.UUID]*DeviceConnection
	// userID -> list of deviceIDs
	userIdx map[uuid.UUID][]uuid.UUID
	// groupID -> list of userIDs
	groupIdx map[uuid.UUID][]uuid.UUID
}

func NewDeliveryHub(log *zap.Logger) *DeliveryHub {
	return &DeliveryHub{
		log:      log,
		devices:  make(map[uuid.UUID]*DeviceConnection),
		userIdx:  make(map[uuid.UUID][]uuid.UUID),
		groupIdx: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (h *DeliveryHub) Register(deviceID, userID uuid.UUID, ch chan []byte) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.devices[deviceID] = &DeviceConnection{
		DeviceID: deviceID,
		UserID:   userID,
		Ch:       ch,
	}

	h.userIdx[userID] = append(h.userIdx[userID], deviceID)

	h.log.Info("device registered in realtime hub", zap.String("device_id", deviceID.String()), zap.String("user_id", userID.String()))
}

func (h *DeliveryHub) Unregister(deviceID uuid.UUID) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	conn, isOK := h.devices[deviceID]
	if !isOK {
		return
	}
	delete(h.devices, deviceID)

	oldList := h.userIdx[conn.UserID]
	newList := make([]uuid.UUID, 0, len(oldList))
	for _, d := range oldList {
		if d != deviceID {
			newList = append(newList, d)
		}
	}
	h.userIdx[conn.UserID] = newList

	h.log.Info("device unregistered from realtime hub", zap.String("device_id", deviceID.String()), zap.String("user_id", conn.UserID.String()))
}

func (h *DeliveryHub) PushToDevice(_ context.Context, deviceID uuid.UUID, raw []byte) error {
	h.mtx.RLock()
	conn, isOK := h.devices[deviceID]
	defer h.mtx.RUnlock()

	if !isOK {
		return fmt.Errorf("device is offline")
	}

	select {
	case conn.Ch <- raw:
		return nil
	default:
		return fmt.Errorf("device channel is full")
	}
}

func (h *DeliveryHub) PushToUser(ctx context.Context, userID uuid.UUID, raw []byte) error {
	h.mtx.RLock()
	ids := h.userIdx[userID]
	h.mtx.RUnlock()

	if len(ids) == 0 {
		return fmt.Errorf("user is offline")
	}

	for _, dev := range ids {
		_ = h.PushToDevice(ctx, dev, raw)
	}

	return nil
}

func (h *DeliveryHub) PushToGroup(ctx context.Context, groupID uuid.UUID, raw []byte) error {
	h.mtx.RLock()
	users := h.groupIdx[groupID]
	h.mtx.RUnlock()

	for _, userID := range users {
		_ = h.PushToUser(ctx, userID, raw)
	}

	return nil
}

func (h *DeliveryHub) Broadcast(ctx context.Context, raw []byte) error {
	h.mtx.RLock()
	for dev := range h.devices {
		_ = h.PushToDevice(ctx, dev, raw)
	}
	h.mtx.RUnlock()

	return nil
}

func (h *DeliveryHub) PushToContacts(ctx context.Context, userID uuid.UUID, raw []byte) error {
	// todo: change using contacts list from database
	return h.Broadcast(ctx, raw)
}

func (h *DeliveryHub) SetGroupMembers(gid uuid.UUID, users []uuid.UUID) {
	h.mtx.Lock()
	h.groupIdx[gid] = users
	h.mtx.Unlock()
}
