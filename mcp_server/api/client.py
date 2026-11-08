"""HTTP API client for WhatsApp Bridge."""

import httpx
from typing import Any, Dict, Optional

from ..config import config


class APIClient:
    """Async HTTP client for WhatsApp Bridge API."""

    def __init__(self, base_url: Optional[str] = None, timeout: Optional[float] = None):
        self.base_url = base_url or config.api_base
        self.timeout = timeout or config.timeout
        self._client: Optional[httpx.AsyncClient] = None

    async def __aenter__(self) -> "APIClient":
        self._client = httpx.AsyncClient(base_url=self.base_url, timeout=self.timeout)
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self._client:
            await self._client.aclose()

    async def _request(
        self, method: str, path: str, json_data: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Make HTTP request with error handling."""
        if not self._client:
            raise RuntimeError("APIClient not initialized. Use async context manager.")

        url = f"{self.base_url}{path}"
        try:
            response = await self._client.request(method, url, json=json_data)
            response.raise_for_status()
            return response.json() if response.content else {"ok": True}
        except httpx.HTTPStatusError as e:
            raise APIError(
                status_code=e.response.status_code,
                message=e.response.text,
                path=path,
            ) from e
        except httpx.RequestError as e:
            raise APIError(
                status_code=0,
                message=f"Request failed: {str(e)}",
                path=path,
            ) from e

    async def get(self, path: str) -> Dict[str, Any]:
        return await self._request("GET", path)

    async def get_bytes(self, path: str) -> tuple[bytes, str]:
        """GET raw bytes (e.g. /media/{id}). Returns (data, content_type)."""
        if not self._client:
            raise RuntimeError("APIClient not initialized. Use async context manager.")

        url = f"{self.base_url}{path}"
        try:
            response = await self._client.get(url)
            response.raise_for_status()
            return (
                response.content,
                response.headers.get("content-type", "application/octet-stream"),
            )
        except httpx.HTTPStatusError as e:
            raise APIError(
                status_code=e.response.status_code,
                message=e.response.text,
                path=path,
            ) from e
        except httpx.RequestError as e:
            raise APIError(
                status_code=0,
                message=f"Request failed: {str(e)}",
                path=path,
            ) from e

    async def post(self, path: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return await self._request("POST", path, data)

    async def delete(self, path: str) -> Dict[str, Any]:
        return await self._request("DELETE", path)


class APIError(Exception):
    """API request error."""

    def __init__(self, status_code: int, message: str, path: str):
        self.status_code = status_code
        self.path = path
        super().__init__(f"HTTP {status_code} on {path}: {message}")
