package encrypter

import "errors"

var (
	// ErrInvalidKeyLength is returned when the encryption key has an invalid length
	ErrInvalidKeyLength = errors.New("encryption key must be 16, 24, or 32 bytes long")
	// ErrCiphertextTooShort is returned when the ciphertext is too short to decrypt
	ErrCiphertextTooShort = errors.New("ciphertext is too short")
	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or key")
)
