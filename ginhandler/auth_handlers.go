package ginhandler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	// "log" // For debugging, consider using a passed-in logger interface instead

	"github.com/gin-gonic/gin"

	"github.com/shawgichan/go-authkit/config"
	"github.com/shawgichan/go-authkit/core"
	"github.com/shawgichan/go-authkit/hash"
	"github.com/shawgichan/go-authkit/token"
)

// AuthGinHandler provides HTTP handlers for authentication routes.
type AuthGinHandler struct {
	store      core.UserStorer
	tokenMaker token.Maker
	hasher     hash.PasswordHasher
	mailer     core.EmailSender // Can be nil if email features aren't used for certain endpoints
	config     *config.AuthConfig
}

// NewAuthGinHandler creates a new AuthGinHandler.
func NewAuthGinHandler(
	store core.UserStorer,
	tokenMaker token.Maker,
	hasher hash.PasswordHasher,
	mailer core.EmailSender, // mailer can be nil
	cfg *config.AuthConfig,
	// logger your_logger_interface,
) *AuthGinHandler {
	return &AuthGinHandler{
		store:      store,
		tokenMaker: tokenMaker,
		hasher:     hasher,
		mailer:     mailer,
		config:     cfg,
		// logger:  logger,
	}
}

// generateSecureToken is a helper for creating random tokens (e.g., for email verification)
// It's good to have this internal to the handlers or in a util package within the SDK.
func generateSecureTokenInternal(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RegisterUser handles user registration.
func (h *AuthGinHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest // From request_response.go
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body", err.Error())
		return
	}

	// Basic validation (more can be added)
	if len(req.Password) < 8 { // Example: Password length check (could be from h.config)
		RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Password must be at least 8 characters long", nil)
		return
	}

	// Check if user already exists
	_, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil { // User found, means email is taken
		MapSDKErrorToHTTP(c, core.ErrDuplicateEmail)
		return
	}
	if !errors.Is(err, core.ErrNotFound) { // Unexpected error during check
		MapSDKErrorToHTTP(c, err)
		return
	}
	// If core.ErrNotFound, proceed with registration

	hashedPassword, err := h.hasher.Hash(req.Password)
	if err != nil {
		// h.logger.Error("Failed to hash password", "error", err)
		MapSDKErrorToHTTP(c, fmt.Errorf("failed to hash password: %w", err)) // Wrap for more context
		return
	}

	createUserParams := core.CreateUserParams{
		Email:        req.Email,
		Username:     req.Email, // Default username to email, or make configurable
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         h.config.DefaultUserRole, // Assign default role from config
		Status:       core.StatusPending,       // New users start as pending verification
	}

	createdUser, err := h.store.CreateUser(c.Request.Context(), createUserParams)
	if err != nil {
		// h.logger.Error("Failed to create user", "error", err, "email", req.Email)
		MapSDKErrorToHTTP(c, err) // Maps ErrDuplicateEmail etc.
		return
	}

	// Send verification email
	if h.mailer != nil {
		verificationToken, err := generateSecureTokenInternal(16) // 16 bytes -> 32 hex chars
		if err != nil {
			// h.logger.Error("Failed to generate verification token", "error", err)
			// Decide if registration fails or proceeds without email. For now, let it proceed but log.
			fmt.Printf("Warning: Failed to generate verification token for %s: %v\n", createdUser.Email, err)
		} else {
			expiresAt := time.Now().Add(h.config.EmailVerificationTokenDuration)
			err = h.store.StoreVerificationData(c.Request.Context(), createdUser.ID, createdUser.Email, verificationToken, expiresAt)
			if err != nil {
				// h.logger.Error("Failed to store verification data", "error", err, "user_id", createdUser.ID)
				fmt.Printf("Warning: Failed to store verification data for %s: %v\n", createdUser.Email, err)
				// Don't fail registration, but email verification might not work
			} else {
				// Construct verification link (using AppBaseURL from config)
				// The token itself is what's important for the /verify-email handler.
				// The link structure depends on how the frontend/API is set up.
				// Common pattern: frontend URL with token, frontend calls backend /verify-email?token=...
				// Or direct backend link:
				verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s", h.config.AppBaseURL, verificationToken)

				// Send email asynchronously to avoid blocking the request
				go func(mailTo, userNm, link string) {
					// Create a new context for the goroutine or pass one that's appropriate
					bgCtx := context.Background() // Or context.TODO() if unsure
					err := h.mailer.SendVerificationEmail(bgCtx, mailTo, userNm, link)
					if err != nil {
						// h.logger.Error("Failed to send verification email", "error", err, "email", mailTo)
						fmt.Printf("Error sending verification email to %s: %v\n", mailTo, err)
					}
				}(createdUser.Email, createdUser.FullName, verificationLink)
			}
		}
	}

	RespondWithSuccess(c, http.StatusCreated, NewSDKUserResponse(createdUser))
}

