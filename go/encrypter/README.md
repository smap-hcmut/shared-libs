# Encrypter Package

The encrypter package provides AES-GCM encryption/decryption and bcrypt password hashing with distributed tracing integration for SMAP services.

## Features

- **AES-GCM Encryption**: Secure symmetric encryption with authentication
- **Password Hashing**: bcrypt-based password hashing with salt
- **Trace Integration**: Automatic trace_id propagation (without exposing sensitive data)
- **Multiple Key Sizes**: Support for AES-128, AES-192, and AES-256
- **Base64 Encoding**: Convenient string-based encryption output
- **Backward Compatibility**: Drop-in replacement for existing encrypter packages
- **Security-First**: No sensitive data logged in trace operations

## Supported Encryption

- **AES-128**: 16-byte key
- **AES-192**: 24-byte key  
- **AES-256**: 32-byte key
- **Mode**: GCM (Galois/Counter Mode) with authentication
- **Password Hashing**: bcrypt with default cost

## Usage

### Basic Usage (Backward Compatible)

```go
import "github.com/smap-hcmut/shared-libs/go/encrypter"

// Create encrypter with 32-byte key (AES-256)
key := "your-32-byte-encryption-key-here!"
enc := encrypter.New(key)

// Encrypt/decrypt strings
ciphertext, err := enc.Encrypt("sensitive data")
plaintext, err := enc.Decrypt(ciphertext)

// Encrypt/decrypt bytes
encrypted, err := enc.EncryptBytesToString([]byte("data"))
decrypted, err := enc.DecryptStringToBytes(encrypted)

// Password hashing
hash, err := enc.HashPassword("user-password")
valid := enc.CheckPasswordHash("user-password", hash)
```

### Advanced Usage with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/encrypter"
    "context"
)

// All operations with trace context
ciphertext, err := enc.EncryptWithTrace(ctx, "sensitive data")
plaintext, err := enc.DecryptWithTrace(ctx, ciphertext)

encrypted, err := enc.EncryptBytesToStringWithTrace(ctx, []byte("data"))
decrypted, err := enc.DecryptStringToBytesWithTrace(ctx, encrypted)

hash, err := enc.HashPasswordWithTrace(ctx, "password")
valid := enc.CheckPasswordHashWithTrace(ctx, "password", hash)
```

### Key Management

```go
// Different key sizes for different security levels
key128 := "16-byte-key-here" // AES-128
key192 := "24-byte-key-here-exactly" // AES-192  
key256 := "32-byte-key-here-for-aes-256-enc!" // AES-256

// Create encrypters
enc128 := encrypter.New(key128)
enc192 := encrypter.New(key192)
enc256 := encrypter.New(key256) // Recommended
```

### Error Handling

```go
ciphertext, err := enc.Encrypt("data")
if err != nil {
    switch {
    case errors.Is(err, encrypter.ErrInvalidKeyLength):
        // Handle invalid key length
    case errors.Is(err, encrypter.ErrCiphertextTooShort):
        // Handle corrupted ciphertext
    case errors.Is(err, encrypter.ErrDecryptionFailed):
        // Handle decryption failure
    default:
        // Handle other errors
    }
}
```

### Service Integration

```go
// Configuration
type Config struct {
    EncryptionKey string `yaml:"encryption_key"`
}

// Service setup
func NewService(cfg Config) *Service {
    enc := encrypter.New(cfg.EncryptionKey)
    return &Service{
        encrypter: enc,
    }
}

