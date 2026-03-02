// Package follow provides the domain model and business logic for user follow relationships.
package follow

import "time"

// Follow represents a follow relationship between two users.
type Follow struct {
	// FollowerID is the ID of the user who initiates the follow.
	FollowerID string `json:"follower_id"`
	// FolloweeID is the ID of the user being followed.
	FolloweeID string `json:"followee_id"`
	// CreatedAt is the timestamp when the follow relationship was established.
	CreatedAt time.Time `json:"created_at"`
}
