package core

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user account.
type UserStatus string

const (
	StatusActive        UserStatus = "active"
	StatusPending       UserStatus = "pending" // Pending email verification
	StatusSuspended     UserStatus = "suspended"
	StatusDeleted       UserStatus = "deleted"
	StatusPendingDelete UserStatus = "pending_delete" // Soft delete
)

// User is the canonical user representation the SDK works with.
// The UserStorer implementation is responsible for mapping this
// to/from the application's actual database schema.
type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string `json:"-"` // Exclude from default JSON responses
	FullName     string
	Role         string // e.g., "user", "admin"
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// For single-device login enforcement
	ActiveToken string `json:"-"` // Store the currently active access token
	// TODO: later we will Consider a separate struct/table for more complex session management
}
