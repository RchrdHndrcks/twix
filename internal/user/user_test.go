package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	"github.com/RchrdHndrcks/twix/internal/user"
)

func TestCreateUser(t *testing.T) {
	store := memory.NewUserStore()
	svc := user.NewService(store)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		usr, err := svc.Create(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if usr.Username != "alice" {
			t.Errorf("expected username alice, got %s", usr.Username)
		}

		if usr.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("empty username", func(t *testing.T) {
		_, err := svc.Create(ctx, "")
		if err != user.ErrUsernameEmpty {
			t.Errorf("expected ErrUsernameEmpty, got %v", err)
		}
	})
}

func TestUser(t *testing.T) {
	store := memory.NewUserStore()
	svc := user.NewService(store)
	ctx := context.Background()

	t.Run("existing user", func(t *testing.T) {
		created, _ := svc.Create(ctx, "bob")

		found, err := svc.User(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if found.ID != created.ID {
			t.Errorf("expected ID %s, got %s", created.ID, found.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.User(ctx, "nonexistent")
		if !errors.Is(err, user.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
