#!/usr/bin/env python3
"""WhatsApp Bridge MCP Server Handler Unit Tests."""

import asyncio
import os
import sys
import unittest
from unittest import mock

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from mcp.types import CallToolResult
from mcp_server.tools.handlers import (  # noqa: E402
    SendHandler,
    SendMediaHandler,
    SendPollHandler,
    HealthHandler,
    AskConfirmHandler,
    AskQuickHandler,
    HANDLERS,
)


class MockAPIClient:
    """Mock for APIClient without network calls."""

    def __init__(self, fail=False):
        self.fail = fail
        self.calls = []

    async def get(self, path, params=None):
        self.calls.append(("GET", path, params))
        if self.fail:
            raise Exception("bridge unreachable")
        if path == "/health":
            return {"status": "ok", "paired": True, "connected": True}
        return {"ok": True}

    async def post(self, path, data=None):
        self.calls.append(("POST", path, data))
        if self.fail:
            raise Exception("bridge unreachable")
        return {"ok": True, "id": "TEST_MSG_ID"}


def run(coro):
    return asyncio.run(coro)


class TestMCPHandlers(unittest.TestCase):
    def setUp(self):
        self.mock_api = MockAPIClient()

    def test_send_handler_success(self):
        handler = SendHandler(self.mock_api)
        result = run(handler.handle({"to": "905551234567", "text": "Hello"}))
        self.assertIsInstance(result, CallToolResult)
        self.assertFalse(result.is_error)
        self.assertEqual(len(self.mock_api.calls), 1)
        self.assertEqual(self.mock_api.calls[0][0], "POST")
        self.assertEqual(self.mock_api.calls[0][1], "/send")

    def test_health_handler_success(self):
        handler = HealthHandler(self.mock_api)
        result = run(handler.handle({}))
        self.assertIsInstance(result, CallToolResult)
        self.assertFalse(result.is_error)
        self.assertIn("paired", result.content[0].text)

    def test_send_poll_handler(self):
        handler = SendPollHandler(self.mock_api)
        result = run(
            handler.handle(
                {
                    "to": "905551234567",
                    "name": "Poll Question",
                    "options": ["Option 1", "Option 2"],
                    "selectable_count": 1,
                }
            )
        )
        self.assertIsInstance(result, CallToolResult)
        self.assertFalse(result.is_error)
        self.assertEqual(self.mock_api.calls[0][1], "/send-poll")

    def test_ask_confirm_handler(self):
        handler = AskConfirmHandler(self.mock_api)
        result = run(
            handler.handle(
                {
                    "to": "905551234567",
                    "question": "Deploy production?",
                }
            )
        )
        self.assertIsInstance(result, CallToolResult)
        self.assertFalse(result.is_error)
        self.assertEqual(self.mock_api.calls[0][1], "/send")
        send_data = self.mock_api.calls[0][2]
        self.assertIn("Deploy production?", send_data["text"])
        self.assertIn("Evet", send_data["text"])
        self.assertIn("Hayır", send_data["text"])

    def test_error_handling_when_bridge_down(self):
        failing_api = MockAPIClient(fail=True)
        handler = SendHandler(failing_api)
        result = run(handler.handle({"to": "905551234567", "text": "Fail test"}))
        self.assertIsInstance(result, CallToolResult)
        self.assertTrue(result.is_error)
        self.assertIn("Error:", result.content[0].text)


if __name__ == "__main__":
    unittest.main()
