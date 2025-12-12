package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/ports"
)

type ctxKey string

const (
	ctxUserIDKey   ctxKey = "user_id"
	ctxDeviceIDKey ctxKey = "device_id"
)

// AuthMiddleware validates Bearer access token and injects user/device IDs into context.
func AuthMiddleware(t ports.TokenManager, store ports.SessionStore, deviceRepo ports.DeviceRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("AUTH MIDDLEWARE: ", r.Method, r.URL.Path)
			fmt.Println("AUTH HEADER = ", r.Header.Get("Authorization"))
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				asError(w, http.StatusUnauthorized, "missing or invalid bearer token")
				return
			}

			token := ""

			// search in header
			header := r.Header.Get("Authorization")
			if strings.HasPrefix(header, "Bearer ") {
				token = strings.TrimPrefix(header, "Bearer ")
			}

			// search in query (for WS)
			if token == "" {
				token = r.URL.Query().Get("access_token")
			}

			if token == "" {
				asError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			userID, deviceID, err := t.ValidateAccess(r.Context(), token)
			if err != nil {
				asError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// check device exists and is active
			devUUID, err := uuid.Parse(deviceID)
			if err != nil {
				asError(w, http.StatusUnauthorized, "invalid device in token")
				return
			}

			d, err := deviceRepo.GetById(r.Context(), devUUID)
			if err != nil {
				asError(w, http.StatusUnauthorized, "failed to fetch device")
				return
			}

			if d == nil || !d.IsActive {
				asError(w, http.StatusUnauthorized, "device is inactive or not found")
				return
			}

			// check revocation
			if store != nil {
				revoked, err := store.IsDeviceRevoked(r.Context(), deviceID)
				if err != nil {
					asError(w, http.StatusUnauthorized, "failed to verify session state")
					return
				}

				if revoked {
					asError(w, http.StatusUnauthorized, "device session is revoked")
					return
				}
			}

			// update last_seen
			go func() {
				_ = deviceRepo.UpdateLastSeen(context.Background(), devUUID, time.Now().UTC())
			}()

			ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
			ctx = context.WithValue(ctx, ctxDeviceIDKey, deviceID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxUserIDKey).(string); ok {
		return v
	}

	return ""
}

func DeviceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxDeviceIDKey).(string); ok {
		return v
	}

	return ""
}
