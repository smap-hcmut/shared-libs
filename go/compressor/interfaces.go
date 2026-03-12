package compressor

import (
	"io"
)

// Level defines the compression level
type Level int

// Compressor defines the interface for compression operations
type Compressor interface {
	// Compress compresses data from reader and returns compressed reader
	Compress(r io.Reader, level Level) (io.ReadCloser, error)

	// Decompress decompresses data from reader and returns decompressed reader
	Decompress(r io.Reader) (io.ReadCloser, error)

	// CompressBytes compresses byte slice and returns compressed bytes
	CompressBytes(data []byte, level Level) ([]byte, error)

	// DecompressBytes decompresses byte slice and returns decompressed bytes
	DecompressBytes(data []byte) ([]byte, error)

	// IsCompressed checks if data is compressed by this compressor's format
	IsCompressed(data []byte) bool
}
