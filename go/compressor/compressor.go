package compressor

import (
	"fmt"
	"io"
)

// Backwards compatibility aliases
type (
	CompressionLevel = Level
)

const (
	CompressionNone    = LevelNone
	CompressionFastest = LevelFastest
	CompressionDefault = LevelDefault
	CompressionBest    = LevelBest
)

// Default global compressor instance
var defaultCompressor Compressor

func init() {
	var err error
	defaultCompressor, err = NewCompressor(DefaultConfig())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize default compressor: %v", err))
	}
}

// SetDefaultCompressor sets the global default compressor
func SetDefaultCompressor(c Compressor) {
	defaultCompressor = c
}

// GetDefaultCompressor returns the global default compressor
func GetDefaultCompressor() Compressor {
	return defaultCompressor
}

// Compress compresses data using the default compressor
func Compress(r io.Reader, level Level) (io.ReadCloser, error) {
	return defaultCompressor.Compress(r, level)
}

// Decompress decompresses data using the default compressor
func Decompress(r io.Reader) (io.ReadCloser, error) {
	return defaultCompressor.Decompress(r)
}

// CompressBytes compresses byte slice using the default compressor
func CompressBytes(data []byte, level Level) ([]byte, error) {
	return defaultCompressor.CompressBytes(data, level)
}

// DecompressBytes decompresses byte slice using the default compressor
func DecompressBytes(data []byte) ([]byte, error) {
	return defaultCompressor.DecompressBytes(data)
}

// IsCompressed checks if data is compressed using the default compressor
func IsCompressed(data []byte) bool {
	return defaultCompressor.IsCompressed(data)
}
