package auth

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "userID"

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// SetUserIDInContext sets the user ID in the context
func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
