package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/RchrdHndrcks/twix/internal/tweet"
)

// TimelineStore is an in-memory implementation of timeline.Store.
type TimelineStore struct {
	mu        sync.RWMutex
	timelines map[string][]tweet.Tweet
}

// NewTimelineStore creates a new in-memory timeline store.
func NewTimelineStore() *TimelineStore {
	return &TimelineStore{
		timelines: make(map[string][]tweet.Tweet),
	}
}

// Append adds a tweet to a user's pre-computed timeline.
func (s *TimelineStore) Append(_ context.Context, userID string, tw tweet.Tweet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.timelines[userID] = append(s.timelines[userID], tw)
	return nil
}

// Remove deletes a tweet from a user's timeline.
func (s *TimelineStore) Remove(_ context.Context, userID, tweetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tweets := s.timelines[userID]
	for i, tw := range tweets {
		if tw.ID == tweetID {
			s.timelines[userID] = append(tweets[:i], tweets[i+1:]...)
			return nil
		}
	}

	return nil
}

// Get returns tweets from a user's timeline older than cursor, up to limit entries.
func (s *TimelineStore) Get(_ context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tweets := s.timelines[userID]

	sorted := make([]tweet.Tweet, len(tweets))
	copy(sorted, tweets)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	var result []tweet.Tweet
	for _, tw := range sorted {
		if tw.CreatedAt.Before(cursor) {
			result = append(result, tw)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}
