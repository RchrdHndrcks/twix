package memory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	"github.com/RchrdHndrcks/twix/internal/user"
)

func TestUserStore_Create(t *testing.T) {
	store := memory.NewUserStore()
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

func TestUserStore_User(t *testing.T) {
	store := memory.NewUserStore()
	ctx := context.Background()

	usr := user.User{ID: "u1", Username: "alice", CreatedAt: time.Now()}
	_ = store.Create(ctx, usr)

	t.Run("existing user", func(t *testing.T) {
		got, err := store.User(ctx, "u1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Username != "alice" {
			t.Errorf("expected username 'alice', got %q", got.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := store.User(ctx, "nonexistent")
		if err != user.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUserStore_ConcurrentAccess(t *testing.T) {
	store := memory.NewUserStore()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			usr := user.User{
				ID:        "u" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
				Username:  "user",
				CreatedAt: time.Now(),
			}
			_ = store.Create(ctx, usr)
			_, _ = store.User(ctx, usr.ID)
		}()
	}
	wg.Wait()
}
