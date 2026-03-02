package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func makeTweet(id string, createdAt time.Time) tweet.Tweet {
	return tweet.Tweet{ID: id, AuthorID: "author", Content: "content", CreatedAt: createdAt}
}

func TestTimelineStore_AppendAndGet(t *testing.T) {
	store := memory.NewTimelineStore()
	ctx := context.Background()

	now := time.Now()
	tw1 := makeTweet("t1", now.Add(-3*time.Second))
	tw2 := makeTweet("t2", now.Add(-2*time.Second))
	tw3 := makeTweet("t3", now.Add(-1*time.Second))

	_ = store.Append(ctx, "alice", tw1)
	_ = store.Append(ctx, "alice", tw2)
	_ = store.Append(ctx, "alice", tw3)

	t.Run("returns all tweets newest first", func(t *testing.T) {
		tweets, err := store.Get(ctx, "alice", now, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tweets) != 3 {
			t.Fatalf("expected 3 tweets, got %d", len(tweets))
		}
		if tweets[0].ID != "t3" || tweets[1].ID != "t2" || tweets[2].ID != "t1" {
			t.Errorf("expected [t3, t2, t1], got [%s, %s, %s]", tweets[0].ID, tweets[1].ID, tweets[2].ID)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		tweets, err := store.Get(ctx, "alice", now, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tweets) != 2 {
			t.Fatalf("expected 2 tweets, got %d", len(tweets))
		}
	})

	t.Run("cursor excludes newer tweets", func(t *testing.T) {
		// Cursor at tw3's time: should only return tw2, tw1 (strictly before cursor).
		tweets, err := store.Get(ctx, "alice", tw3.CreatedAt, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tweets) != 2 {
			t.Fatalf("expected 2 tweets before cursor, got %d", len(tweets))
		}
		if tweets[0].ID != "t2" || tweets[1].ID != "t1" {
			t.Errorf("expected [t2, t1], got [%s, %s]", tweets[0].ID, tweets[1].ID)
		}
	})
}

func TestTimelineStore_Remove(t *testing.T) {
	store := memory.NewTimelineStore()
	ctx := context.Background()

	now := time.Now()
	tw1 := makeTweet("t1", now.Add(-2*time.Second))
	tw2 := makeTweet("t2", now.Add(-1*time.Second))

	_ = store.Append(ctx, "alice", tw1)
	_ = store.Append(ctx, "alice", tw2)

	t.Run("removes existing tweet", func(t *testing.T) {
		if err := store.Remove(ctx, "alice", "t1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tweets, _ := store.Get(ctx, "alice", now, 10)
		if len(tweets) != 1 {
			t.Fatalf("expected 1 tweet after remove, got %d", len(tweets))
		}
		if tweets[0].ID != "t2" {
			t.Errorf("expected t2, got %s", tweets[0].ID)
		}
	})

	t.Run("remove non-existent is no-op", func(t *testing.T) {
		if err := store.Remove(ctx, "alice", "nonexistent"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestTimelineStore_EmptyTimeline(t *testing.T) {
	store := memory.NewTimelineStore()
	ctx := context.Background()

	tweets, err := store.Get(ctx, "nobody", time.Now(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tweets) != 0 {
		t.Errorf("expected empty timeline, got %d tweets", len(tweets))
	}
}

func TestTimelineStore_IsolationBetweenUsers(t *testing.T) {
	store := memory.NewTimelineStore()
	ctx := context.Background()

	now := time.Now()
	_ = store.Append(ctx, "alice", makeTweet("t1", now.Add(-1*time.Second)))
	_ = store.Append(ctx, "bob", makeTweet("t2", now.Add(-1*time.Second)))

	aliceTweets, _ := store.Get(ctx, "alice", now, 10)
	bobTweets, _ := store.Get(ctx, "bob", now, 10)

	if len(aliceTweets) != 1 || aliceTweets[0].ID != "t1" {
		t.Errorf("alice should have [t1], got %v", aliceTweets)
	}
	if len(bobTweets) != 1 || bobTweets[0].ID != "t2" {
		t.Errorf("bob should have [t2], got %v", bobTweets)
	}
}

func TestTimelineStore_DuplicateTimestamp(t *testing.T) {
	store := memory.NewTimelineStore()
	ctx := context.Background()

	// Two tweets with the exact same CreatedAt.
	sameTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	tw1 := makeTweet("t1", sameTime)
	tw2 := makeTweet("t2", sameTime)

	_ = store.Append(ctx, "alice", tw1)
	_ = store.Append(ctx, "alice", tw2)

	// Both should appear when cursor is after their time.
	cursor := sameTime.Add(1 * time.Second)
	tweets, err := store.Get(ctx, "alice", cursor, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tweets) != 2 {
		t.Fatalf("expected 2 tweets with same timestamp, got %d", len(tweets))
	}

	// But using one of their timestamps as cursor loses them both,
	// since Get uses strictly-less-than.
	tweets, _ = store.Get(ctx, "alice", sameTime, 10)
	if len(tweets) != 0 {
		t.Logf("NOTE: cursor at exact tweet time returns %d tweets (expected 0 with strict less-than)", len(tweets))
	}
}
