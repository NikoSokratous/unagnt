"""Streaming support for Agent Runtime Python SDK"""

import asyncio
from typing import AsyncIterator, Dict, Any
import httpx

from .types import Run
from .errors import AgentRuntimeError, APIError


class StreamEvent:
    """Represents a streaming event from the agent runtime"""
    
    def __init__(self, data: Dict[str, Any]):
        self.run_id = data.get("run_id")
        self.step_id = data.get("step_id")
        self.timestamp = data.get("timestamp")
        self.type = data.get("type")
        self.agent = data.get("agent")
        self.data = data.get("data", {})
        self.model = data.get("model", {})
    
    def __repr__(self):
        return f"StreamEvent(type={self.type}, run_id={self.run_id})"


async def stream_events(
    base_url: str,
    run_id: str,
    api_key: str = None,
    timeout: float = 300.0
) -> AsyncIterator[StreamEvent]:
    """
    Stream events for a run using Server-Sent Events.
    
    Args:
        base_url: Base URL of the agent runtime API
        run_id: The run ID to stream events for
        api_key: Optional API key for authentication
        timeout: Request timeout in seconds
    
    Yields:
        StreamEvent objects as they arrive
    
    Example:
        async for event in stream_events("http://localhost:8080", run_id):
            print(f"Event: {event.type}")
    """
    url = f"{base_url.rstrip('/')}/v1/runs/{run_id}/stream"
    
    headers = {
        "Accept": "text/event-stream",
    }
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"
    
    async with httpx.AsyncClient(timeout=timeout) as client:
        try:
            async with client.stream("GET", url, headers=headers) as response:
                if response.status_code != 200:
                    raise APIError(
                        response.status_code,
                        f"Stream failed: {response.text}"
                    )
                
                async for line in response.aiter_lines():
                    # Skip empty lines and comments (heartbeats)
                    if not line or line.startswith(":"):
                        continue
                    
                    # SSE format: "data: {...}"
                    if line.startswith("data: "):
                        data_str = line[6:]
                        
                        try:
                            import json
                            data = json.loads(data_str)
                            yield StreamEvent(data)
                        except json.JSONDecodeError as e:
                            raise AgentRuntimeError(f"Failed to parse event: {e}")
                        
        except httpx.TimeoutException as e:
            raise AgentRuntimeError(f"Stream timeout: {e}")
        except httpx.RequestError as e:
            raise AgentRuntimeError(f"Stream request failed: {e}")
