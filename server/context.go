package server

import (
	"context"
	"report-generation/db/store"
)

type UserCtxKey struct{}

func ContextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, UserCtxKey{}, user)
}

func UserFromContext(ctx context.Context) (*store.User, bool) {
	user, ok := ctx.Value(UserCtxKey{}).(*store.User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}
