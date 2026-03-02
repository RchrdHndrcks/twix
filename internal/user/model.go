// Package user provides the domain model and business logic for user management.
package user

import "time"

// User represents a registered user of the platform.
type User struct {
	// ID is the unique identifier for the user.
	ID string `json:"id"`
	// Username is the display name chosen by the user.
	Username string `json:"username"`
	// CreatedAt is the timestamp when the user was created.
	CreatedAt time.Time `json:"created_at"`
}
