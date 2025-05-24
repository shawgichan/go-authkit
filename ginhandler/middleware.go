package ginhandler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/shawgichan/go-authkit/config" // Adjust import path
	"github.com/shawgichan/go-authkit/core"   // Adjust import path
	"github.com/shawgichan/go-authkit/token"  // Adjust import path
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload" // Key for storing payload in Gin context
)

// AuthMiddleware creates a Gin middleware for request authorization.
// It verifies the access token and checks the user's status and active session.
func AuthMiddleware(tokenMaker token.Maker, userStorer core.UserStorer, cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is not provided", nil)
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format", nil)
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			errDetails := fmt.Sprintf("Unsupported authorization type: %s", authorizationType)
			RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unsupported authorization type", errDetails)
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			sdkErr := core.ErrTokenInvalid             // Default to invalid
			if errors.Is(err, token.ErrExpiredToken) { // Assuming your token.Maker returns a specific error for expiration
				sdkErr = core.ErrTokenExpired
			}
			MapSDKErrorToHTTP(c, sdkErr) // Use the mapper for consistent error responses
			return
		}

		// Fetch user from UserStorer using UserID from token payload
		user, err := userStorer.GetUserByID(c.Request.Context(), payload.UserID)
		if err != nil {
			if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrUserDeleted) {
				RespondWithError(c, http.StatusUnauthorized, "USER_NOT_FOUND", "Authenticated user not found or has been deleted", nil)
				return
			}
			MapSDKErrorToHTTP(c, err) // Handle other potential DB errors
			return
		}

		// Check user status
		if user.Status != core.StatusActive {
			var sdkErr error
			switch user.Status {
			case core.StatusPending:
				sdkErr = core.ErrUserNotVerified
			case core.StatusSuspended:
				sdkErr = core.ErrUserSuspended
			default: // Other non-active statuses
				sdkErr = errors.New("user account is not active") // Generic non-active
			}
			MapSDKErrorToHTTP(c, sdkErr)
			return
		}

		// Enforce single-device login if configured
		if cfg.EnforceSingleDeviceLogin {
			if user.ActiveToken == "" { // User has no active token, meaning they've been logged out elsewhere or session expired
				RespondWithError(c, http.StatusUnauthorized, "SESSION_EXPIRED", "Session expired, please login again (no active token)", nil)
				return
			}
			if user.ActiveToken != accessToken {
				RespondWithError(c, http.StatusUnauthorized, "MULTI_DEVICE_LOGIN", "User logged in with a different device or session", nil)
				return
			}
		}

		// Set payload in context for downstream handlers
		c.Set(AuthorizationPayloadKey, payload)
		c.Next()
	}
}

// RoleMiddleware creates a Gin middleware for role-based access control.
// It should be used *after* AuthMiddleware.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, exists := GetAuthPayload(c)
		if !exists {
			// This should ideally not happen if AuthMiddleware ran successfully
			RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization payload not found for role check", nil)
			return
		}

		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if strings.EqualFold(payload.Role, allowedRole) {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			errDetails := fmt.Sprintf("Access denied. Your role '%s' is not in allowed roles: %v", payload.Role, allowedRoles)
			RespondWithError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource", errDetails)
			return
		}
		c.Next()
	}
}

// GetAuthPayload retrieves the authorization payload from the Gin context.
// Returns the payload and true if found, otherwise nil and false.
func GetAuthPayload(c *gin.Context) (*token.Payload, bool) {
	payloadVal, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, false
	}
	payload, ok := payloadVal.(*token.Payload) // Ensure type assertion from your token package
	if !ok {
		// This would indicate a programming error (wrong type stored in context)
		return nil, false
	}
	return payload, true
}
