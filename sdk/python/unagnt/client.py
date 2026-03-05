"""Agent Runtime Python Client"""

import time
from typing import Optional
import httpx

from .types import Run, CreateRunRequest, CreateRunResponse, ListRunsResponse
from .errors import APIError, NotFoundError, UnauthorizedError, TimeoutError


class AgentRuntime:
    """Synchronous client for Agent Runtime API"""
    
    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        api_key: Optional[str] = None,
        timeout: float = 30.0,
    ):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout
        self._client = httpx.Client(timeout=timeout)
    
    def __enter__(self):
        return self
    
    def __exit__(self, *args):
        self.close()
    
    def close(self):
        """Close the HTTP client"""
        self._client.close()
    
    def _headers(self) -> dict:
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers
    
    def _request(self, method: str, path: str, **kwargs) -> dict:
        """Make an API request"""
        url = f"{self.base_url}{path}"
        headers = self._headers()
        
        try:
            response = self._client.request(method, url, headers=headers, **kwargs)
            response.raise_for_status()
            return response.json()
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 404:
                raise NotFoundError(e.response.text)
            elif e.response.status_code == 401:
                raise UnauthorizedError()
            else:
                raise APIError(e.response.status_code, e.response.text)
        except httpx.TimeoutException as e:
            raise TimeoutError(f"Request timed out: {e}")
    
    def create_run(self, agent_name: str, goal: str) -> str:
        """Create a new agent run
        
        Args:
            agent_name: Name of the agent to run
            goal: The goal or task for the agent
            
        Returns:
            The run ID
        """
        req = CreateRunRequest(agent_name=agent_name, goal=goal)
        resp = self._request("POST", "/v1/runs", json=req.model_dump())
        return CreateRunResponse(**resp).run_id
    
    def get_run(self, run_id: str) -> Run:
        """Get details of a specific run
        
        Args:
            run_id: The run ID
            
        Returns:
            Run object with details
        """
        data = self._request("GET", f"/v1/runs/{run_id}")
        return Run(**data)
    
    def list_runs(self, limit: int = 100) -> list[str]:
        """List recent runs
        
        Args:
            limit: Maximum number of runs to return
            
        Returns:
            List of run IDs
        """
        data = self._request("GET", f"/v1/runs?limit={limit}")
        return ListRunsResponse(**data).run_ids
    
    def cancel_run(self, run_id: str) -> None:
        """Cancel an ongoing run
        
        Args:
            run_id: The run ID to cancel
        """
        self._request("POST", f"/v1/runs/{run_id}/cancel")
    
    def wait_for_run(
        self,
        run_id: str,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
    ) -> Run:
        """Wait for a run to complete
        
        Args:
            run_id: The run ID to wait for
            poll_interval: Seconds between polls
            timeout: Maximum seconds to wait (None = no timeout)
            
        Returns:
            Completed run object
            
        Raises:
            TimeoutError: If timeout is exceeded
        """
        start = time.time()
        
        while True:
            run = self.get_run(run_id)
            
            if run.state in ("completed", "failed", "cancelled"):
                return run
            
            if timeout and (time.time() - start) > timeout:
                raise TimeoutError(f"Run {run_id} did not complete within {timeout}s")
            
            time.sleep(poll_interval)
    
    def health_check(self) -> bool:
        """Check if the service is healthy
        
        Returns:
            True if healthy, False otherwise
        """
        try:
            response = self._client.get(f"{self.base_url}/health")
            return response.status_code == 200
        except Exception:
            return False


class AsyncAgentRuntime:
    """Asynchronous client for Agent Runtime API"""
    
    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        api_key: Optional[str] = None,
        timeout: float = 30.0,
    ):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout
        self._client = httpx.AsyncClient(timeout=timeout)
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, *args):
        await self.close()
    
    async def close(self):
        """Close the HTTP client"""
        await self._client.aclose()
    
    def _headers(self) -> dict:
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers
    
    async def _request(self, method: str, path: str, **kwargs) -> dict:
        """Make an async API request"""
        url = f"{self.base_url}{path}"
        headers = self._headers()
        
        try:
            response = await self._client.request(method, url, headers=headers, **kwargs)
            response.raise_for_status()
            return response.json()
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 404:
                raise NotFoundError(e.response.text)
            elif e.response.status_code == 401:
                raise UnauthorizedError()
            else:
                raise APIError(e.response.status_code, e.response.text)
        except httpx.TimeoutException as e:
            raise TimeoutError(f"Request timed out: {e}")
    
    async def create_run(self, agent_name: str, goal: str) -> str:
        """Create a new agent run (async)"""
        req = CreateRunRequest(agent_name=agent_name, goal=goal)
        resp = await self._request("POST", "/v1/runs", json=req.model_dump())
        return CreateRunResponse(**resp).run_id
    
    async def get_run(self, run_id: str) -> Run:
        """Get run details (async)"""
        data = await self._request("GET", f"/v1/runs/{run_id}")
        return Run(**data)
    
    async def list_runs(self, limit: int = 100) -> list[str]:
        """List runs (async)"""
        data = await self._request("GET", f"/v1/runs?limit={limit}")
        return ListRunsResponse(**data).run_ids
    
    async def cancel_run(self, run_id: str) -> None:
        """Cancel run (async)"""
        await self._request("POST", f"/v1/runs/{run_id}/cancel")
    
    async def wait_for_run(
        self,
        run_id: str,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
    ) -> Run:
        """Wait for run completion (async)"""
        import asyncio
        
        start = time.time()
        
        while True:
            run = await self.get_run(run_id)
            
            if run.state in ("completed", "failed", "cancelled"):
                return run
            
            if timeout and (time.time() - start) > timeout:
                raise TimeoutError(f"Run {run_id} did not complete within {timeout}s")
            
            await asyncio.sleep(poll_interval)
