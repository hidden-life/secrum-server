package context

import "context"

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDCtxKey).(string); ok {
		return v
	}

	return ""
}

func DeviceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(DeviceIDCtxKey).(string); ok {
		return v
	}

	return ""
}
