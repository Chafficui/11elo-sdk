"""elevenelo – Python client for the 11elo Soccer ELO API."""

from .client import Client
from .exceptions import (
    ApiError,
    AuthenticationError,
    ElevenEloError,
    NotFoundError,
    RateLimitError,
)

__all__ = [
    "Client",
    "ElevenEloError",
    "AuthenticationError",
    "RateLimitError",
    "NotFoundError",
    "ApiError",
]

__version__ = "0.1.0"
