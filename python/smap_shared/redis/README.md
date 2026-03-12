# Redis Client with Trace Logging

Enhanced Redis client migrated from `analysis-srv/pkg/redis` with automatic trace_id injection in operation logs.

## Features

- **Automatic trace_id logging**: All Redis operations include trace_id when available
- **Graceful fallback**: Continues logging without trace_id when not in context
- **Backward compatibility**: Drop-in replacement for existing RedisCache
- **Async support**: Built on redis-py async client
- **JSON serialization**: Automatic JSON handling for complex data types
- **Connection pooling**: Configurable connection pool with health checks

## Migration from analysis-srv

```python
# Before (analysis-srv/pkg/redis)
from pkg.redis.redis import RedisCache
from pkg.redis.type import RedisConfig

# After (smap-shared-libs)
from smap_shared.redis import TracedRedisClient, RedisConfig
# or use backward compatibility alias:
from smap_shared.redis import RedisCache
```

## Usage

```python
from smap_shared.redis import TracedRedisClient, RedisConfig
from smap_shared.tracing import set_trace_id

# Configure client
config = RedisConfig(
    host="localhost",
    port=6379,
    max_connections=50
)

client = TracedRedisClient(config)

# Set trace context
set_trace_id("550e8400-e29b-41d4-a716-446655440000")

# Use Redis (automatic trace logging)
await client.set("user:123", {"name": "John"})
# Logs: trace_id=550e8400-e29b-41d4-a716-446655440000 operation=SET key=user:123

value = await client.get("user:123")
# Logs: trace_id=550e8400-e29b-41d4-a716-446655440000 operation=GET key=user:123
```

## Log Format

- **With trace_id**: `trace_id={uuid} operation={op} key={key}`
- **Without trace_id**: `operation={op} key={key}`