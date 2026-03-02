package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store defines the persistence contract for users.
type Store interface {
	Create(ctx context.Context, usr User) error
	User(ctx context.Context, id string) (User, error)
}

var (
	// ErrNotFound is returned when a user is not found.
	ErrNotFound = errors.New("user not found")

	// ErrUsernameEmpty is returned when trying to create a user with an empty username.
	ErrUsernameEmpty = errors.New("username cannot be empty")

	// ErrAlreadyExists is returned when trying to create a user that already exists.
	ErrAlreadyExists = errors.New("user already exists")
)

// Service handles user business logic.
type Service struct {
	store Store
}

// NewService creates a new user service with the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create validates the input and persists a new user.
func (s *Service) Create(ctx context.Context, username string) (User, error) {
	if username == "" {
		return User{}, ErrUsernameEmpty
	}

	usr := User{
		ID:        uuid.NewString(),
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := s.store.Create(ctx, usr); err != nil {
		return User{}, fmt.Errorf("creating user: %w", err)
	}

	return usr, nil
}

// User retrieves a user by their unique identifier.
func (s *Service) User(ctx context.Context, id string) (User, error) {
	usr, err := s.store.User(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("querying user: %w", err)
	}

	return usr, nil
}
