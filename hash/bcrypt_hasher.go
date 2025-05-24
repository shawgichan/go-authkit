package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher uses bcrypt for password hashing.
type BcryptHasher struct {
	Cost int
}

// NewBcryptHasher creates a new BcryptHasher.
// If cost is 0, bcrypt.DefaultCost is used.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{Cost: cost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), h.Cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hashedPassword), nil
}

func (h *BcryptHasher) Check(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