// LoginUser handles user login.
func (h *AuthGinHandler) LoginUser(c *gin.Context) {
	var req LoginRequest // From request_response.go
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body", err.Error())
		return
	}

	user, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrUserDeleted) {
			MapSDKErrorToHTTP(c, core.ErrInvalidCredentials) // Generic error for non-existent user
			return
		}
		MapSDKErrorToHTTP(c, err)
		return
	}

	// Check user status before checking password
	if user.Status == core.StatusPending {
		MapSDKErrorToHTTP(c, core.ErrUserNotVerified)
		return
	}
	if user.Status != core.StatusActive { // e.g. suspended
		RespondWithError(c, http.StatusForbidden, "ACCOUNT_INACTIVE", "User account is not active", nil)
		return
	}

	err = h.hasher.Check(user.PasswordHash, req.Password)
	if err != nil { // Password mismatch
		MapSDKErrorToHTTP(c, core.ErrInvalidCredentials)
		return
	}

	// Password is correct, generate token
	accessToken, payload, err := h.tokenMaker.CreateToken(user.ID, user.Username, user.Role, h.config.AccessTokenDuration)
	if err != nil {
		// h.logger.Error("Failed to create access token", "error", err, "user_id", user.ID)
		MapSDKErrorToHTTP(c, fmt.Errorf("failed to create access token: %w", err))
		return
	}

	// If enforcing single device login, store the new access token as the active one
	if h.config.EnforceSingleDeviceLogin {
		updateParams := core.UpdateUserParams{ActiveToken: &accessToken}
		_, updateErr := h.store.UpdateUser(c.Request.Context(), user.ID, updateParams)
		if updateErr != nil {
			// h.logger.Error("Failed to update active token for user", "error", updateErr, "user_id", user.ID)
			// Log error but proceed with login, as token creation was successful.
			// Or, decide if this is a critical failure for login.
			fmt.Printf("Warning: Failed to update active token for %s: %v\n", user.Email, updateErr)
		}
	}

	tokenResponse := TokenResponse{
		AccessToken: accessToken,
		User:        NewSDKUserResponse(user),
		ExpiresAt:   payload.ExpiredAt,
	}
	RespondWithSuccess(c, http.StatusOK, tokenResponse)
}

// VerifyEmailHandler handles the email verification link.
func (h *AuthGinHandler) VerifyEmailHandler(c *gin.Context) {
	var req VerifyEmailRequest                      // From request_response.go, expects token in query
	if err := c.ShouldBindQuery(&req); err != nil { // Use ShouldBindQuery for GET requests
		RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST_QUERY", "Invalid query parameters", err.Error())
		return
	}

	userID, userEmail, err := h.store.GetVerificationData(c.Request.Context(), req.Token)
	if err != nil {
		MapSDKErrorToHTTP(c, err) // Handles ErrVerificationNotFound
		return
	}

	// User found, token is valid (not expired, exists)
	// Mark user as active
	activeStatus := core.StatusActive
	updateParams := core.UpdateUserParams{Status: &activeStatus}
	updatedUser, err := h.store.UpdateUser(c.Request.Context(), userID, updateParams)
	if err != nil {
		// h.logger.Error("Failed to update user status to active", "error", err, "user_id", userID)
		MapSDKErrorToHTTP(c, fmt.Errorf("failed to activate user: %w", err))
		return
	}

	// Delete the verification token now that it's used
	err = h.store.DeleteVerificationData(c.Request.Context(), req.Token)
	if err != nil {
		// h.logger.Error("Failed to delete verification data", "error", err, "token", req.Token)
		// Log error, but verification was successful.
		fmt.Printf("Warning: Failed to delete verification token %s: %v\n", req.Token, err)
	}

	// Optional: Redirect to a success page on the frontend
	// Or return a success message
	// For an API, a JSON response is usually better.
	// frontendSuccessURL := h.config.AppBaseURL + "/email-verified?status=success" // Example
	// c.Redirect(http.StatusFound, frontendSuccessURL)

	RespondWithSuccess(c, http.StatusOK, MessageResponse{Message: fmt.Sprintf("Email for %s successfully verified.", userEmail)})
	fmt.Printf("User %s (ID: %s) successfully verified. Status: %s\n", updatedUser.Email, updatedUser.ID, updatedUser.Status)
}

// UserInfoHandler retrieves information for the authenticated user.
// This handler relies on AuthMiddleware to have run and set the payload.
func (h *AuthGinHandler) UserInfoHandler(c *gin.Context) {
	authPayload, exists := GetAuthPayload(c) // Use helper from middleware.go
	if !exists {
		RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization payload not found", nil)
		return
	}

	// The payload contains UserID. Fetch the full user details.
	user, err := h.store.GetUserByID(c.Request.Context(), authPayload.UserID)
	if err != nil {
		MapSDKErrorToHTTP(c, err) // Handles ErrNotFound etc.
		return
	}

	RespondWithSuccess(c, http.StatusOK, NewSDKUserResponse(user))
}

// --- TODO: Implement other handlers  ---
// - LogoutUserHandler
// - ForgotPasswordHandler
// - ResetPasswordHandler
// - ChangePasswordHandler
// - ChangeNameHandler
