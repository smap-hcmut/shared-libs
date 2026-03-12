"""
SMAP Shared Python Libraries

Unified shared library for SMAP services with distributed tracing support.
Provides consolidated packages with trace_id propagation for end-to-end request tracking.
"""

__version__ = "1.0.0"
__author__ = "SMAP Team"
__email__ = "team@smap.com"

# Core exports
from .tracing import TraceContext, HTTPPropagator, KafkaPropagator
from .logger import Logger, LoggerConfig, LogLevel

# Import submodules to make them available (only import completed modules)
from . import http

__all__ = [
    "TraceContext",
    "HTTPPropagator", 
    "KafkaPropagator",
    "Logger",
    "LoggerConfig",
    "LogLevel",
    "http",
]