package compressor

// NewCompressor creates a new compressor with the specified configuration
func NewCompressor(cfg Config) (Compressor, error) {
	switch cfg.Implementation {
	case ImplementationZstd:
		zstdCfg := ZstdConfig{
			DefaultLevel: LevelDefault,
		}
		if cfg.Zstd != nil {
			zstdCfg = *cfg.Zstd
		}
		return NewZstdCompressor(zstdCfg), nil
	default:
		return nil, ErrUnsupportedImplementation
	}
}
