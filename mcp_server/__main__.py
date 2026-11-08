"""Main entry point for WhatsApp Bridge MCP Server."""

import asyncio
import logging
import sys

from .config import config
from .api.client import APIClient
from mcp_server.server import create_server


def setup_logging():
    """Configure logging."""
    logging.basicConfig(
        level=getattr(logging, config.log_level),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        stream=sys.stderr,
    )


async def main():
    """Main entry point."""
    setup_logging()
    logger = logging.getLogger(__name__)

    logger.info(f"Starting WhatsApp Bridge MCP Server")
    logger.info(f"API Base: {config.api_base}")
    logger.info(f"Timeout: {config.timeout}s")

    async with APIClient() as api_client:
        server = create_server(api_client)
        await server.run()


if __name__ == "__main__":
    asyncio.run(main())
