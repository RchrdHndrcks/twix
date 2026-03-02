package memory

import (
	"context"
	"sync"

	"github.com/RchrdHndrcks/twix/internal/follow"
)

// FollowStore is an in-memory implementation of follow.Store.
type FollowStore struct {
	mu sync.RWMutex
	// followers maps userID to the set of user IDs that follow them.
	followers map[string]map[string]bool
	// following maps userID to the set of user IDs they follow.
	following map[string]map[string]bool
}

// NewFollowStore creates a new in-memory follow store.
func NewFollowStore() *FollowStore {
	return &FollowStore{
		followers: make(map[string]map[string]bool),
		following: make(map[string]map[string]bool),
	}
}

// Follow persists a new follow relationship.
func (s *FollowStore) Follow(_ context.Context, f follow.Follow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.followers[f.FolloweeID] == nil {
		s.followers[f.FolloweeID] = make(map[string]bool)
	}
	s.followers[f.FolloweeID][f.FollowerID] = true

	if s.following[f.FollowerID] == nil {
		s.following[f.FollowerID] = make(map[string]bool)
	}
	s.following[f.FollowerID][f.FolloweeID] = true

	return nil
}

// Unfollow removes a follow relationship.
func (s *FollowStore) Unfollow(_ context.Context, followerID, followeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.followers[followeeID], followerID)
	delete(s.following[followerID], followeeID)

	return nil
}

// Followers returns the list of user IDs that follow the given user.
func (s *FollowStore) Followers(_ context.Context, userID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set := s.followers[userID]
	result := make([]string, 0, len(set))
	for id := range set {
		result = append(result, id)
	}

	return result, nil
}

// Following returns the list of user IDs the given user follows.
func (s *FollowStore) Following(_ context.Context, userID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set := s.following[userID]
	result := make([]string, 0, len(set))
	for id := range set {
		result = append(result, id)
	}

	return result, nil
}

// IsFollowing checks whether followerID is following followeeID.
func (s *FollowStore) IsFollowing(_ context.Context, followerID, followeeID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.following[followerID][followeeID], nil
}
