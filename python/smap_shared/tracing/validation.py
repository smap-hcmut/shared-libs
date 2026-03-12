"""
Validation utilities for trace_id management.

Provides UUID v4 validation functions consistent with Go implementation.
"""

import re
from typing import Optional


# UUID v4 validation regex pattern (consistent with Go implementation)
# Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
# Where x is any hexadecimal digit and y is one of 8, 9, A, or B
UUID_V4_PATTERN = re.compile(
    r'^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
)


def validate_uuid_v4(trace_id: str) -> bool:
    """
    Validates if a string is a valid UUID v4 format.
    
    Args:
        trace_id: String to validate
        
    Returns:
        True if valid UUID v4, False otherwise
    """
    if not trace_id or not isinstance(trace_id, str):
        return False
    
    # Convert to lowercase for validation (consistent with Go implementation)
    trace_id_lower = trace_id.lower()
    
    # Check format using regex
    return bool(UUID_V4_PATTERN.match(trace_id_lower))


def is_valid_trace_id(trace_id: Optional[str]) -> bool:
    """
    Checks if a trace_id is valid and non-empty.
    
    Args:
        trace_id: Trace ID to check
        
    Returns:
        True if valid and non-empty, False otherwise
    """
    return bool(trace_id) and validate_uuid_v4(trace_id)


def normalize_trace_id(trace_id: str) -> Optional[str]:
    """
    Normalizes a trace_id to lowercase format if valid.
    
    Args:
        trace_id: Trace ID to normalize
        
    Returns:
        Normalized trace_id or None if invalid
    """
    if not trace_id or not isinstance(trace_id, str):
        return None
    
    trace_id_lower = trace_id.lower()
    
    if validate_uuid_v4(trace_id_lower):
        return trace_id_lower
    
    return None


def sanitize_trace_id(trace_id: Optional[str]) -> Optional[str]:
    """
    Sanitizes and validates a trace_id, returning None for invalid inputs.
    
    Args:
        trace_id: Trace ID to sanitize
        
    Returns:
        Sanitized trace_id or None if invalid
    """
    if not trace_id:
        return None
    
    # Strip whitespace and convert to string
    try:
        clean_trace_id = str(trace_id).strip()
        return normalize_trace_id(clean_trace_id)
    except (ValueError, TypeError):
        return None