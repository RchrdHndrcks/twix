package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/internal/follow"
	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	"github.com/RchrdHndrcks/twix/internal/timeline"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func setup() (*timeline.Service, *tweet.Service, *follow.Service) {
	tweetStore := memory.NewTweetStore()
	followStore := memory.NewFollowStore()
	timelineStore := memory.NewTimelineStore()

	tweetSvc := tweet.NewService(tweetStore)
	followSvc := follow.NewService(followStore)
	timelineSvc := timeline.NewService(timelineStore, tweetSvc, followSvc)

	return timelineSvc, tweetSvc, followSvc
}

func TestPublishTweet(t *testing.T) {
	timelineSvc, _, followSvc := setup()
	ctx := context.Background()

	_ = followSvc.Follow(ctx, "alice", "bob")

	t.Run("tweet appears in follower timeline", func(t *testing.T) {
		tw, err := timelineSvc.PublishTweet(ctx, "bob", "Hello from Bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tw.Content != "Hello from Bob" {
			t.Errorf("expected content 'Hello from Bob', got %s", tw.Content)
		}

		tweets, err := timelineSvc.Timeline(ctx, "alice", time.Time{}, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(tweets) != 1 {
			t.Fatalf("expected 1 tweet in timeline, got %d", len(tweets))
		}

		if tweets[0].Content != "Hello from Bob" {
			t.Errorf("expected 'Hello from Bob', got %s", tweets[0].Content)
		}
	})

	t.Run("tweet does not appear for non-followers", func(t *testing.T) {
		tweets, err := timelineSvc.Timeline(ctx, "charlie", time.Time{}, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(tweets) != 0 {
			t.Errorf("expected 0 tweets, got %d", len(tweets))
		}
	})
}

func TestTimelinePagination(t *testing.T) {
	timelineSvc, _, followSvc := setup()
	ctx := context.Background()

	_ = followSvc.Follow(ctx, "alice", "bob")

	for i := 0; i < 5; i++ {
		_, _ = timelineSvc.PublishTweet(ctx, "bob", "Tweet number "+string(rune('1'+i)))
		time.Sleep(time.Millisecond)
	}

	t.Run("limit results", func(t *testing.T) {
		tweets, err := timelineSvc.Timeline(ctx, "alice", time.Time{}, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(tweets) != 3 {
			t.Fatalf("expected 3 tweets, got %d", len(tweets))
		}
	})

	t.Run("cursor pagination", func(t *testing.T) {
		page1, err := timelineSvc.Timeline(ctx, "alice", time.Time{}, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(page1) != 3 {
			t.Fatalf("expected 3 tweets on page 1, got %d", len(page1))
		}

		cursor := page1[len(page1)-1].CreatedAt
		page2, err := timelineSvc.Timeline(ctx, "alice", cursor, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(page2) != 2 {
			t.Fatalf("expected 2 tweets on page 2, got %d", len(page2))
		}

		for _, tw1 := range page1 {
			for _, tw2 := range page2 {
				if tw1.ID == tw2.ID {
					t.Errorf("duplicate tweet found across pages: %s", tw1.ID)
				}
			}
		}
	})
}

func TestPublishTweetValidation(t *testing.T) {
	timelineSvc, _, _ := setup()
	ctx := context.Background()

	t.Run("empty content", func(t *testing.T) {
		_, err := timelineSvc.PublishTweet(ctx, "bob", "")
		if err == nil {
			t.Error("expected error for empty content")
		}
	})

	t.Run("content too long", func(t *testing.T) {
		long := make([]byte, 281)
		for i := range long {
			long[i] = 'a'
		}
		_, err := timelineSvc.PublishTweet(ctx, "bob", string(long))
		if err == nil {
			t.Error("expected error for content too long")
		}
	})
}
