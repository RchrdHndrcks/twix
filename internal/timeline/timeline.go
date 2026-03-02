package timeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RchrdHndrcks/twix/internal/tweet"
)

// Store defines the persistence contract for pre-computed user timelines.
type Store interface {
	Append(ctx context.Context, userID string, tw tweet.Tweet) error
	Remove(ctx context.Context, userID, tweetID string) error
	Get(ctx context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error)
}

// DefaultLimit is the number of tweets returned when no limit is specified.
const DefaultLimit = 20

// TweetCreator defines the contract for creating tweets.
type TweetCreator interface {
	Create(ctx context.Context, authorID, content string) (tweet.Tweet, error)
}

// FollowerLister defines the contract for listing a user's followers.
type FollowerLister interface {
	Followers(ctx context.Context, userID string) ([]string, error)
}

// Service orchestrates timeline operations including fan-out on write.
type Service struct {
	timelineStore Store
	tweetService  TweetCreator
	followStore   FollowerLister
}

// NewService creates a new timeline service with the given dependencies.
func NewService(ts Store, tw TweetCreator, fs FollowerLister) *Service {
	return &Service{
		timelineStore: ts,
		tweetService:  tw,
		followStore:   fs,
	}
}

// PublishTweet creates a tweet and fans it out to all followers' timelines.
func (s *Service) PublishTweet(ctx context.Context, authorID, content string) (tweet.Tweet, error) {
	tw, err := s.tweetService.Create(ctx, authorID, content)
	if err != nil {
		return tweet.Tweet{}, fmt.Errorf("creating tweet: %w", err)
	}

	followers, err := s.followStore.Followers(ctx, authorID)
	if err != nil {
		return tweet.Tweet{}, fmt.Errorf("listing followers: %w", err)
	}

	for _, followerID := range followers {
		if err := s.timelineStore.Append(ctx, followerID, tw); err != nil {
			log.Printf("fan-out: failed to append tweet %s to timeline of user %s: %v", tw.ID, followerID, err)
		}
	}

	return tw, nil
}

// Timeline returns the pre-computed timeline for a user with cursor-based pagination.
func (s *Service) Timeline(ctx context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}

	if cursor.IsZero() {
		cursor = time.Now()
	}

	tweets, err := s.timelineStore.Get(ctx, userID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("querying timeline: %w", err)
	}

	return tweets, nil
}
