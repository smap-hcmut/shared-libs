"""
Logger configuration for Python services with trace integration.
"""

from dataclasses import dataclass
from enum import Enum
from typing import Optional


class LogLevel(str, Enum):
    """Log levels supported by the logger."""
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARNING = "WARNING"
    ERROR = "ERROR"
    CRITICAL = "CRITICAL"


@dataclass
class LoggerConfig:
    """
    Logger configuration with trace integration support.
    
    Attributes:
        level: Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        enable_console: Enable console output
        colorize: Enable colored console output (development mode)
        json_output: Enable JSON structured output (production mode)
        service_name: Service name for structured logging
        enable_trace_id: Enable automatic trace_id injection
        enable_request_id: Enable request_id tracking (optional)
    """
    
    level: LogLevel = LogLevel.INFO
    enable_console: bool = True
    colorize: bool = True
    json_output: bool = False
    service_name: str = "python-service"
    enable_trace_id: bool = True  # Default: enabled for trace integration
    enable_request_id: bool = False  # Optional request tracking
    
    def __post_init__(self):
        """Validate configuration after initialization."""
        # Convert string level to enum if needed
        if isinstance(self.level, str):
            try:
                self.level = LogLevel(self.level.upper())
            except ValueError:
                valid_levels = [level.value for level in LogLevel]
                raise ValueError(
                    f"Invalid log level: {self.level}. Must be one of {valid_levels}"
                )
        
        # Validate service name
        if not self.service_name or not isinstance(self.service_name, str):
            raise ValueError("service_name must be a non-empty string")


# Default configurations for different environments
DEFAULT_DEVELOPMENT_CONFIG = LoggerConfig(
    level=LogLevel.DEBUG,
    colorize=True,
    json_output=False,
    enable_trace_id=True,
)

DEFAULT_PRODUCTION_CONFIG = LoggerConfig(
    level=LogLevel.INFO,
    colorize=False,
    json_output=True,
    enable_trace_id=True,
)

DEFAULT_TESTING_CONFIG = LoggerConfig(
    level=LogLevel.WARNING,
    colorize=False,
    json_output=False,
    enable_trace_id=False,  # Simplified for testing
)