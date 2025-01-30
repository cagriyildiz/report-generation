package server

import (
	"context"
	"report-generation/db/store"
)

type UserCtxKey struct{}

func ContextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, UserCtxKey{}, user)
}
