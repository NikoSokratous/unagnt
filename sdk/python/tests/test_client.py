"""Tests for Agent Runtime Python SDK"""

import pytest
from unagnt import AgentRuntime, AsyncAgentRuntime
from unagnt.errors import APIError


def test_client_creation():
    """Test client instantiation"""
    client = AgentRuntime(base_url="http://localhost:8080", api_key="test-key")
    
    assert client.base_url == "http://localhost:8080"
    assert client.api_key == "test-key"


def test_client_context_manager():
    """Test client as context manager"""
    with AgentRuntime() as client:
        assert client is not None


@pytest.mark.asyncio
async def test_async_client_creation():
    """Test async client instantiation"""
    async with AsyncAgentRuntime(base_url="http://localhost:8080") as client:
        assert client.base_url == "http://localhost:8080"


@pytest.mark.asyncio
async def test_async_create_run():
    """Test async run creation (requires running server)"""
    pytest.skip("Requires running agentd server")
    
    async with AsyncAgentRuntime() as client:
        run_id = await client.create_run("demo-agent", "test goal")
        assert run_id is not None
        assert len(run_id) > 0
