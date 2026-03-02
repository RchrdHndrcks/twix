package memory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func TestTweetStore_Create(t *testing.T) {
	store := memory.NewTweetStore()
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
}

func TestTweetStore_CreateOverwrites(t *testing.T) {
	store := memory.NewTweetStore()
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

func TestTweetStore_NotFound(t *testing.T) {
	store := memory.NewTweetStore()
	ctx := context.Background()

	_, err := store.Tweet(ctx, "nonexistent")
	if err != tweet.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTweetStore_ConcurrentAccess(t *testing.T) {
	store := memory.NewTweetStore()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tw := tweet.Tweet{
				ID:        "t" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
				AuthorID:  "u1",
				Content:   "tweet",
				CreatedAt: time.Now(),
			}
			_ = store.Create(ctx, tw)
			_, _ = store.Tweet(ctx, tw.ID)
		}()
	}
	wg.Wait()
}
