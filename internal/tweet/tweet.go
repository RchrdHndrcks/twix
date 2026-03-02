package tweet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store defines the persistence contract for tweets.
type Store interface {
	Create(ctx context.Context, tw Tweet) error
	Tweet(ctx context.Context, id string) (Tweet, error)
}

// MaxContentLength is the maximum number of characters allowed in a tweet.
const MaxContentLength = 280

var (
	// ErrNotFound is returned when a tweet is not found.
	ErrNotFound = errors.New("tweet not found")

	// ErrContentEmpty is returned when trying to create a tweet with empty content.
	ErrContentEmpty = errors.New("tweet content cannot be empty")

	// ErrContentTooLong is returned when the tweet content exceeds MaxContentLength.
	ErrContentTooLong = errors.New("tweet content exceeds 280 characters")
)

// Service handles tweet business logic.
type Service struct {
	store Store
}

// NewService creates a new tweet service with the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create validates the content and persists a new tweet.
func (s *Service) Create(ctx context.Context, authorID, content string) (Tweet, error) {
	if content == "" {
		return Tweet{}, ErrContentEmpty
	}

	if len([]rune(content)) > MaxContentLength {
		return Tweet{}, ErrContentTooLong
	}

	tw := Tweet{
		ID:        uuid.NewString(),
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.store.Create(ctx, tw); err != nil {
		return Tweet{}, fmt.Errorf("creating tweet: %w", err)
	}

	return tw, nil
}

// Tweet retrieves a tweet by its unique identifier.
func (s *Service) Tweet(ctx context.Context, id string) (Tweet, error) {
	tw, err := s.store.Tweet(ctx, id)
	if err != nil {
		return Tweet{}, fmt.Errorf("querying tweet: %w", err)
	}

	return tw, nil
}
