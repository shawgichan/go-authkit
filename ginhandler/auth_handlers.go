package ginhandler

import (
	// ... other imports
	"github.com/shawgichan/go-authkit/config"
	"github.com/shawgichan/go-authkit/core"
	"github.com/shawgichan/go-authkit/hash"
	"github.com/shawgichan/go-authkit/token"
)

type AuthGinHandler struct {
	store      core.UserStorer
	tokenMaker token.Maker
	hasher     hash.PasswordHasher
	mailer     core.EmailSender // Can be nil if email features aren't used for certain endpoints
	config     *config.AuthConfig
	// logger     // some logger interface
}

func NewAuthGinHandler(
	store core.UserStorer,
	tokenMaker token.Maker,
	hasher hash.PasswordHasher,
	mailer core.EmailSender,
	cfg *config.AuthConfig,
) *AuthGinHandler {
	return &AuthGinHandler{
		store:      store,
		tokenMaker: tokenMaker,
		hasher:     hasher,
		mailer:     mailer,
		config:     cfg,
	}
}
