package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/internal/follow"
	redisstore "github.com/RchrdHndrcks/twix/internal/platform/storage/redis"
)

func TestRedisFollowStore_Follow(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewFollowStore(client)
	ctx := context.Background()

	f := follow.Follow{FollowerID: "alice", FolloweeID: "bob", CreatedAt: time.Now()}

	if err := store.Follow(ctx, f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("appears in followers of followee", func(t *testing.T) {
		followers, err := store.Followers(ctx, "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(followers) != 1 || followers[0] != "alice" {
			t.Errorf("expected [alice], got %v", followers)
		}
	})

	t.Run("appears in following of follower", func(t *testing.T) {
		following, err := store.Following(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(following) != 1 || following[0] != "bob" {
			t.Errorf("expected [bob], got %v", following)
		}
	})

	t.Run("IsFollowing returns true", func(t *testing.T) {
		ok, err := store.IsFollowing(ctx, "alice", "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected IsFollowing to be true")
		}
	})

	t.Run("reverse direction is false", func(t *testing.T) {
		ok, err := store.IsFollowing(ctx, "bob", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected IsFollowing (reverse) to be false")
		}
	})
}

func TestRedisFollowStore_Unfollow(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewFollowStore(client)
	ctx := context.Background()

	f := follow.Follow{FollowerID: "alice", FolloweeID: "bob", CreatedAt: time.Now()}
	_ = store.Follow(ctx, f)

	if err := store.Unfollow(ctx, "alice", "bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("no longer in followers", func(t *testing.T) {
		followers, _ := store.Followers(ctx, "bob")
		if len(followers) != 0 {
			t.Errorf("expected 0 followers, got %d", len(followers))
		}
	})

	t.Run("no longer in following", func(t *testing.T) {
		following, _ := store.Following(ctx, "alice")
		if len(following) != 0 {
			t.Errorf("expected 0 following, got %d", len(following))
		}
	})

	t.Run("IsFollowing returns false", func(t *testing.T) {
		ok, _ := store.IsFollowing(ctx, "alice", "bob")
		if ok {
			t.Error("expected IsFollowing to be false after unfollow")
		}
	})
}

func TestRedisFollowStore_EmptyLists(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewFollowStore(client)
	ctx := context.Background()

	followers, err := store.Followers(ctx, "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(followers) != 0 {
		t.Errorf("expected empty followers, got %v", followers)
	}

	following, err := store.Following(ctx, "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(following) != 0 {
		t.Errorf("expected empty following, got %v", following)
	}
}

func TestRedisFollowStore_MultipleFollowers(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewFollowStore(client)
	ctx := context.Background()

	_ = store.Follow(ctx, follow.Follow{FollowerID: "alice", FolloweeID: "charlie", CreatedAt: time.Now()})
	_ = store.Follow(ctx, follow.Follow{FollowerID: "bob", FolloweeID: "charlie", CreatedAt: time.Now()})

	followers, _ := store.Followers(ctx, "charlie")
	if len(followers) != 2 {
		t.Fatalf("expected 2 followers, got %d", len(followers))
	}

	got := map[string]bool{followers[0]: true, followers[1]: true}
	if !got["alice"] || !got["bob"] {
		t.Errorf("expected alice and bob, got %v", followers)
	}
}
