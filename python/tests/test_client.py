"""Tests for the elevenelo Python client."""

import json
import pytest
import responses as rsps_lib

from elevenelo import Client
from elevenelo.exceptions import (
    ApiError,
    AuthenticationError,
    ElevenEloError,
    NotFoundError,
    RateLimitError,
)

BASE = "https://11elo.com"


@pytest.fixture
def client():
    return Client(api_key="11e_fre_testkey")


# ---------------------------------------------------------------------------
# Constructor validation
# ---------------------------------------------------------------------------

def test_empty_api_key_raises():
    with pytest.raises(ValueError, match="api_key"):
        Client(api_key="")


# ---------------------------------------------------------------------------
# Successful responses
# ---------------------------------------------------------------------------

@rsps_lib.activate
def test_get_teams(client):
    payload = [{"teamName": "Bayern München", "currentElo": 1847}]
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/teams", json=payload)
    result = client.get_teams()
    assert result == payload
    assert rsps_lib.calls[0].request.headers["X-API-Key"] == "11e_fre_testkey"


@rsps_lib.activate
def test_get_team(client):
    payload = {"team": {"name": "Bayern München", "currentElo": 1847}}
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/teams/Bayern%20M%C3%BCnchen", json=payload)
    result = client.get_team("Bayern München")
    assert result == payload


@rsps_lib.activate
def test_get_head_to_head(client):
    payload = [{"date": "2026-02-15", "result": "2:1"}]
    rsps_lib.add(
        rsps_lib.GET,
        f"{BASE}/api/teams/Bayern%20M%C3%BCnchen/head-to-head/Borussia%20Dortmund",
        json=payload,
    )
    result = client.get_head_to_head("Bayern München", "Borussia Dortmund")
    assert result == payload


@rsps_lib.activate
def test_get_matches_with_params(client):
    payload = [{"id": 1, "homeTeam": "Bayern München"}]
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/matches", json=payload)
    result = client.get_matches(season="2024/2025", limit=10, offset=0)
    assert result == payload
    qs = rsps_lib.calls[0].request.url
    assert "season=2024%2F2025" in qs
    assert "limit=10" in qs


@rsps_lib.activate
def test_get_upcoming_matches(client):
    payload = [{"id": 2, "homeTeam": "BVB"}]
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/matches/upcoming", json=payload)
    result = client.get_upcoming_matches(league="BL1", limit=5)
    assert result == payload


@rsps_lib.activate
def test_get_match(client):
    payload = {"match": {"id": 12345, "homeTeam": "Bayern München"}}
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/matches/12345", json=payload)
    result = client.get_match(12345)
    assert result == payload


@rsps_lib.activate
def test_get_seasons(client):
    payload = {"seasons": ["2025/2026", "2024/2025"], "latestSeason": "2025/2026"}
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/seasons", json=payload)
    result = client.get_seasons()
    assert result == payload


@rsps_lib.activate
def test_get_season(client):
    payload = [{"teamName": "Bayern München", "change": 27}]
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/seasons/2024%2F2025", json=payload)
    result = client.get_season("2024/2025", league="BL1")
    assert result == payload


@rsps_lib.activate
def test_get_comparison_history(client):
    payload = {
        "Bayern München": [{"Date": 1709856000000, "ELO": 1847}],
        "Borussia Dortmund": [{"Date": 1709856000000, "ELO": 1720}],
    }
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/comparison/history", json=payload)
    result = client.get_comparison_history(["Bayern München", "Borussia Dortmund"])
    assert result == payload
    qs = rsps_lib.calls[0].request.url
    assert "teams=" in qs


# ---------------------------------------------------------------------------
# Error handling
# ---------------------------------------------------------------------------

@rsps_lib.activate
def test_authentication_error(client):
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/teams", status=401)
    with pytest.raises(AuthenticationError):
        client.get_teams()


@rsps_lib.activate
def test_rate_limit_error(client):
    rsps_lib.add(
        rsps_lib.GET,
        f"{BASE}/api/teams",
        status=429,
        headers={"X-RateLimit-Reset": "2026-03-13T00:00:00Z"},
    )
    with pytest.raises(RateLimitError) as exc_info:
        client.get_teams()
    assert exc_info.value.reset_at == "2026-03-13T00:00:00Z"


@rsps_lib.activate
def test_not_found_error(client):
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/teams/Unknown%20FC", status=404)
    with pytest.raises(NotFoundError):
        client.get_team("Unknown FC")


@rsps_lib.activate
def test_api_error(client):
    rsps_lib.add(rsps_lib.GET, f"{BASE}/api/teams", status=500, body="Internal Server Error")
    with pytest.raises(ApiError) as exc_info:
        client.get_teams()
    assert exc_info.value.status_code == 500


def test_comparison_history_requires_two_teams(client):
    with pytest.raises(ValueError):
        client.get_comparison_history(["Bayern München"])
