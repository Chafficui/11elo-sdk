"""Custom exceptions for the 11elo Python client."""


class ElevenEloError(Exception):
    """Base exception for all 11elo client errors."""


class AuthenticationError(ElevenEloError):
    """Raised when the API key is missing or invalid."""


class RateLimitError(ElevenEloError):
    """Raised when the daily rate limit for the API key tier has been exceeded."""

    def __init__(self, message: str, reset_at: str | None = None) -> None:
        super().__init__(message)
        self.reset_at = reset_at


class NotFoundError(ElevenEloError):
    """Raised when a requested resource does not exist."""


class ApiError(ElevenEloError):
    """Raised for unexpected API errors (4xx / 5xx responses)."""

    def __init__(self, message: str, status_code: int) -> None:
        super().__init__(message)
        self.status_code = status_code
