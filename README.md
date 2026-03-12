# SMAP Shared Libraries

Unified shared library repository for SMAP social media analytics platform, providing consolidated packages with comprehensive distributed tracing capabilities.

## Overview

This repository consolidates duplicate packages across all SMAP services while adding trace_id propagation for end-to-end request tracking. It supports both Go and Python implementations with consistent behavior across all 7 microservices.

## Structure

```
smap-shared-libs/
├── go/                     # Go shared libraries
├── python/                 # Python shared libraries  
├── docs/                   # Cross-language documentation
└── README.md              # This file
```

## Services Supported

- **Go Services**: identity-srv, project-srv, ingest-srv, knowledge-srv, notification-srv
- **Python Services**: analysis-srv, scapper-srv

## Key Features

- **Trace Propagation**: Automatic trace_id management across HTTP, Kafka, and database operations
- **Code Deduplication**: Eliminates duplicate pkg implementations across services
- **Cross-Language Consistency**: Unified behavior between Go and Python services
- **Performance Optimized**: <1ms latency impact per operation
- **Backward Compatible**: Seamless migration from service-specific packages

## Quick Start

### Go Services
```go
import "github.com/smap/shared-libs/go/tracing"
import "github.com/smap/shared-libs/go/log"
import "github.com/smap/shared-libs/go/http"
```

### Python Services
```python
from smap_shared.tracing import TraceContext
from smap_shared.logger import Logger
from smap_shared.http import TracedHTTPClient
```

## Documentation

- [Migration Guide](docs/migration-guide.md) - Step-by-step migration from service-specific packages
- [Tracing Guide](docs/tracing-guide.md) - Comprehensive trace_id management usage
- [Examples](docs/examples/) - Usage examples for all packages

## License

Internal SMAP project - All rights reserved