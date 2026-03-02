package tweet_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func TestCreateTweet(t *testing.T) {
	store := memory.NewTweetStore()
	svc := tweet.NewService(store)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tw, err := svc.Create(ctx, "user-1", "Hello world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tw.AuthorID != "user-1" {
			t.Errorf("expected author user-1, got %s", tw.AuthorID)
		}

		if tw.Content != "Hello world" {
			t.Errorf("expected content 'Hello world', got %s", tw.Content)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := svc.Create(ctx, "user-1", "")
		if err != tweet.ErrContentEmpty {
			t.Errorf("expected ErrContentEmpty, got %v", err)
		}
	})

	t.Run("content too long", func(t *testing.T) {
		long := strings.Repeat("a", 281)
		_, err := svc.Create(ctx, "user-1", long)
		if err != tweet.ErrContentTooLong {
			t.Errorf("expected ErrContentTooLong, got %v", err)
		}
	})

	t.Run("exactly 280 characters", func(t *testing.T) {
		content := strings.Repeat("a", 280)
		tw, err := svc.Create(ctx, "user-1", content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len([]rune(tw.Content)) != 280 {
			t.Errorf("expected 280 chars, got %d", len([]rune(tw.Content)))
		}
	})
}

func TestTweet(t *testing.T) {
	store := memory.NewTweetStore()
	svc := tweet.NewService(store)
	ctx := context.Background()

	t.Run("existing tweet", func(t *testing.T) {
		created, _ := svc.Create(ctx, "user-1", "test tweet")

		found, err := svc.Tweet(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if found.Content != "test tweet" {
			t.Errorf("expected content 'test tweet', got %s", found.Content)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Tweet(ctx, "nonexistent")
		if !errors.Is(err, tweet.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
