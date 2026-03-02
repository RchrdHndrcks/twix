package redis_test

import (
	"context"
	"testing"
	"time"

	redisstore "github.com/RchrdHndrcks/twix/internal/platform/storage/redis"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func makeTestTweet(id string, createdAt time.Time) tweet.Tweet {
	return tweet.Tweet{ID: id, AuthorID: "author", Content: "content", CreatedAt: createdAt}
}

func TestRedisTimelineStore_AppendAndGet(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTimelineStore(client)
	ctx := context.Background()

	now := time.Now()
	tw1 := makeTestTweet("t1", now.Add(-3*time.Second))
	tw2 := makeTestTweet("t2", now.Add(-2*time.Second))
	tw3 := makeTestTweet("t3", now.Add(-1*time.Second))

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

func TestRedisTimelineStore_Remove(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTimelineStore(client)
	ctx := context.Background()

	now := time.Now()
	tw1 := makeTestTweet("t1", now.Add(-2*time.Second))
	tw2 := makeTestTweet("t2", now.Add(-1*time.Second))

	_ = store.Append(ctx, "alice", tw1)
	_ = store.Append(ctx, "alice", tw2)

	t.Run("removes existing tweet via index", func(t *testing.T) {
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

func TestRedisTimelineStore_EmptyTimeline(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTimelineStore(client)
	ctx := context.Background()

	tweets, err := store.Get(ctx, "nobody", time.Now(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tweets) != 0 {
		t.Errorf("expected empty timeline, got %d tweets", len(tweets))
	}
}

func TestRedisTimelineStore_IsolationBetweenUsers(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTimelineStore(client)
	ctx := context.Background()

	now := time.Now()
	_ = store.Append(ctx, "alice", makeTestTweet("t1", now.Add(-1*time.Second)))
	_ = store.Append(ctx, "bob", makeTestTweet("t2", now.Add(-1*time.Second)))

	aliceTweets, _ := store.Get(ctx, "alice", now, 10)
	bobTweets, _ := store.Get(ctx, "bob", now, 10)

	if len(aliceTweets) != 1 || aliceTweets[0].ID != "t1" {
		t.Errorf("alice should have [t1], got %v", aliceTweets)
	}
	if len(bobTweets) != 1 || bobTweets[0].ID != "t2" {
		t.Errorf("bob should have [t2], got %v", bobTweets)
	}
}

func TestRedisTimelineStore_TrimToMaxSize(t *testing.T) {
	client := setupRedis(t)
	store := redisstore.NewTimelineStore(client)
	ctx := context.Background()

	// Append more than maxTimelineSize (1000) entries and verify trimming.
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 1005 {
		tw := makeTestTweet("t"+time.Duration(i).String(), baseTime.Add(time.Duration(i)*time.Millisecond))
		_ = store.Append(ctx, "alice", tw)
	}

	// Get all: should be at most 1000 (the sorted set is trimmed on each append).
	cursor := baseTime.Add(2 * time.Hour)
	tweets, err := store.Get(ctx, "alice", cursor, 1100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tweets) > 1000 {
		t.Errorf("expected at most 1000 tweets after trim, got %d", len(tweets))
	}
}
