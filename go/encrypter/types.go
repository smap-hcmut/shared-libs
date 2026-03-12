package encrypter

import "context"

// Encrypter provides encryption, decryption, and password hashing with trace integration
type Encrypter interface {
	// Encrypt encrypts a plaintext string and returns a base64-encoded ciphertext
	Encrypt(plaintext string) (string, error)
	// EncryptWithTrace encrypts with trace context
	EncryptWithTrace(ctx context.Context, plaintext string) (string, error)

	// Decrypt decrypts a base64-encoded ciphertext string and returns the plaintext
	Decrypt(ciphertext string) (string, error)
	// DecryptWithTrace decrypts with trace context
	DecryptWithTrace(ctx context.Context, ciphertext string) (string, error)

	// EncryptBytesToString encrypts a byte slice and returns a base64-encoded ciphertext
	EncryptBytesToString(data []byte) (string, error)
	// EncryptBytesToStringWithTrace encrypts bytes with trace context
	EncryptBytesToStringWithTrace(ctx context.Context, data []byte) (string, error)

	// DecryptStringToBytes decrypts a base64-encoded ciphertext string and returns plaintext bytes
	DecryptStringToBytes(ciphertext string) ([]byte, error)
	// DecryptStringToBytesWithTrace decrypts to bytes with trace context
	DecryptStringToBytesWithTrace(ctx context.Context, ciphertext string) ([]byte, error)

	// HashPassword hashes a password using bcrypt with the default cost
	HashPassword(password string) (string, error)
	// HashPasswordWithTrace hashes password with trace context
	HashPasswordWithTrace(ctx context.Context, password string) (string, error)

	// CheckPasswordHash compares a password with its bcrypt hash
	CheckPasswordHash(password, hash string) bool
	// CheckPasswordHashWithTrace checks password with trace context
	CheckPasswordHashWithTrace(ctx context.Context, password, hash string) bool
}

// implEncrypter implements Encrypter with trace integration
type implEncrypter struct {
	key string
}
