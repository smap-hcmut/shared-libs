package encrypter

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/smap-hcmut/shared-libs/go/tracing"
	"golang.org/x/crypto/bcrypt"
)

// New creates a new Encrypter instance with the provided key
// The key must be 16, 24, or 32 bytes long for AES-128, AES-192, or AES-256 respectively
func New(key string) Encrypter {
	return &implEncrypter{
		key: key,
	}
}

// validateKey checks if the key has a valid length for AES (16, 24, or 32 bytes)
func validateKey(key []byte) error {
	keyLen := len(key)
	if keyLen != AESKeyLen128 && keyLen != AESKeyLen192 && keyLen != AESKeyLen256 {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidKeyLength, keyLen)
	}
	return nil
}

func (e *implEncrypter) createByteKey() ([]byte, error) {
	key := []byte(e.key)
	if err := validateKey(key); err != nil {
		return nil, err
	}
	return key, nil
}

// getGCM creates a GCM cipher from the key
func (e *implEncrypter) getGCM() (cipher.AEAD, error) {
	key, err := e.createByteKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return gcm, nil
}

// Encrypt encrypts a plaintext string using AES-GCM and returns a base64-encoded ciphertext
func (e *implEncrypter) Encrypt(plaintext string) (string, error) {
	return e.EncryptWithTrace(context.Background(), plaintext)
}

// EncryptWithTrace encrypts with trace context
func (e *implEncrypter) EncryptWithTrace(ctx context.Context, plaintext string) (string, error) {
	// Note: We don't log the plaintext for security reasons, only trace the operation
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here without exposing sensitive data
		// Example: log.Info("Encryption operation", "trace_id", traceID, "operation", "encrypt")
	}
	return e.EncryptBytesToString([]byte(plaintext))
}

// Decrypt decrypts a base64-encoded ciphertext string and returns the plaintext
func (e *implEncrypter) Decrypt(ciphertext string) (string, error) {
	return e.DecryptWithTrace(context.Background(), ciphertext)
}

// DecryptWithTrace decrypts with trace context
func (e *implEncrypter) DecryptWithTrace(ctx context.Context, ciphertext string) (string, error) {
	// Note: We don't log the ciphertext or plaintext for security reasons
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here without exposing sensitive data
		// Example: log.Info("Decryption operation", "trace_id", traceID, "operation", "decrypt")
	}
	plaintext, err := e.DecryptStringToBytes(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptBytesToString encrypts a byte slice using AES-GCM and returns a base64-encoded ciphertext
func (e *implEncrypter) EncryptBytesToString(data []byte) (string, error) {
	return e.EncryptBytesToStringWithTrace(context.Background(), data)
}

// EncryptBytesToStringWithTrace encrypts bytes with trace context
func (e *implEncrypter) EncryptBytesToStringWithTrace(ctx context.Context, data []byte) (string, error) {
	gcm, err := e.getGCM()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptStringToBytes decrypts a base64-encoded ciphertext string and returns the plaintext bytes
func (e *implEncrypter) DecryptStringToBytes(ciphertext string) ([]byte, error) {
	return e.DecryptStringToBytesWithTrace(context.Background(), ciphertext)
}

// DecryptStringToBytesWithTrace decrypts to bytes with trace context
func (e *implEncrypter) DecryptStringToBytesWithTrace(ctx context.Context, ciphertext string) ([]byte, error) {
	gcm, err := e.getGCM()
	if err != nil {
		return nil, err
	}

	ciphertextByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextByte) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, ciphertextByte := ciphertextByte[:nonceSize], ciphertextByte[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextByte, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// HashPassword hashes a password using bcrypt with the default cost
func (e *implEncrypter) HashPassword(password string) (string, error) {
	return e.HashPasswordWithTrace(context.Background(), password)
}

// HashPasswordWithTrace hashes password with trace context
func (e *implEncrypter) HashPasswordWithTrace(ctx context.Context, password string) (string, error) {
	// Note: We don't log the password for security reasons, only trace the operation
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here without exposing sensitive data
		// Example: log.Info("Password hashing operation", "trace_id", traceID, "operation", "hash_password")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password with its bcrypt hash
func (e *implEncrypter) CheckPasswordHash(password, hash string) bool {
	return e.CheckPasswordHashWithTrace(context.Background(), password, hash)
}

// CheckPasswordHashWithTrace checks password with trace context
func (e *implEncrypter) CheckPasswordHashWithTrace(ctx context.Context, password, hash string) bool {
	// Note: We don't log the password or hash for security reasons, only trace the operation
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here without exposing sensitive data
		// Example: log.Info("Password verification operation", "trace_id", traceID, "operation", "check_password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
