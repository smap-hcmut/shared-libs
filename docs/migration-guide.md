# Migration Guide: Service-Specific to Shared Libraries

This guide provides step-by-step instructions for migrating from service-specific packages to the unified shared library.

## Overview

The migration consolidates duplicate packages across all SMAP services while adding comprehensive distributed tracing capabilities. Each service will replace local `pkg/` imports with shared library imports.

## Pre-Migration Checklist

- [ ] Backup current service code
- [ ] Ensure shared library is available in your environment
- [ ] Review breaking changes section
- [ ] Plan rollback strategy

## Go Services Migration

### Step 1: Update go.mod

Replace local pkg imports with shared library:

```go
// Before
module identity-srv
require (
    // local dependencies
)

// After  
module identity-srv
require (
    github.com/smap/shared-libs/go v1.0.0
    // other dependencies
)
```

### Step 2: Update Import Statements

```go
// Before
import "identity-srv/pkg/log"
import "identity-srv/pkg/http"
import "identity-srv/pkg/kafka"

// After
import "github.com/smap/shared-libs/go/log"
import "github.com/smap/shared-libs/go/http"
import "github.com/smap/shared-libs/go/kafka"
```

### Step 3: Update Middleware Integration

```go
// Before
func setupRoutes() *gin.Engine {
    r := gin.Default()
    // existing middleware
    return r
}

// After
import "github.com/smap/shared-libs/go/tracing"

func setupRoutes() *gin.Engine {
    r := gin.Default()
    
    // Add trace middleware
    tracer := tracing.NewTraceContext()
    propagator := tracing.NewHTTPPropagator()
    r.Use(tracing.TraceMiddleware(tracer, propagator))
    
    return r
}
```

### Step 4: Remove Local pkg Directory

After validation:
```bash
rm -rf pkg/
```

## Python Services Migration

### Step 1: Update pyproject.toml

```toml
# Before
[project]
dependencies = [
    # local pkg dependencies
]

# After
[project]
dependencies = [
    "smap-shared-python>=1.0.0",
    # other dependencies
]
```

### Step 2: Update Import Statements

```python
# Before
from pkg.logger import Logger
from pkg.kafka import KafkaProducer

# After
from smap_shared.logger import Logger
from smap_shared.kafka import TracedKafkaProducer
```

### Step 3: Update FastAPI Integration

```python
# Before
from fastapi import FastAPI
app = FastAPI()

# After
from fastapi import FastAPI
from smap_shared.tracing import trace_middleware

app = FastAPI()
app.middleware("http")(trace_middleware)
```

### Step 4: Remove Local pkg Directory

After validation:
```bash
rm -rf pkg/
```

## Service-Specific Migration Notes

### identity-srv
- Already has basic middleware foundation
- Focus on enhancing existing trace context
- Update Kafka producer for audit logging

### project-srv  
- Critical orchestration service
- Ensure HTTP client enhancement for downstream calls
- Update Kafka producer for event publishing

### knowledge-srv
- Important: HTTP client enhancement for project-srv calls
- Add Qdrant gRPC tracing integration
- Update logging throughout

### analysis-srv
- Focus on Kafka consumer enhancement
- Update database client integration
- Enhance logging with trace context

### ingest-srv
- Update Kafka producer for data pipeline
- Enhance HTTP client for external API calls
- Add trace context to data processing

### notification-srv
- Add WebSocket trace propagation
- Update Redis Pub/Sub integration
- Enhance real-time notification tracing

### scapper-srv
- Enhance RabbitMQ integration
- Update HTTP client for web scraping
- Add trace context to worker processes

## Breaking Changes

### Interface Changes
- Logger interface now requires context parameter for trace-aware methods
- HTTP client methods now require context parameter
- Kafka producer/consumer interfaces enhanced with trace support

### Behavioral Changes
- Automatic trace_id injection in logs (new field added)
- HTTP requests automatically include X-Trace-Id header
- Kafka messages automatically include trace headers

## Validation Steps

### 1. Compilation Check
```bash
go build ./...  # For Go services
pip install -e .  # For Python services
```

### 2. Unit Tests
```bash
go test ./...  # For Go services
pytest  # For Python services
```

### 3. Integration Tests
- Test HTTP service-to-service calls
- Test Kafka message flow
- Test database operations
- Verify trace_id propagation

### 4. Performance Validation
- Measure latency impact (<1ms requirement)
- Check memory usage
- Monitor trace propagation success rates

## Rollback Procedure

If issues arise during migration:

### 1. Revert Import Changes
```bash
git checkout HEAD~1 go.mod  # Go services
git checkout HEAD~1 pyproject.toml  # Python services
```

### 2. Restore Local pkg Directory
```bash
git checkout HEAD~1 pkg/
```

### 3. Rebuild and Test
```bash
go build ./...
go test ./...
```

## Troubleshooting

### Common Issues

**Import Resolution Errors**
- Ensure shared library is properly installed
- Check go.mod/pyproject.toml syntax
- Verify network access to library repository

**Context Propagation Issues**
- Ensure context is passed through all function calls
- Check goroutine/async task context inheritance
- Verify middleware is properly configured

**Performance Degradation**
- Check trace_id validation performance
- Monitor context overhead
- Verify efficient header injection

**Trace ID Not Propagating**
- Verify middleware order
- Check header extraction logic
- Ensure context is properly passed

### Getting Help

- Check [Tracing Guide](tracing-guide.md) for usage patterns
- Review [Examples](examples/) for implementation patterns
- Contact platform team for migration support

## Post-Migration Checklist

- [ ] All services compile successfully
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Trace_id propagates end-to-end
- [ ] Performance requirements met (<1ms impact)
- [ ] Monitoring and alerting configured
- [ ] Documentation updated
- [ ] Team trained on new interfaces