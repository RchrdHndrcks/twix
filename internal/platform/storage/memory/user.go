// Package memory provides in-memory implementations of the store interfaces
// for local development and testing.
package memory

import (
	"context"
	"sync"

	"github.com/RchrdHndrcks/twix/internal/user"
)

// UserStore is an in-memory implementation of user.Store.
type UserStore struct {
	mu    sync.RWMutex
	users map[string]user.User
}

// NewUserStore creates a new in-memory user store.
func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]user.User),
	}
}

// Create persists a new user. Returns user.ErrAlreadyExists if the ID is taken.
func (s *UserStore) Create(_ context.Context, usr user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[usr.ID]; exists {
		return user.ErrAlreadyExists
	}

	s.users[usr.ID] = usr
	return nil
}

// User retrieves a user by ID. Returns user.ErrNotFound if not present.
func (s *UserStore) User(_ context.Context, id string) (user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usr, exists := s.users[id]
	if !exists {
		return user.User{}, user.ErrNotFound
	}

	return usr, nil
}
