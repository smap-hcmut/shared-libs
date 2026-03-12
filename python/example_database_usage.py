"""
Example usage of enhanced database clients with trace logging.

This example demonstrates how to use the migrated PostgreSQL and Redis clients
with automatic trace_id injection in logs.
"""

import asyncio
from smap_shared.postgres import TracedPostgresClient, PostgresConfig
from smap_shared.redis import TracedRedisClient, RedisConfig
from smap_shared.tracing import set_trace_id, generate_trace_id


async def example_postgres_usage():
    """Example of using PostgreSQL client with trace logging."""
    print("=== PostgreSQL Client Example ===")
    
    # Set trace_id for this request
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    print(f"Request trace_id: {trace_id}")
    
    # Create PostgreSQL client
    config = PostgresConfig(
        database_url="postgresql+asyncpg://user:password@localhost:5432/mydb",
        schema="public",
        pool_size=10,
        enable_trace_logging=True
    )
    
    # In a real application, you would use actual database operations:
    print("\n# Example database operations (would log with trace_id):")
    print(f"# trace_id={trace_id} query=SELECT * FROM users WHERE active = true")
    print(f"# trace_id={trace_id} query=INSERT INTO audit_log (action, user_id) VALUES ($1, $2)")
    print(f"# trace_id={trace_id} query=UPDATE user_sessions SET last_seen = NOW() WHERE user_id = $1")
    
    # Demonstrate configuration
    print(f"\nConfiguration:")
    print(f"  Database URL: {config.get_connection_url()}")
    print(f"  Schema: {config.schema}")
    print(f"  Pool Size: {config.pool_size}")
    print(f"  Trace Logging: {config.enable_trace_logging}")


async def example_redis_usage():
    """Example of using Redis client with trace logging."""
    print("\n=== Redis Client Example ===")
    
    # Set trace_id for this request
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    print(f"Request trace_id: {trace_id}")
    
    # Create Redis client
    config = RedisConfig(
        host="localhost",
        port=6379,
        db=0,
        max_connections=50,
        enable_trace_logging=True
    )
    
    # In a real application, you would use actual Redis operations:
    print("\n# Example Redis operations (would log with trace_id):")
    print(f"# trace_id={trace_id} operation=GET key=user:session:123")
    print(f"# trace_id={trace_id} operation=SET key=user:profile:456 ttl=3600")
    print(f"# trace_id={trace_id} operation=DEL key=cache:expired:789")
    print(f"# trace_id={trace_id} operation=INCR key=counter:page_views")
    
    # Demonstrate configuration
    print(f"\nConfiguration:")
    print(f"  Connection URL: {config.get_connection_url()}")
    print(f"  Max Connections: {config.max_connections}")
    print(f"  SSL Enabled: {config.is_ssl_enabled()}")
    print(f"  Trace Logging: {config.enable_trace_logging}")


async def example_migration_from_analysis_srv():
    """Example showing migration from analysis-srv packages."""
    print("\n=== Migration Example ===")
    
    print("Before (analysis-srv/pkg/postgre):")
    print("  from pkg.postgre.postgres import PostgresDatabase")
    print("  from pkg.postgre.type import PostgresConfig")
    print("  # No trace logging")
    
    print("\nAfter (smap-shared-libs):")
    print("  from smap_shared.postgres import TracedPostgresClient, PostgresConfig")
    print("  # Automatic trace_id logging included")
    
    print("\nBefore (analysis-srv/pkg/redis):")
    print("  from pkg.redis.redis import RedisCache")
    print("  from pkg.redis.type import RedisConfig")
    print("  # No trace logging")
    
    print("\nAfter (smap-shared-libs):")
    print("  from smap_shared.redis import TracedRedisClient, RedisConfig")
    print("  # Automatic trace_id logging included")
    
    print("\nBackward Compatibility:")
    print("  # Old aliases still work:")
    print("  from smap_shared.postgres import PostgresDatabase  # alias for TracedPostgresClient")
    print("  from smap_shared.redis import RedisCache  # alias for TracedRedisClient")


async def example_trace_propagation():
    """Example showing trace propagation across operations."""
    print("\n=== Trace Propagation Example ===")
    
    # Simulate incoming request with trace_id
    incoming_trace_id = "550e8400-e29b-41d4-a716-446655440000"
    set_trace_id(incoming_trace_id)
    print(f"Incoming request trace_id: {incoming_trace_id}")
    
    # All database operations will automatically include this trace_id
    print("\nDatabase operations in this request context:")
    print(f"# trace_id={incoming_trace_id} query=SELECT * FROM orders WHERE user_id = $1")
    print(f"# trace_id={incoming_trace_id} operation=GET key=user:cart:123")
    print(f"# trace_id={incoming_trace_id} query=INSERT INTO order_items (...)")
    print(f"# trace_id={incoming_trace_id} operation=SET key=order:status:456 ttl=86400")
    
    print("\nBenefits:")
    print("  ✓ End-to-end request tracking")
    print("  ✓ Correlate database queries with specific requests")
    print("  ✓ Debug performance issues across service boundaries")
    print("  ✓ Audit trail for compliance")


async def example_error_handling():
    """Example showing graceful error handling."""
    print("\n=== Error Handling Example ===")
    
    # Clear trace context to simulate missing trace_id
    from smap_shared.tracing import clear_trace_id
    clear_trace_id()
    print("No trace_id in context")
    
    print("\nDatabase operations without trace_id:")
    print("# query=SELECT * FROM users  (graceful fallback)")
    print("# operation=GET key=session:123  (graceful fallback)")
    
    print("\nError scenarios handled gracefully:")
    print("  ✓ Missing trace_id → Log without trace_id")
    print("  ✓ Invalid trace_id → Generate new trace_id")
    print("  ✓ Logging failures → Continue database operations")
    print("  ✓ Database connection issues → Standard error handling")


async def main():
    """Run all examples."""
    print("Database Clients with Trace Logging - Usage Examples")
    print("=" * 60)
    
    await example_postgres_usage()
    await example_redis_usage()
    await example_migration_from_analysis_srv()
    await example_trace_propagation()
    await example_error_handling()
    
    print("\n" + "=" * 60)
    print("For more information, see:")
    print("  - smap-shared-libs/docs/migration-guide.md")
    print("  - smap-shared-libs/docs/tracing-guide.md")


if __name__ == "__main__":
    asyncio.run(main())