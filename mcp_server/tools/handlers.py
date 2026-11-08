"""Tool handlers for WhatsApp Bridge MCP Server."""

import json
from typing import Any, Dict, Optional

from mcp.types import CallToolResult, TextContent

from ..api.client import APIClient, APIError


class ToolHandler:
    """Base class for tool handlers."""

    def __init__(self, api_client: APIClient):
        self.api = api_client

    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        raise NotImplementedError

    def _success(self, data: Any) -> CallToolResult:
        """Create successful result."""
        text = json.dumps(
            data,
            ensure_ascii=False,
            indent=2 if isinstance(data, (dict, list)) else None,
        )
        return CallToolResult(content=[TextContent(type="text", text=text)])

    def _error(self, message: str) -> CallToolResult:
        """Create error result."""
        return CallToolResult(
            content=[TextContent(type="text", text=f"Error: {message}")], isError=True
        )

    def _handle_api_error(self, e: APIError) -> CallToolResult:
        return CallToolResult(
            content=[
                TextContent(type="text", text=f"HTTP {e.status_code} on {e.path}: {e}")
            ],
            isError=True,
        )

    def _handle_generic_error(self, e: Exception) -> CallToolResult:
        return CallToolResult(
            content=[TextContent(type="text", text=f"Error: {str(e)}")], isError=True
        )


class SendHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post(
                "/send", {"to": arguments["to"], "text": arguments["text"]}
            )
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class SendMediaHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post(
                "/send-media",
                {
                    "to": arguments["to"],
                    "media_base64": arguments["media_base64"],
                    "mime_type": arguments["mime_type"],
                    "caption": arguments.get("caption", ""),
                    "file_name": arguments.get("file_name", ""),
                },
            )
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class SendPollHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post(
                "/send-poll",
                {
                    "to": arguments["to"],
                    "name": arguments["name"],
                    "options": arguments["options"],
                    "selectable_count": arguments.get("selectable_count", 1),
                },
            )
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class GetContactsHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.get("/contacts")
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class GetGroupsHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.get("/groups")
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class GetMessagesHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            jid = arguments["jid"]
            count = arguments.get("count", 50)
            result = await self.api.get(f"/messages/{jid}?count={count}")
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class GetMediaHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        import base64

        try:
            message_id = arguments["message_id"]
            data, content_type = await self.api.get_bytes(f"/media/{message_id}")
            return self._success(
                {
                    "message_id": message_id,
                    "mime_type": content_type,
                    "size_bytes": len(data),
                    "media_base64": base64.b64encode(data).decode("ascii"),
                }
            )
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class MarkReadHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post(
                "/mark-read",
                {"jid": arguments["jid"], "message_ids": arguments["message_ids"]},
            )
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class HealthHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.get("/health")
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class SetWebhookHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post("/webhook", {"url": arguments["url"]})
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class DeleteWebhookHandler(ToolHandler):
    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.delete("/webhook")
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class AskConfirmHandler(ToolHandler):
    """Send a yes/no confirmation question to a WhatsApp contact.

    Returns immediately with a tracking ID. The user's response can be
    retrieved later with whatsapp_get_messages.
    """

    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            to = arguments["to"]
            question = arguments["question"]
            yes_text = arguments.get("yes_text", "Evet")
            no_text = arguments.get("no_text", "Hayır")

            text = f"❓ {question}\n\n1) {yes_text}\n2) {no_text}\n\nYanıt için: 1 veya 2 yaz"

            result = await self.api.post("/send", {"to": to, "text": text})
            return self._success(
                {
                    "sent": True,
                    "to": to,
                    "question": question,
                    "tracking": f"confirm:{to}",
                    "note": "Yanıtı almak için whatsapp_get_messages ile sohbeti kontrol et",
                }
            )
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class AskQuickHandler(ToolHandler):
    """Send a multiple-choice quick question to a WhatsApp contact.

    Returns immediately with a tracking ID. The user's response can be
    retrieved later with whatsapp_get_messages.
    """

    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            to = arguments["to"]
            question = arguments["question"]
            options = arguments["options"]

            if len(options) < 2 or len(options) > 5:
                return self._error("options must have 2-5 items")

            lines = [f"❓ {question}", ""]
            for i, opt in enumerate(options, 1):
                lines.append(f"{i}) {opt}")
            lines.append("")
            lines.append("Yanıt için: numara yaz")

            text = "\n".join(lines)

            result = await self.api.post("/send", {"to": to, "text": text})
            return self._success(
                {
                    "sent": True,
                    "to": to,
                    "question": question,
                    "options": options,
                    "tracking": f"quick:{to}",
                    "note": "Yanıtı almak için whatsapp_get_messages ile sohbeti kontrol et",
                }
            )
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class ReactionHandler(ToolHandler):
    """Add or remove an emoji reaction on a message."""

    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post(
                "/reaction",
                {
                    "to": arguments["to"],
                    "message_id": arguments["message_id"],
                    "emoji": arguments.get("emoji", ""),
                    "sender_jid": arguments.get("sender_jid", ""),
                },
            )
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


class RevokeHandler(ToolHandler):
    """Delete a previously sent message for everyone."""

    async def handle(self, arguments: Dict[str, Any]) -> CallToolResult:
        try:
            result = await self.api.post(
                "/revoke",
                {
                    "to": arguments["to"],
                    "message_id": arguments["message_id"],
                },
            )
            return self._success(result)
        except APIError as e:
            return self._handle_api_error(e)
        except Exception as e:
            return self._handle_generic_error(e)


# Handler registry
HANDLERS = {
    "whatsapp_send": SendHandler,
    "whatsapp_send_media": SendMediaHandler,
    "whatsapp_send_poll": SendPollHandler,
    "whatsapp_reaction": ReactionHandler,
    "whatsapp_revoke": RevokeHandler,
    "whatsapp_get_contacts": GetContactsHandler,
    "whatsapp_get_groups": GetGroupsHandler,
    "whatsapp_get_messages": GetMessagesHandler,
    "whatsapp_get_media": GetMediaHandler,
    "whatsapp_mark_read": MarkReadHandler,
    "whatsapp_health": HealthHandler,
    "whatsapp_set_webhook": SetWebhookHandler,
    "whatsapp_delete_webhook": DeleteWebhookHandler,
    "whatsapp_ask_confirm": AskConfirmHandler,
    "whatsapp_ask_quick": AskQuickHandler,
}


def get_handler(name: str) -> Optional[type]:
    """Get handler class for tool name."""
    return HANDLERS.get(name)
