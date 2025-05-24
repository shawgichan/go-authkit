package core

import "errors"

var (
	ErrNotFound              = errors.New("requested resource not found")
	ErrDuplicateEmail        = errors.New("email address already in use")
	ErrDuplicateUsername     = errors.New("username already in use")
	ErrInvalidCredentials    = errors.New("invalid credentials provided")
	ErrUserNotVerified       = errors.New("user account is not verified")
	ErrUserSuspended         = errors.New("user account is suspended")
	ErrUserDeleted           = errors.New("user account has been deleted")
	ErrTokenInvalid          = errors.New("token is invalid")
	ErrTokenExpired          = errors.New("token has expired")
	ErrVerificationNotFound  = errors.New("verification data not found or already used")
	ErrPasswordResetNotFound = errors.New("password reset token not found or already used")
	ErrForbidden             = errors.New("action is forbidden for this user")
	// TODO: Add more later
)
