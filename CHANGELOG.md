# Changelog

All notable changes to the SMAP Shared Libraries will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-12

### Added
- Initial release of unified shared library
- Core tracing libraries for Go and Python
- Enhanced logging with automatic trace_id injection
- HTTP client with automatic trace propagation
- Kafka producer/consumer with trace headers
- Database clients with trace logging
- Redis client with trace context
- Authentication utilities with trace integration
- Comprehensive documentation and migration guides

### Features
- UUID v4 trace_id generation and validation
- Cross-language trace propagation (Go ↔ Python)
- HTTP middleware for automatic trace extraction
- Kafka message header trace management
- Context-aware database operation logging
- Performance optimized (<1ms latency impact)
- Backward compatible with existing service packages

### Migration
- Consolidates duplicate packages from all 7 SMAP services
- Provides step-by-step migration guide
- Maintains interface compatibility during transition
- Includes rollback procedures for safe migration

## [Unreleased]

### Planned
- Performance monitoring and metrics
- Advanced trace analytics
- Cross-service trace visualization
- Enhanced error handling and recovery