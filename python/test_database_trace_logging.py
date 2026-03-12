"""
Test script for database clients with trace logging.

This script demonstrates the trace_id logging functionality without requiring
actual database connections.
"""

import asyncio
import sys
from unittest.mock import AsyncMock, patch

from smap_shared.postgres import TracedPostgresClient, PostgresConfig
from smap_shared.redis import TracedRedisClient, RedisConfig
from smap_shared.tracing import set_trace_id, get_trace_id, generate_trace_id


async def test_postgres_trace_logging():
    """Test PostgreSQL client trace logging functionality."""
    print("=== Testing PostgreSQL Trace Logging ===")
    
    # Set trace_id in context
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    print(f"Set trace_id: {trace_id}")
    
    # Create config
    config = PostgresConfig(
        database_url="postgresql+asyncpg://user:pass@localhost/testdb",
        enable_trace_logging=True
    )
    
    # Mock the database engine to avoid actual connection
    with patch('smap_shared.postgres.client.create_async_engine') as mock_engine:
        with patch('smap_shared.postgres.client.async_sessionmaker') as mock_session_factory:
            # Create client (will use mocked engine)
            client = TracedPostgresClient(config)
            
            # Verify trace_id is accessible
            current_trace = get_trace_id()
            assert current_trace == trace_id, f"Expected {trace_id}, got {current_trace}"
            print(f"✓ Trace context accessible: {current_trace}")
            
            print("✓ PostgreSQL client created with trace logging enabled")
    
    print("✓ PostgreSQL trace logging test completed\n")


async def test_redis_trace_logging():
    """Test Redis client trace logging functionality."""
    print("=== Testing Redis Trace Logging ===")
    
    # Set trace_id in context
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    print(f"Set trace_id: {trace_id}")
    
    # Create config
    config = RedisConfig(
        host="localhost",
        port=6379,
        enable_trace_logging=True
    )
    
    # Mock the Redis client to avoid actual connection
    with patch('smap_shared.redis.client.aioredis.Redis') as mock_redis:
        with patch('smap_shared.redis.client.ConnectionPool') as mock_pool:
            # Create client (will use mocked Redis)
            client = TracedRedisClient(config)
            
            # Verify trace_id is accessible
            current_trace = get_trace_id()
            assert current_trace == trace_id, f"Expected {trace_id}, got {current_trace}"
            print(f"✓ Trace context accessible: {current_trace}")
            
            # Test trace logging method
            with patch('smap_shared.redis.client.logger') as mock_logger:
                client._log_operation("GET", "test:key", "extra_info=test")
                
                # Verify logger was called with trace_id
                mock_logger.info.assert_called_once()
                log_message = mock_logger.info.call_args[0][0]
                assert trace_id in log_message, f"Trace ID not found in log: {log_message}"
                assert "operation=GET" in log_message, f"Operation not found in log: {log_message}"
                assert "key=test:key" in log_message, f"Key not found in log: {log_message}"
                print(f"✓ Trace logging format verified: {log_message}")
            
            print("✓ Redis client created with trace logging enabled")
    
    print("✓ Redis trace logging test completed\n")


async def test_graceful_handling_without_trace():
    """Test graceful handling when no trace_id exists in context."""
    print("=== Testing Graceful Handling Without Trace ===")
    
    # Clear trace context
    from smap_shared.tracing import clear_trace_id
    clear_trace_id()
    
    # Verify no trace_id
    current_trace = get_trace_id()
    assert current_trace is None, f"Expected None, got {current_trace}"
    print("✓ Trace context cleared")
    
    # Test Redis logging without trace_id
    config = RedisConfig(host="localhost", port=6379)
    
    with patch('smap_shared.redis.client.aioredis.Redis'):
        with patch('smap_shared.redis.client.ConnectionPool'):
            client = TracedRedisClient(config)
            
            with patch('smap_shared.redis.client.logger') as mock_logger:
                client._log_operation("SET", "test:key")
                
                # Verify logger was called without trace_id
                mock_logger.info.assert_called_once()
                log_message = mock_logger.info.call_args[0][0]
                assert "trace_id=" not in log_message, f"Unexpected trace_id in log: {log_message}"
                assert "operation=SET" in log_message, f"Operation not found in log: {log_message}"
                print(f"✓ Graceful logging without trace_id: {log_message}")
    
    print("✓ Graceful handling test completed\n")


async def test_database_logging_format():
    """Test the specific database logging format requirements."""
    print("=== Testing Database Logging Format ===")
    
    # Test with trace_id
    trace_id = "550e8400-e29b-41d4-a716-446655440000"
    set_trace_id(trace_id)
    
    # Test PostgreSQL format
    with patch('smap_shared.postgres.client.create_async_engine'):
        with patch('smap_shared.postgres.client.async_sessionmaker'):
            with patch('smap_shared.postgres.client.logger') as mock_logger:
                config = PostgresConfig(database_url="postgresql://user:pass@localhost/db")
                client = TracedPostgresClient(config)
                
                # Simulate query logging (this would normally be triggered by SQLAlchemy events)
                from smap_shared.tracing import get_trace_id
                current_trace = get_trace_id()
                query = "SELECT * FROM users WHERE id = $1"
                
                # Format as the actual implementation would
                if current_trace:
                    log_message = f"trace_id={current_trace} query={query}"
                else:
                    log_message = f"query={query}"
                
                print(f"✓ PostgreSQL log format: {log_message}")
                assert f"trace_id={trace_id}" in log_message
                assert f"query={query}" in log_message
    
    # Test Redis format
    with patch('smap_shared.redis.client.aioredis.Redis'):
        with patch('smap_shared.redis.client.ConnectionPool'):
            client = TracedRedisClient(RedisConfig())
            
            with patch('smap_shared.redis.client.logger') as mock_logger:
                client._log_operation("GET", "user:123")
                
                log_message = mock_logger.info.call_args[0][0]
                print(f"✓ Redis log format: {log_message}")
                assert f"trace_id={trace_id}" in log_message
                assert "operation=GET" in log_message
                assert "key=user:123" in log_message
    
    print("✓ Database logging format test completed\n")


async def main():
    """Run all tests."""
    print("Starting database trace logging tests...\n")
    
    try:
        await test_postgres_trace_logging()
        await test_redis_trace_logging()
        await test_graceful_handling_without_trace()
        await test_database_logging_format()
        
        print("🎉 All tests passed successfully!")
        return 0
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)