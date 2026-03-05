"""Custom exceptions for Agent Runtime SDK"""


class AgentRuntimeError(Exception):
    """Base exception for Agent Runtime SDK"""
    pass


class APIError(AgentRuntimeError):
    """API request failed"""
    
    def __init__(self, status_code: int, message: str):
        self.status_code = status_code
        self.message = message
        super().__init__(f"API Error {status_code}: {message}")


class NotFoundError(APIError):
    """Resource not found (404)"""
    
    def __init__(self, message: str):
        super().__init__(404, message)


class UnauthorizedError(APIError):
    """Authentication failed (401)"""
    
    def __init__(self, message: str = "Invalid or missing API key"):
        super().__init__(401, message)


class TimeoutError(AgentRuntimeError):
    """Operation timed out"""
    pass
