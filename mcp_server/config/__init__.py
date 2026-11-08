"""Configuration management for WhatsApp Bridge MCP Server."""

import os
from dataclasses import dataclass
from typing import Optional


@dataclass
class Config:
    """Application configuration."""

    api_base: str = "http://127.0.0.1:31415"
    timeout: float = 30.0
    log_level: str = "INFO"

    @classmethod
    def from_env(cls) -> "Config":
        """Load configuration from environment variables."""
        return cls(
            api_base=os.getenv("WHATSAPP_API_BASE", "http://127.0.0.1:31415"),
            timeout=float(os.getenv("WHATSAPP_API_TIMEOUT", "30.0")),
            log_level=os.getenv("WHATSAPP_LOG_LEVEL", "INFO"),
        )


config = Config.from_env()
