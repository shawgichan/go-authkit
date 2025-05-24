package token

import (
	"time"

	"github.com/google/uuid"
)

type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(userID uuid.UUID, username string, role string, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
