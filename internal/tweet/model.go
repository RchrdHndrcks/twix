// Package tweet provides the domain model and business logic for tweet management.
package tweet

import "time"

// Tweet represents a short message published by a user.
type Tweet struct {
	// ID is the unique identifier for the tweet.
	ID string `json:"id"`
	// AuthorID is the ID of the user who published the tweet.
	AuthorID string `json:"author_id"`
	// Content is the text body of the tweet, up to 280 characters.
	Content string `json:"content"`
	// CreatedAt is the timestamp when the tweet was published.
	CreatedAt time.Time `json:"created_at"`
}
