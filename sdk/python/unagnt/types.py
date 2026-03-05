"""Type definitions for Agent Runtime API"""

from datetime import datetime
from typing import Optional, Dict, Any
from pydantic import BaseModel, Field


class Run(BaseModel):
    """Represents an agent run"""
    
    run_id: str = Field(..., description="Unique run identifier")
    agent_name: str = Field(..., description="Name of the agent")
    goal: str = Field(..., description="The goal or task")
    state: str = Field(..., description="Current state (pending, running, completed, failed, cancelled)")
    step_count: int = Field(default=0, description="Number of steps executed")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    

class CreateRunRequest(BaseModel):
    """Request to create a new run"""
    
    agent_name: str = Field(..., description="Name of the agent to run")
    goal: str = Field(..., description="The goal or task for the agent")


class CreateRunResponse(BaseModel):
    """Response from creating a run"""
    
    run_id: str = Field(..., description="The created run ID")


class ListRunsResponse(BaseModel):
    """Response from listing runs"""
    
    run_ids: list[str] = Field(default_factory=list, description="List of run IDs")
