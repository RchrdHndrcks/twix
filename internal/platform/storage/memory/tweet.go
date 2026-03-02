package memory

import (
	"context"
	"sync"

	"github.com/RchrdHndrcks/twix/internal/tweet"
)

// TweetStore is an in-memory implementation of tweet.Store.
type TweetStore struct {
	mu     sync.RWMutex
	tweets map[string]tweet.Tweet
}

// NewTweetStore creates a new in-memory tweet store.
func NewTweetStore() *TweetStore {
	return &TweetStore{
		tweets: make(map[string]tweet.Tweet),
	}
}

// Create persists a new tweet.
func (s *TweetStore) Create(_ context.Context, tw tweet.Tweet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tweets[tw.ID] = tw
	return nil
}

// Tweet retrieves a tweet by ID. Returns tweet.ErrNotFound if not present.
func (s *TweetStore) Tweet(_ context.Context, id string) (tweet.Tweet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tw, exists := s.tweets[id]
	if !exists {
		return tweet.Tweet{}, tweet.ErrNotFound
	}

	return tw, nil
}
