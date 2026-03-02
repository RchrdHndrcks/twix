package follow_test

import (
	"context"
	"testing"

	"github.com/RchrdHndrcks/twix/internal/follow"
	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
)

func TestFollow(t *testing.T) {
	store := memory.NewFollowStore()
	svc := follow.NewService(store)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		err := svc.Follow(ctx, "alice", "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("self follow", func(t *testing.T) {
		err := svc.Follow(ctx, "charlie", "charlie")
		if err != follow.ErrSelfFollow {
			t.Errorf("expected ErrSelfFollow, got %v", err)
		}
	})

	t.Run("already following", func(t *testing.T) {
		err := svc.Follow(ctx, "alice", "bob")
		if err != follow.ErrAlreadyFollowing {
			t.Errorf("expected ErrAlreadyFollowing, got %v", err)
		}
	})
}

func TestUnfollow(t *testing.T) {
	store := memory.NewFollowStore()
	svc := follow.NewService(store)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		_ = svc.Follow(ctx, "alice", "bob")

		err := svc.Unfollow(ctx, "alice", "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not following", func(t *testing.T) {
		err := svc.Unfollow(ctx, "alice", "charlie")
		if err != follow.ErrNotFollowing {
			t.Errorf("expected ErrNotFollowing, got %v", err)
		}
	})
}

func TestFollowers(t *testing.T) {
	store := memory.NewFollowStore()
	svc := follow.NewService(store)
	ctx := context.Background()

	_ = svc.Follow(ctx, "alice", "bob")
	_ = svc.Follow(ctx, "charlie", "bob")

	followers, err := svc.Followers(ctx, "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(followers) != 2 {
		t.Errorf("expected 2 followers, got %d", len(followers))
	}
}
