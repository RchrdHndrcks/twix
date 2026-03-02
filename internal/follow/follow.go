package follow

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Store defines the persistence contract for follow relationships.
type Store interface {
	Follow(ctx context.Context, f Follow) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	Followers(ctx context.Context, userID string) ([]string, error)
	Following(ctx context.Context, userID string) ([]string, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
}

var (
	// ErrSelfFollow is returned when a user tries to follow themselves.
	ErrSelfFollow = errors.New("a user cannot follow themselves")

	// ErrAlreadyFollowing is returned when the follow relationship already exists.
	ErrAlreadyFollowing = errors.New("already following this user")

	// ErrNotFollowing is returned when trying to unfollow a user that is not being followed.
	ErrNotFollowing = errors.New("not following this user")
)

// Service handles follow relationship business logic.
type Service struct {
	store Store
}

// NewService creates a new follow service with the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Follow establishes a follow relationship between two users.
func (s *Service) Follow(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrSelfFollow
	}

	following, err := s.store.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("checking follow status: %w", err)
	}

	if following {
		return ErrAlreadyFollowing
	}

	f := Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		CreatedAt:  time.Now(),
	}

	if err := s.store.Follow(ctx, f); err != nil {
		return fmt.Errorf("creating follow: %w", err)
	}

	return nil
}

// Unfollow removes a follow relationship between two users.
func (s *Service) Unfollow(ctx context.Context, followerID, followeeID string) error {
	following, err := s.store.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("checking follow status: %w", err)
	}

	if !following {
		return ErrNotFollowing
	}

	if err := s.store.Unfollow(ctx, followerID, followeeID); err != nil {
		return fmt.Errorf("removing follow: %w", err)
	}

	return nil
}

// Followers returns the list of user IDs that follow the given user.
func (s *Service) Followers(ctx context.Context, userID string) ([]string, error) {
	followers, err := s.store.Followers(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing followers: %w", err)
	}

	return followers, nil
}
