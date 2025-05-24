package config

import "time"

// AuthConfig holds configuration for the auth SDK.
type AuthConfig struct {
	TokenSymmetricKey   string // For Paseto/JWT
	AccessTokenDuration time.Duration
	//RefreshTokenDuration           time.Duration // TODO: later when we implement refresh tokens
	PasswordResetTokenDuration     time.Duration
	EmailVerificationTokenDuration time.Duration

	AppBaseURL string //  for constructing email links

	// Role definitions // TODO: will add more later
	DefaultUserRole string
	AdminRole       string

	EnforceSingleDeviceLogin bool
}

// DefaultConfig returns a config with sensible defaults.
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		AccessTokenDuration:            time.Hour * 24,
		PasswordResetTokenDuration:     time.Hour * 1,
		EmailVerificationTokenDuration: time.Hour * 24,
		DefaultUserRole:                "user",
		AdminRole:                      "admin",
		EnforceSingleDeviceLogin:       true,
		AppBaseURL:                     "http://localhost:3000", // Placeholder
	}
}