// Usage in handlers
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // Hash password with trace context
    hashedPassword, err := s.encrypter.HashPasswordWithTrace(ctx, req.Password)
    if err != nil {
        return err
    }
    
    // Encrypt sensitive data
    encryptedSSN, err := s.encrypter.EncryptWithTrace(ctx, req.SSN)
    if err != nil {
        return err
    }
    
    // Store user with encrypted data
    return s.userRepo.Create(ctx, User{
        Username:       req.Username,
        HashedPassword: hashedPassword,
        EncryptedSSN:   encryptedSSN,
    })
}
```

## API Reference

### Interface Methods

#### String Encryption
- `Encrypt(plaintext string) (string, error)`: Encrypt string to base64
- `EncryptWithTrace(ctx, plaintext) (string, error)`: Encrypt with trace context
- `Decrypt(ciphertext string) (string, error)`: Decrypt base64 to string
- `DecryptWithTrace(ctx, ciphertext) (string, error)`: Decrypt with trace context

#### Byte Encryption
- `EncryptBytesToString(data []byte) (string, error)`: Encrypt bytes to base64
- `EncryptBytesToStringWithTrace(ctx, data) (string, error)`: Encrypt bytes with trace
- `DecryptStringToBytes(ciphertext string) ([]byte, error)`: Decrypt base64 to bytes
- `DecryptStringToBytesWithTrace(ctx, ciphertext) ([]byte, error)`: Decrypt with trace

#### Password Hashing
- `HashPassword(password string) (string, error)`: Hash password with bcrypt
- `HashPasswordWithTrace(ctx, password) (string, error)`: Hash with trace context
- `CheckPasswordHash(password, hash string) bool`: Verify password against hash
- `CheckPasswordHashWithTrace(ctx, password, hash) bool`: Verify with trace context

### Constructor
- `New(key string) Encrypter`: Create new encrypter with key

## Constants

### Key Lengths
- `AESKeyLen128`: 16 bytes (AES-128)
- `AESKeyLen192`: 24 bytes (AES-192)
- `AESKeyLen256`: 32 bytes (AES-256)

## Errors

- `ErrInvalidKeyLength`: Key must be 16, 24, or 32 bytes
- `ErrCiphertextTooShort`: Ciphertext too short to contain nonce
- `ErrDecryptionFailed`: Invalid ciphertext or wrong key

## Security Considerations

### Encryption Security
- **AES-GCM**: Provides both confidentiality and authenticity
- **Random Nonces**: Each encryption uses a unique random nonce
- **Key Management**: Store keys securely (environment variables, key vaults)
- **Key Rotation**: Regularly rotate encryption keys

### Password Security
- **bcrypt**: Industry-standard password hashing with salt
- **Default Cost**: Uses bcrypt.DefaultCost (currently 10)
- **Timing Attacks**: bcrypt is designed to be resistant to timing attacks

### Trace Security
- **No Data Logging**: Sensitive data is never logged in trace operations
- **Operation Tracking**: Only operation types are traced, not content
- **Secure by Default**: Trace integration doesn't compromise security

## Migration Guide

### From Local Encrypter Package

1. Update imports:
```go
// Before
import "your-service/pkg/encrypter"

// After
import "github.com/smap-hcmut/shared-libs/go/encrypter"
```

2. No code changes needed for basic usage
3. Optional: Add trace integration for enhanced debugging

### Key Size Validation

The package validates key sizes at runtime:
```go
// These will work
enc := encrypter.New("16-byte-key-here") // AES-128
enc := encrypter.New("24-byte-key-here-exactly") // AES-192
enc := encrypter.New("32-byte-key-here-for-aes-256-enc!") // AES-256

// This will fail at first operation
enc := encrypter.New("invalid-key") // Returns ErrInvalidKeyLength
```

### Trace Integration Benefits

- **Security Auditing**: Track encryption/decryption operations without exposing data
- **Performance Monitoring**: Measure crypto operation latency
- **Debugging**: Easier troubleshooting with trace context
- **Compliance**: Enhanced logging for security compliance (without data exposure)

## Best Practices

1. **Use AES-256**: Prefer 32-byte keys for maximum security
2. **Secure Key Storage**: Never hardcode keys, use environment variables or key vaults
3. **Key Rotation**: Implement regular key rotation procedures
4. **Error Handling**: Always check encryption/decryption errors
5. **Trace Integration**: Use trace-aware methods for better observability
6. **Password Policies**: Enforce strong password requirements before hashing
7. **Constant-Time Comparison**: Use `CheckPasswordHash` for password verification