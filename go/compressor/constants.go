package compressor

const (
	// Compression levels
	LevelNone    Level = 0
	LevelFastest Level = 1
	LevelDefault Level = 2
	LevelBest    Level = 3
)

// Implementation defines the type of compressor implementation
type Implementation string

const (
	// ImplementationZstd uses Zstd compression
	ImplementationZstd Implementation = "zstd"
)
