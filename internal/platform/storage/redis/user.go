// Package redis provides Redis-backed implementations of the store interfaces
// for production deployments.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RchrdHndrcks/twix/internal/user"
	"github.com/redis/go-redis/v9"
)

// UserStore is a Redis-backed implementation of user.Store.
type UserStore struct {
	client *redis.Client
}

// NewUserStore creates a new Redis user store.
func NewUserStore(client *redis.Client) *UserStore {
	return &UserStore{client: client}
}

func userKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

// Create persists a new user. Returns user.ErrAlreadyExists if the ID is taken.
func (s *UserStore) Create(ctx context.Context, usr user.User) error {
	data, err := json.Marshal(usr)
	if err != nil {
		return err
	}

	err = s.client.SetArgs(ctx, userKey(usr.ID), data, redis.SetArgs{
		Mode: "NX",
	}).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return user.ErrAlreadyExists
		}
		return err
	}

	return nil
}

// User retrieves a user by ID. Returns user.ErrNotFound if not present.
func (s *UserStore) User(ctx context.Context, id string) (user.User, error) {
	data, err := s.client.Get(ctx, userKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, err
	}

	var usr user.User
	if err := json.Unmarshal(data, &usr); err != nil {
		return user.User{}, err
	}

	return usr, nil
}
