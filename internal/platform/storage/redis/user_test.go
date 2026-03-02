package redis_test

import (
	"context"
	"testing"
	"time"

	redisstore "github.com/RchrdHndrcks/twix/internal/platform/storage/redis"
	"github.com/RchrdHndrcks/twix/internal/user"
)

func TestRedisUserStore_Create(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewUserStore(client)
	ctx := context.Background()

	usr := user.User{ID: "u1", Username: "alice", CreatedAt: time.Now()}

	t.Run("success", func(t *testing.T) {
		if err := store.Create(ctx, usr); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("duplicate ID returns ErrAlreadyExists", func(t *testing.T) {
		err := store.Create(ctx, usr)
		if err != user.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestRedisUserStore_User(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewUserStore(client)
	ctx := context.Background()

	usr := user.User{ID: "u1", Username: "alice", CreatedAt: time.Now().Truncate(time.Millisecond)}
	_ = store.Create(ctx, usr)

	t.Run("existing user", func(t *testing.T) {
		got, err := store.User(ctx, "u1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Username != "alice" {
			t.Errorf("expected username 'alice', got %q", got.Username)
		}
		if got.ID != "u1" {
			t.Errorf("expected ID 'u1', got %q", got.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := store.User(ctx, "nonexistent")
		if err != user.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
