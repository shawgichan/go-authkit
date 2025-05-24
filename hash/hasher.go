package hash

// PasswordHasher defines an interface for hashing and checking passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(hashedPassword, password string) error
}
