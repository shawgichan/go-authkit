package ginhandler

import (
	"time"

	"github.com/google/uuid"
	"github.com/shawgichan/go-authkit/core" // Adjust import path
)

// === Request Structs (for SDK-provided handlers) ===

// RegisterRequest defines the expected body for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"` // Assuming min length from config
	FullName string `json:"full_name" binding:"required"`
	// Role is typically set by the system (e.g., default role from config)
	// or handled by application-specific logic if allowed from request.
	// For a generic SDK, role might not be part of the direct request here.
}

// LoginRequest defines the expected body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// VerifyEmailRequest defines query parameters for email verification.
// The SDK handler would get the token from the query.
type VerifyEmailRequest struct {
	Token string `form:"token" binding:"required"` // Or "secret_code" depending on link
}

// ForgotPasswordRequest defines the expected body for initiating password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest defines the expected body for resetting a password.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePasswordRequest for authenticated users changing their own password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangeNameRequest for authenticated users changing their full name.
type ChangeNameRequest struct {
	FullName string `json:"full_name" binding:"required"`
}

// === Response Structs (for SDK-provided handlers) ===

// UserResponse is a generic representation of a user for API responses.
// It omits sensitive information like PasswordHash.
type UserResponse struct {
	ID        uuid.UUID       `json:"id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	FullName  string          `json:"full_name"`
	Role      string          `json:"role"`
	Status    core.UserStatus `json:"status"` // Use core.UserStatus type
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	// Add other non-sensitive fields you want to expose
}

// NewUserResponse maps a core.User to a UserResponse.
func NewSDKUserResponse(user core.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// TokenResponse is returned on successful login or token refresh.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	// RefreshToken string    `json:"refresh_token,omitempty"` // If you implement refresh tokens
	User      UserResponse `json:"user"`
	ExpiresAt time.Time    `json:"expires_at"` // Expiry of the access token
}

// MessageResponse is for simple status messages.
type MessageResponse struct {
	Message string `json:"message"`
}
