package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CreateUserParams for UserStorer.CreateUser
type CreateUserParams struct {
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	Role         string
	Status       UserStatus
}

// UpdateUserParams for UserStorer.UpdateUser
// Use pointers or a map for partial updates if needed, or dedicated methods.
type UpdateUserParams struct {
	FullName     *string
	PasswordHash *string
	Role         *string
	Status       *UserStatus
	ActiveToken  *string // To set or clear the active token
}

// UserStorer defines methods an application must implement for user persistence.
type UserStorer interface {
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, params UpdateUserParams) (User, error)
	// TODO: DeleteUser(ctx context.Context, userID uuid.UUID) error // For hard delete later

	// For Email Verification
	StoreVerificationData(ctx context.Context, userID uuid.UUID, email string, token string, expiresAt time.Time) error
	GetVerificationData(ctx context.Context, token string) (userID uuid.UUID, email string, err error)
	DeleteVerificationData(ctx context.Context, token string) error
	DeleteVerificationDataByUserID(ctx context.Context, userID uuid.UUID) error // If user changes email

	// For Password Reset
	StorePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, token string) (userID uuid.UUID, err error)
	DeletePasswordResetToken(ctx context.Context, token string) error
}

// EmailSender defines methods an application must implement for sending emails.
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, toEmail, username, verificationLink string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, username, resetLink string) error
	// TODO: Add other email types as needed (e.g., SendWelcomeEmail)
}
