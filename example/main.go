// go-authkit/example/main.go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	// Adjust import paths to your SDK
	"github.com/shawgichan/go-authkit/config"
	"github.com/shawgichan/go-authkit/core"
	"github.com/shawgichan/go-authkit/ginhandler"
	"github.com/shawgichan/go-authkit/hash"
	"github.com/shawgichan/go-authkit/token"
)

// --- Minimal Mock UserStorer ---
type InMemoryUserStore struct {
	mu         sync.RWMutex
	users      map[uuid.UUID]core.User // Store by User ID
	emailIndex map[string]uuid.UUID    // Email to User ID
	// Simplified verification/reset token storage for example
	verificationTokens  map[string]uuid.UUID // token -> userID
	passwordResetTokens map[string]uuid.UUID // token -> userID
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:               make(map[uuid.UUID]core.User),
		emailIndex:          make(map[string]uuid.UUID),
		verificationTokens:  make(map[string]uuid.UUID),
		passwordResetTokens: make(map[string]uuid.UUID),
	}
}

func (s *InMemoryUserStore) CreateUser(ctx context.Context, params core.CreateUserParams) (core.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.emailIndex[params.Email]; exists {
		return core.User{}, core.ErrDuplicateEmail
	}

	newUser := core.User{
		ID:           uuid.New(),
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		FullName:     params.FullName,
		Role:         params.Role,
		Status:       params.Status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	s.users[newUser.ID] = newUser
	s.emailIndex[newUser.Email] = newUser.ID
	return newUser, nil
}

func (s *InMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (core.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, exists := s.emailIndex[email]
	if !exists {
		return core.User{}, core.ErrNotFound
	}
	user, userExists := s.users[userID]
	if !userExists { // Should not happen if emailIndex is consistent
		return core.User{}, core.ErrNotFound
	}
	return user, nil
}

func (s *InMemoryUserStore) GetUserByID(ctx context.Context, id uuid.UUID) (core.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.users[id]
	if !exists {
		return core.User{}, core.ErrNotFound
	}
	return user, nil
}

func (s *InMemoryUserStore) UpdateUser(ctx context.Context, userID uuid.UUID, params core.UpdateUserParams) (core.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, exists := s.users[userID]
	if !exists {
		return core.User{}, core.ErrNotFound
	}
	if params.FullName != nil {
		user.FullName = *params.FullName
	}
	if params.PasswordHash != nil {
		user.PasswordHash = *params.PasswordHash
	}
	if params.Status != nil {
		user.Status = *params.Status
	}
	if params.ActiveToken != nil {
		user.ActiveToken = *params.ActiveToken
	}
	user.UpdatedAt = time.Now()
	s.users[userID] = user
	return user, nil
}

// --- Simplified Stubs for other UserStorer methods (expand as needed for example) ---
func (s *InMemoryUserStore) StoreVerificationData(ctx context.Context, userID uuid.UUID, email string, token string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verificationTokens[token] = userID
	log.Printf("Mock: Stored verification token %s for user %s", token, userID)
	return nil
}
func (s *InMemoryUserStore) GetVerificationData(ctx context.Context, token string) (uuid.UUID, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, exists := s.verificationTokens[token]
	if !exists {
		return uuid.Nil, "", core.ErrVerificationNotFound
	}
	user, _ := s.users[userID] // In a real scenario, check expiry too
	log.Printf("Mock: Got verification data for token %s, user %s", token, userID)
	return userID, user.Email, nil
}
func (s *InMemoryUserStore) DeleteVerificationData(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.verificationTokens, token)
	log.Printf("Mock: Deleted verification token %s", token)
	return nil
}
func (s *InMemoryUserStore) DeleteVerificationDataByUserID(ctx context.Context, userID uuid.UUID) error { /* ... */
	return nil
}
func (s *InMemoryUserStore) StorePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error { /* ... */
	return nil
}
func (s *InMemoryUserStore) GetPasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) { /* ... */
	return uuid.Nil, errors.New("not implemented in simple mock")
}
func (s *InMemoryUserStore) DeletePasswordResetToken(ctx context.Context, token string) error { /* ... */
	return nil
}

// --- Minimal Mock EmailSender ---
type MockEmailSender struct{}

func (m *MockEmailSender) SendVerificationEmail(ctx context.Context, toEmail, username, verificationLink string) error {
	log.Printf("MOCK EMAIL: To: %s, User: %s, Verification Link: %s\n", toEmail, username, verificationLink)
	return nil
}
func (m *MockEmailSender) SendPasswordResetEmail(ctx context.Context, toEmail, username, resetLink string) error {
	log.Printf("MOCK EMAIL: To: %s, User: %s, Reset Link: %s\n", toEmail, username, resetLink)
	return nil
}

func main() {
	log.Println("Starting go-authkit example...")

	// 1. SDK Config
	sdkConfig := config.DefaultAuthConfig()
	sdkConfig.TokenSymmetricKey = os.Getenv("PASETO_KEY")
	if sdkConfig.TokenSymmetricKey == "" {
		log.Println("Warning: PASETO_KEY not set, using default insecure key for example.")
		sdkConfig.TokenSymmetricKey = "12345678901234567890123456789012" // 32 bytes
	}
	sdkConfig.AppBaseURL = "http://localhost:8080" // For email links

	// 2. SDK Components
	tokenMaker, err := token.NewPasetoMaker(sdkConfig.TokenSymmetricKey)
	if err != nil {
		log.Fatalf("TokenMaker error: %v", err)
	}
	passwordHasher := hash.NewBcryptHasher(0)

	// 3. Mock Implementations
	userStore := NewInMemoryUserStore()
	emailSender := &MockEmailSender{}

	// 4. SDK Gin Handler
	authAPI := ginhandler.NewAuthGinHandler(userStore, tokenMaker, passwordHasher, emailSender, sdkConfig)

	// 5. Gin Router
	router := gin.Default()
	authRoutes := router.Group("/auth")
	{
		// Using SDK's provided handlers
		authRoutes.POST("/register", authAPI.RegisterUser)
		authRoutes.POST("/login", authAPI.LoginUser)
		authRoutes.GET("/verify-email", authAPI.VerifyEmailHandler) // ?token=...
	}

	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(ginhandler.AuthMiddleware(tokenMaker, userStore, sdkConfig))
	{
		protectedRoutes.GET("/me", authAPI.UserInfoHandler)
	}

	log.Println("Example server running on :8080")
	router.Run(":8080")
}
