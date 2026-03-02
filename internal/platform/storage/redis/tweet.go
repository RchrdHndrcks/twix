package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RchrdHndrcks/twix/internal/tweet"
	"github.com/redis/go-redis/v9"
)

// TweetStore is a Redis-backed implementation of tweet.Store.
type TweetStore struct {
	client *redis.Client
}

// NewTweetStore creates a new Redis tweet store.
func NewTweetStore(client *redis.Client) *TweetStore {
	return &TweetStore{client: client}
}

func tweetKey(id string) string {
	return fmt.Sprintf("tweet:%s", id)
}

// Create persists a new tweet.
func (s *TweetStore) Create(ctx context.Context, tw tweet.Tweet) error {
	data, err := json.Marshal(tw)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, tweetKey(tw.ID), data, 0).Err()
}

// Tweet retrieves a tweet by ID. Returns tweet.ErrNotFound if not present.
func (s *TweetStore) Tweet(ctx context.Context, id string) (tweet.Tweet, error) {
	data, err := s.client.Get(ctx, tweetKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return tweet.Tweet{}, tweet.ErrNotFound
		}
		return tweet.Tweet{}, err
	}

	var tw tweet.Tweet
	if err := json.Unmarshal(data, &tw); err != nil {
		return tweet.Tweet{}, err
	}

	return tw, nil
}
