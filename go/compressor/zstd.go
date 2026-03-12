package compressor

import (
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// zstdCompressor implements Compressor interface using Zstd
type zstdCompressor struct {
	config ZstdConfig
}

// NewZstdCompressor creates a new Zstd compressor with configuration
func NewZstdCompressor(cfg ZstdConfig) Compressor {
	return &zstdCompressor{
		config: cfg,
	}
}

// Compress compresses data from reader using Zstd
func (c *zstdCompressor) Compress(r io.Reader, level Level) (io.ReadCloser, error) {
	if level == LevelNone {
		return io.NopCloser(r), nil
	}

	if level == 0 {
		level = c.config.DefaultLevel
	}

	zstdLevel := mapLevel(level)

	// Create pipe for streaming
	pr, pw := io.Pipe()

	// Create encoder
	encoder, err := zstd.NewWriter(pw, zstd.WithEncoderLevel(zstdLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming encoder: %w", err)
	}

	// Start compression in background
	go func() {
		defer pw.Close()
		defer encoder.Close()

		_, err := io.Copy(encoder, r)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("compression error: %w", err))
		}
	}()

	return pr, nil
}

// Decompress decompresses data from reader using Zstd
func (c *zstdCompressor) Decompress(r io.Reader) (io.ReadCloser, error) {
	// Create decoder
	decoder, err := zstd.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming decoder: %w", err)
	}

	return decoder.IOReadCloser(), nil
}

// CompressBytes compresses byte slice using Zstd
func (c *zstdCompressor) CompressBytes(data []byte, level Level) ([]byte, error) {
	if level == LevelNone {
		return data, nil
	}

	if level == 0 {
		level = c.config.DefaultLevel
	}

	zstdLevel := mapLevel(level)

	// Create encoder with specified level
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstdLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}
	defer encoder.Close()

	// Compress data
	compressed := encoder.EncodeAll(data, make([]byte, 0, len(data)))

	return compressed, nil
}

// DecompressBytes decompresses byte slice using Zstd
func (c *zstdCompressor) DecompressBytes(data []byte) ([]byte, error) {
	// Check if data is compressed
	if !c.IsCompressed(data) {
		return data, nil
	}

	// Create decoder
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}
	defer decoder.Close()

	// Decompress data
	decompressed, err := decoder.DecodeAll(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	return decompressed, nil
}

// IsCompressed checks if data is Zstd compressed by checking magic bytes
func (c *zstdCompressor) IsCompressed(data []byte) bool {
	// Zstd magic number: 0x28, 0xB5, 0x2F, 0xFD
	if len(data) < 4 {
		return false
	}

	return data[0] == 0x28 && data[1] == 0xB5 && data[2] == 0x2F && data[3] == 0xFD
}

func mapLevel(level Level) zstd.EncoderLevel {
	switch level {
	case LevelFastest:
		return zstd.SpeedFastest
	case LevelDefault:
		return zstd.SpeedDefault
	case LevelBest:
		return zstd.SpeedBestCompression
	default:
		return zstd.SpeedDefault
	}
}
