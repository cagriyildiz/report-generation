package store

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUserStore(t *testing.T) {
	testDB := NewTestDB(t)
	cleanup := testDB.Setup(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	now := time.Now()
	ctx := context.Background()
	userStore := NewUserStore(testDB.DB)
	user, err := userStore.CreateUser(ctx, "test@test.com", "test")
	require.NoError(t, err)

	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("test"))
	require.Less(t, now.UnixNano(), user.CreatedAt.UnixNano())

	createdUser, err := userStore.FindUserById(ctx, user.Id)
	require.NoError(t, err)
	require.Equal(t, user.Email, createdUser.Email)
	require.Equal(t, user.Id, createdUser.Id)
	require.Equal(t, user.HashedPasswordBase64, createdUser.HashedPasswordBase64)
	require.Equal(t, user.CreatedAt.UnixNano(), createdUser.CreatedAt.UnixNano())

	createdUser, err = userStore.FindUserById(ctx, user.Id)
	require.NoError(t, err)
	require.Equal(t, user.Email, createdUser.Email)
	require.Equal(t, user.Id, createdUser.Id)
	require.Equal(t, user.HashedPasswordBase64, createdUser.HashedPasswordBase64)
	require.Equal(t, user.CreatedAt.UnixNano(), createdUser.CreatedAt.UnixNano())
}
