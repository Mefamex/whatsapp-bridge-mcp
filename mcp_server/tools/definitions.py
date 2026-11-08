"""Tool definitions for WhatsApp Bridge MCP Server."""

from mcp.types import Tool


TOOLS = [
    Tool(
        name="whatsapp_send",
        description="Send a text message to a WhatsApp contact or group",
        inputSchema={
            "type": "object",
            "properties": {
                "to": {
                    "type": "string",
                    "description": "Phone number (905xxxxxxxxx) or JID (905xxxxxxxxx@s.whatsapp.net)",
                },
                "text": {"type": "string", "description": "Message text to send"},
            },
            "required": ["to", "text"],
        },
    ),
    Tool(
        name="whatsapp_send_media",
        description="Send media (image, video, audio, document) to a WhatsApp contact or group",
        inputSchema={
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Phone number or JID"},
                "media_base64": {
                    "type": "string",
                    "description": "Base64 encoded media file",
                },
                "mime_type": {
                    "type": "string",
                    "description": "MIME type: image/jpeg, video/mp4, audio/ogg, application/pdf, etc.",
                },
                "caption": {
                    "type": "string",
                    "description": "Optional caption for media",
                },
                "file_name": {
                    "type": "string",
                    "description": "Optional file name for documents",
                },
            },
            "required": ["to", "media_base64", "mime_type"],
        },
    ),
    Tool(
        name="whatsapp_send_poll",
        description="Send a poll (survey) message with 2-12 options to a WhatsApp contact or group",
        inputSchema={
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Phone number or JID"},
                "name": {"type": "string", "description": "Poll question/title"},
                "options": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "2-12 answer options",
                    "minItems": 2,
                    "maxItems": 12,
                },
                "selectable_count": {
                    "type": "integer",
                    "description": "How many options can be selected (default 1)",
                    "default": 1,
                },
            },
            "required": ["to", "name", "options"],
        },
    ),
    Tool(
        name="whatsapp_get_contacts",
        description="Get all WhatsApp contacts",
        inputSchema={"type": "object", "properties": {}},
    ),
    Tool(
        name="whatsapp_get_groups",
        description="Get all WhatsApp groups you're a member of",
        inputSchema={"type": "object", "properties": {}},
    ),
    Tool(
        name="whatsapp_get_messages",
        description="Get recent messages from a chat (from in-memory store)",
        inputSchema={
            "type": "object",
            "properties": {
                "jid": {
                    "type": "string",
                    "description": "Chat JID (phone@s.whatsapp.net or group@g.us)",
                },
                "count": {
                    "type": "integer",
                    "description": "Number of messages to retrieve",
                    "default": 50,
                    "maximum": 200,
                },
            },
            "required": ["jid"],
        },
    ),
    Tool(
        name="whatsapp_get_media",
        description="Download a message attachment (voice note, image, video, document) as base64. Use the message id from whatsapp_get_messages when type is audio/image/video/document.",
        inputSchema={
            "type": "object",
            "properties": {
                "message_id": {
                    "type": "string",
                    "description": "WhatsApp message ID (e.g. from the id field of whatsapp_get_messages)",
                },
            },
            "required": ["message_id"],
        },
    ),
    Tool(
        name="whatsapp_mark_read",
        description="Mark messages as read (send read receipt)",
        inputSchema={
            "type": "object",
            "properties": {
                "jid": {"type": "string", "description": "Chat JID"},
                "message_ids": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "List of message IDs to mark as read",
                },
            },
            "required": ["jid", "message_ids"],
        },
    ),
    Tool(
        name="whatsapp_health",
        description="Check WhatsApp bridge connection status",
        inputSchema={"type": "object", "properties": {}},
    ),
    Tool(
        name="whatsapp_set_webhook",
        description="Register a webhook URL to receive real-time events",
        inputSchema={
            "type": "object",
            "properties": {
                "url": {
                    "type": "string",
                    "description": "Webhook URL (POST receives JSON events)",
                },
            },
            "required": ["url"],
        },
    ),
    Tool(
        name="whatsapp_delete_webhook",
        description="Remove the registered webhook",
        inputSchema={"type": "object", "properties": {}},
    ),
    Tool(
        name="whatsapp_ask_confirm",
        description="Send a yes/no confirmation question to a WhatsApp contact and return immediately. The user's response can be retrieved later with whatsapp_get_messages.",
        inputSchema={
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Phone number or JID"},
                "question": {
                    "type": "string",
                    "description": "Confirmation question to ask",
                },
                "yes_text": {
                    "type": "string",
                    "description": "Text for yes option (default: 'Evet')",
                    "default": "Evet",
                },
                "no_text": {
                    "type": "string",
                    "description": "Text for no option (default: 'Hayır')",
                    "default": "Hayır",
                },
            },
            "required": ["to", "question"],
        },
    ),
    Tool(
        name="whatsapp_reaction",
        description="Add or remove an emoji reaction on a WhatsApp message. Send empty string to remove reaction.",
        inputSchema={
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Phone number or JID of the chat"},
                "message_id": {"type": "string", "description": "Target WhatsApp message ID to react to"},
                "emoji": {"type": "string", "description": "Single emoji character (e.g. 👍, ❤️, 🔥, 🚀) or empty string to remove"},
                "sender_jid": {"type": "string", "description": "Optional JID of original message sender if in a group"},
            },
            "required": ["to", "message_id", "emoji"],
        },
    ),
    Tool(
        name="whatsapp_revoke",
        description="Delete a previously sent message for everyone (unsend message).",
        inputSchema={
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Phone number or JID of the chat"},
                "message_id": {"type": "string", "description": "Message ID of the message to revoke/delete"},
            },
            "required": ["to", "message_id"],
        },
    ),
]


TOOL_MAP = {tool.name: tool for tool in TOOLS}
