package app

import (
	"context"

	"github.com/nahue/setlist_manager/internal/app/shared/types"
)

// UserContextKey is the key used to store user in request context
type UserContextKey struct{}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *types.User {
	if user, ok := ctx.Value(UserContextKey{}).(*types.User); ok {
		return user
	}
	return nil
}
