# PostgreSQL Client with Trace Logging

Enhanced PostgreSQL client migrated from `analysis-srv/pkg/postgre` with automatic trace_id injection in query logs.

## Features

- **Automatic trace_id logging**: All SQL queries include trace_id when available
- **Graceful fallback**: Continues logging without trace_id when not in context
- **Backward compatibility**: Drop-in replacement for existing PostgresDatabase
- **Async support**: Built on SQLAlchemy async with asyncpg driver
- **Connection pooling**: Configurable connection pool with health checks

## Migration from analysis-srv

```python
# Before (analysis-srv/pkg/postgre)
from pkg.postgre.postgres import PostgresDatabase
from pkg.postgre.type import PostgresConfig

# After (smap-shared-libs)
from smap_shared.postgres import TracedPostgresClient, PostgresConfig
# or use backward compatibility alias:
from smap_shared.postgres import PostgresDatabase
```

## Usage

```python
from smap_shared.postgres import TracedPostgresClient, PostgresConfig
from smap_shared.tracing import set_trace_id

# Configure client
config = PostgresConfig(
    database_url="postgresql+asyncpg://user:pass@localhost/db",
    schema="public",
    pool_size=20
)

client = TracedPostgresClient(config)

# Set trace context
set_trace_id("550e8400-e29b-41d4-a716-446655440000")

# Use database (automatic trace logging)
async with client.get_session() as session:
    result = await session.execute(text("SELECT * FROM users"))
    # Logs: trace_id=550e8400-e29b-41d4-a716-446655440000 query=SELECT * FROM users
```

## Log Format

- **With trace_id**: `trace_id={uuid} query={sql}`
- **Without trace_id**: `query={sql}`