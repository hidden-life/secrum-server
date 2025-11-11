package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/hidden-life/secrum-server/internal/ports"
)

type ctxKey int

const (
	ctxUserIDKey ctxKey = iota
	ctxDeviceIDKey
)

// AuthMiddleware validates Bearer access token and injects user/device IDs into context.
func AuthMiddleware(t ports.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				asError(w, http.StatusUnauthorized, "missing or invalid bearer token")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			userID, deviceID, err := t.ValidateAccess(r.Context(), token)
			if err != nil {
				asError(w, http.StatusUnauthorized, "invalid token")
				return
			}

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
