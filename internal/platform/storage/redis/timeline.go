package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RchrdHndrcks/twix/internal/tweet"
	"github.com/redis/go-redis/v9"
)

const maxTimelineSize = 1000

// TimelineStore is a Redis-backed implementation of timeline.Store
// using sorted sets where the score is the tweet timestamp.
type TimelineStore struct {
	client *redis.Client
}

// NewTimelineStore creates a new Redis timeline store.
func NewTimelineStore(client *redis.Client) *TimelineStore {
	return &TimelineStore{client: client}
}

func timelineKey(userID string) string {
	return fmt.Sprintf("timeline:%s", userID)
}

func timelineIndexKey(userID string) string {
	return fmt.Sprintf("timeline_idx:%s", userID)
}

func timeScore(t time.Time) float64 {
	return float64(t.UnixNano())
}

// Append adds a tweet to a user's pre-computed timeline and trims old entries.
// It also maintains a secondary index (hash) mapping tweetID → JSON for O(1) removal.
func (s *TimelineStore) Append(ctx context.Context, userID string, tw tweet.Tweet) error {
	data, err := json.Marshal(tw)
	if err != nil {
		return err
	}

	member := string(data)
	pipe := s.client.Pipeline()

	pipe.ZAdd(ctx, timelineKey(userID), redis.Z{
		Score:  timeScore(tw.CreatedAt),
		Member: member,
	})
	pipe.HSet(ctx, timelineIndexKey(userID), tw.ID, member)

	pipe.ZRemRangeByRank(ctx, timelineKey(userID), 0, -maxTimelineSize-1)

	_, err = pipe.Exec(ctx)
	return err
}

// Remove deletes a tweet from a user's timeline using the secondary index for O(1) lookup.
func (s *TimelineStore) Remove(ctx context.Context, userID, tweetID string) error {
	member, err := s.client.HGet(ctx, timelineIndexKey(userID), tweetID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	pipe := s.client.Pipeline()
	pipe.ZRem(ctx, timelineKey(userID), member)
	pipe.HDel(ctx, timelineIndexKey(userID), tweetID)
	_, err = pipe.Exec(ctx)
	return err
}

// Get returns tweets from a user's timeline older than cursor, up to limit entries.
func (s *TimelineStore) Get(ctx context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error) {
	maxScore := fmt.Sprintf("(%f", timeScore(cursor))

	// go-redis swaps Start/Stop when Rev+ByScore are set, so we pass them
	// in logical order (min, max) and let go-redis reorder for the ZRANGE command.
	results, err := s.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     timelineKey(userID),
		Start:   "-inf",
		Stop:    maxScore,
		ByScore: true,
		Rev:     true,
		Count:   int64(limit),
	}).Result()
	if err != nil {
		return nil, err
	}

	tweets := make([]tweet.Tweet, 0, len(results))
	for _, data := range results {
		var tw tweet.Tweet
		if err := json.Unmarshal([]byte(data), &tw); err != nil {
			return nil, err
		}
		tweets = append(tweets, tw)
	}

	return tweets, nil
}
