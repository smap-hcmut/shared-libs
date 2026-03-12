# Compressor Package

The `compressor` package provides a modular and extensible interface for data compression and decompression. It follows the standard architectural patterns used across the shared libraries, featuring a factory-based implementation with support for multiple compression algorithms (currently Zstd).

## Features

- **Standard Interface**: Common `Compressor` interface for all implementations.
- **Factory Pattern**: Easy initialization via `NewCompressor(cfg Config)`.
- **Zstd Support**: High-performance compression using the `klauspost/compress/zstd` library.
- **Streaming Support**: Efficiently handles large data streams using `io.Reader` and `io.Writer` without loading entire payloads into memory.
- **Backward Compatibility**: Maintains support for legacy global functions like `CompressBytes` and `DecompressBytes`.
- **Configurable**: Support for compression levels (None, Fastest, Default, Best).

## New Structure

- `interfaces.go`: Defines the core `Compressor` interface and `Level` type.
- `constants.go`: Defines compression levels and implementation type constants.
- `config.go`: Configuration structures for the compressor and specific implementations.
- `factory.go`: Factory function for creating compressor instances.
- `errors.go`: Package-specific error definitions.
- `zstd.go`: Zstd-specific implementation of the `Compressor` interface.
- `compressor.go`: Global default instance and backward compatibility layer.

## Usage

### Using the Global Instance (Default)

The simplest way is to use the package-level functions which use a default Zstd compressor:

```go
import "github.com/smap-hcmut/shared-libs/go/compressor"

// Compress bytes
compressed, err := compressor.CompressBytes(data, compressor.LevelDefault)

// Decompress bytes
original, err := compressor.DecompressBytes(compressed)

// Stream compression
compressedReader, err := compressor.Compress(dataReader, compressor.LevelFastest)
```

### Creating a Custom Compressor

You can create a custom compressor instance with specific configuration:

```go
cfg := compressor.Config{
    Implementation: compressor.ImplementationZstd,
    Zstd: &compressor.ZstdConfig{
        DefaultLevel: compressor.LevelBest,
    },
}

c, err := compressor.NewCompressor(cfg)
if err != nil {
    // handle error
}

// Use the instance
compressed, err := c.CompressBytes(data, 0) // Uses default level from config
```

### Implementing New Algorithms

To add a new compression algorithm:

1. Define a new implementation constant in `constants.go`.
2. Create a new file (e.g., `gzip.go`) and implement the `Compressor` interface.
3. Update the `NewCompressor` factory in `factory.go`.
