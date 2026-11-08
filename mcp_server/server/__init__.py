"""MCP Server setup for WhatsApp Bridge."""

import logging
from typing import Any, Dict, Optional

from mcp.server.lowlevel.server import Server, ServerRequestContext
from mcp.server.stdio import stdio_server
from mcp.types import (
    CallToolRequestParams,
    CallToolResult,
    EmptyResult,
    ListPromptsResult,
    ListResourcesResult,
    ListToolsResult,
    PaginatedRequestParams,
    ReadResourceResult,
    SubscribeRequestParams,
)

from mcp_server.tools import TOOLS, get_handler

logger = logging.getLogger(__name__)


class WhatsAppBridgeServer:
    """WhatsApp Bridge MCP Server."""

    def __init__(self, api_client):
        self.api_client = api_client
        self.server = self._create_server()

    def _create_server(self) -> Server:
        """Create and configure MCP server."""
        return Server(
            name="whatsapp-bridge",
            on_list_tools=self.handle_list_tools,
            on_call_tool=self.handle_call_tool,
            on_list_resources=self.stub_list_resources,
            on_read_resource=self.stub_read_resource,
            on_subscribe_resource=self.stub_subscribe_resource,
            on_list_prompts=self.stub_list_prompts,
            on_get_prompt=self.stub_get_prompt,
            on_completion=self.stub_complete,
            on_set_logging_level=self.stub_set_logging_level,
            on_ping=self.stub_ping,
            on_roots_list_changed=self.stub_roots_changed,
            on_progress=self.stub_progress,
        )

    # ==================== Tool Handlers ====================

    async def handle_list_tools(
        self,
        context: ServerRequestContext,
        params: Optional[PaginatedRequestParams] = None,
    ) -> ListToolsResult:
        """List available tools."""
        logger.debug("Listing tools")
        return ListToolsResult(tools=TOOLS)

    async def handle_call_tool(
        self,
        context: ServerRequestContext,
        params: CallToolRequestParams,
    ) -> CallToolResult:
        """Handle tool call."""
        name = params.name
        arguments = params.arguments or {}

        logger.debug(f"Calling tool: {name} with args: {arguments}")

        handler_class = get_handler(name)
        if not handler_class:
            return CallToolResult(
                content=[{"type": "text", "text": f"Unknown tool: {name}"}],
                isError=True,
            )

        handler = handler_class(self.api_client)
        return await handler.handle(arguments)

    # ==================== Stub Handlers (Unused Features) ====================

    async def stub_list_resources(self, context, params):
        return ListResourcesResult(resources=[])

    async def stub_read_resource(self, context, params):
        return ReadResourceResult(contents=[])

    async def stub_subscribe_resource(self, context, params):
        return EmptyResult()

    async def stub_list_prompts(self, context, params):
        return ListPromptsResult(prompts=[])

    async def stub_get_prompt(self, context, params):
        return ListPromptsResult(prompts=[])

    async def stub_complete(self, context, params):
        return {"completion": []}

    async def stub_set_logging_level(self, context, params):
        return EmptyResult()

    async def stub_ping(self, context, params):
        return EmptyResult()

    async def stub_roots_changed(self, context, params):
        return None

    async def stub_progress(self, context, params):
        return None

    # ==================== Run ====================

    async def run(self):
        """Run the server."""
        logger.info("Starting WhatsApp Bridge MCP Server")
        async with stdio_server() as (read_stream, write_stream):
            await self.server.run(
                read_stream, write_stream, self.server.create_initialization_options()
            )


def create_server(api_client) -> WhatsAppBridgeServer:
    """Factory function to create server instance."""
    return WhatsAppBridgeServer(api_client)
