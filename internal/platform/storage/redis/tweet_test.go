package redis_test

import (
	"context"
	"testing"
	"time"

	redisstore "github.com/RchrdHndrcks/twix/internal/platform/storage/redis"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func TestRedisTweetStore_Create(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTweetStore(client)
	ctx := context.Background()

	tw := tweet.Tweet{ID: "t1", AuthorID: "u1", Content: "hello", CreatedAt: time.Now()}

	if err := store.Create(ctx, tw); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.Tweet(ctx, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != "hello" {
		t.Errorf("expected content 'hello', got %q", got.Content)
	}
	if got.AuthorID != "u1" {
		t.Errorf("expected authorID 'u1', got %q", got.AuthorID)
	}
}

func TestRedisTweetStore_CreateOverwrites(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTweetStore(client)
	ctx := context.Background()

	tw1 := tweet.Tweet{ID: "t1", AuthorID: "u1", Content: "first", CreatedAt: time.Now()}
	tw2 := tweet.Tweet{ID: "t1", AuthorID: "u1", Content: "second", CreatedAt: time.Now()}

	_ = store.Create(ctx, tw1)
	_ = store.Create(ctx, tw2)

	got, _ := store.Tweet(ctx, "t1")
	if got.Content != "second" {
		t.Errorf("expected overwritten content 'second', got %q", got.Content)
	}
}

func TestRedisTweetStore_NotFound(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTweetStore(client)
	ctx := context.Background()

	_, err := store.Tweet(ctx, "nonexistent")
	if err != tweet.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
