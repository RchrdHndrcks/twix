package redis

import (
	"context"
	"fmt"

	"github.com/RchrdHndrcks/twix/internal/follow"
	"github.com/redis/go-redis/v9"
)

// FollowStore is a Redis-backed implementation of follow.Store.
type FollowStore struct {
	client *redis.Client
}

// NewFollowStore creates a new Redis follow store.
func NewFollowStore(client *redis.Client) *FollowStore {
	return &FollowStore{client: client}
}

func followersKey(userID string) string {
	return fmt.Sprintf("followers:%s", userID)
}

func followingKey(userID string) string {
	return fmt.Sprintf("following:%s", userID)
}

// Follow persists a new follow relationship using Redis sets.
func (s *FollowStore) Follow(ctx context.Context, f follow.Follow) error {
	pipe := s.client.Pipeline()
	pipe.SAdd(ctx, followersKey(f.FolloweeID), f.FollowerID)
	pipe.SAdd(ctx, followingKey(f.FollowerID), f.FolloweeID)
	_, err := pipe.Exec(ctx)
	return err
}

// Unfollow removes a follow relationship.
func (s *FollowStore) Unfollow(ctx context.Context, followerID, followeeID string) error {
	pipe := s.client.Pipeline()
	pipe.SRem(ctx, followersKey(followeeID), followerID)
	pipe.SRem(ctx, followingKey(followerID), followeeID)
	_, err := pipe.Exec(ctx)
	return err
}

// Followers returns the list of user IDs that follow the given user.
func (s *FollowStore) Followers(ctx context.Context, userID string) ([]string, error) {
	return s.client.SMembers(ctx, followersKey(userID)).Result()
}

// Following returns the list of user IDs the given user follows.
func (s *FollowStore) Following(ctx context.Context, userID string) ([]string, error) {
	return s.client.SMembers(ctx, followingKey(userID)).Result()
}

// IsFollowing checks whether followerID is following followeeID.
func (s *FollowStore) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return s.client.SIsMember(ctx, followingKey(followerID), followeeID).Result()
}
