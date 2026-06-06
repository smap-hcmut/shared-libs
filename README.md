# SMAP Shared Libraries (Go)

Shared Go packages for the SMAP social media analytics platform.
Consolidates common pkg implementations across Go services with consistent tracing,
logging, and HTTP middleware behavior.

## Overview

This repository hosts the Go shared libraries imported by the SMAP Go services
(identity-srv, project-srv, ingest-srv, knowledge-srv, notification-srv).

> Python shared libraries have been removed. Python services (analysis-srv,
> scapper-srv) keep local copies of the helpers they need.

## Structure

```
shared-libs/
├── go/         # Go shared libraries (modules)
└── docs/       # Cross-language documentation (legacy)
```

## Quick Start

```go
import (
    "github.com/smap-hcmut/shared-libs/go/tracing"
    "github.com/smap-hcmut/shared-libs/go/log"
    "github.com/smap-hcmut/shared-libs/go/middleware"
)
```

## Documentation

- [Migration Guide](docs/migration-guide.md) — historical context for the consolidation work.
- [Tracing Guide](docs/tracing-guide.md) — trace_id propagation across HTTP, Kafka, and DB calls.

## License

Internal SMAP project — all rights reserved.
