"""Python client for the 11elo Soccer ELO API.

Basic usage::

    from elevenelo import Client

    client = Client(api_key="11e_fre_your_key_here")

    teams = client.get_teams()
    team  = client.get_team("Bayern München")
    matches = client.get_matches(season="2024/2025", limit=50)
"""

from __future__ import annotations

from typing import Any
from urllib.parse import urlencode, quote

import requests

from .exceptions import (
    ApiError,
    AuthenticationError,
    ElevenEloError,
    NotFoundError,
    RateLimitError,
)

__all__ = ["Client"]

_DEFAULT_BASE_URL = "https://11elo.com"
_DEFAULT_TIMEOUT = 30


class Client:
    """Synchronous HTTP client for the 11elo public API.

    Parameters
    ----------
    api_key:
        Your 11elo API key (format ``11e_<tier>_<hex>``).  Obtain one for free
        at https://11elo.com/developer.
    base_url:
        Override the default API base URL.  Useful for self-hosted instances or
        local development.
    timeout:
        Request timeout in seconds (default: 30).
    session:
        Optional :class:`requests.Session` to use.  A new session is created
        when *None* is passed.
    """

    def __init__(
        self,
        api_key: str,
        base_url: str = _DEFAULT_BASE_URL,
        timeout: int = _DEFAULT_TIMEOUT,
        session: requests.Session | None = None,
    ) -> None:
        if not api_key:
            raise ValueError("api_key must not be empty")
        self._api_key = api_key
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout
        self._session = session or requests.Session()
        self._session.headers.update(
            {
                "X-API-Key": self._api_key,
                "Accept": "application/json",
            }
        )

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _url(self, path: str) -> str:
        return f"{self._base_url}{path}"

    def _get(self, path: str, params: dict[str, Any] | None = None) -> Any:
        """Perform a GET request and return the decoded JSON body."""
        # Remove None values from query params
        if params:
            params = {k: v for k, v in params.items() if v is not None}
        response = self._session.get(
            self._url(path), params=params or {}, timeout=self._timeout
        )
        return self._handle_response(response)

    def _handle_response(self, response: requests.Response) -> Any:
        if response.status_code == 401:
            raise AuthenticationError(
                "Invalid or missing API key.  Obtain one at https://11elo.com/developer"
            )
        if response.status_code == 429:
            reset_at = response.headers.get("X-RateLimit-Reset")
            raise RateLimitError(
                "Daily rate limit exceeded.  Upgrade your plan at https://11elo.com/developer",
                reset_at=reset_at,
            )
        if response.status_code == 404:
            raise NotFoundError(f"Resource not found: {response.url}")
        if not response.ok:
            raise ApiError(
                f"API request failed with status {response.status_code}: {response.text}",
                status_code=response.status_code,
            )
        try:
            return response.json()
        except ValueError as exc:
            raise ElevenEloError(
                f"Failed to parse JSON response: {response.text}"
            ) from exc

    # ------------------------------------------------------------------
    # Teams
    # ------------------------------------------------------------------

    def get_teams(self) -> list[dict[str, Any]]:
        """Return all teams with their current ELO stats and league info.

        Returns
        -------
        list of dict
            Each dict contains keys such as ``teamName``, ``currentElo``,
            ``league``, ``leagueRank``, ``trendEloChange``, etc.

        Example
        -------
        ::

            teams = client.get_teams()
            for team in teams:
                print(team["teamName"], team["currentElo"])
        """
        return self._get("/api/teams")

    def get_team(self, team_name: str) -> dict[str, Any]:
        """Return detailed information for a single team.

        Parameters
        ----------
        team_name:
            The canonical team name (e.g. ``"Bayern München"``).

        Returns
        -------
        dict
            Contains ``team``, ``eloHistory``, ``recentForm``,
            ``significantMatches``, ``stats``, and ``upcomingMatches``.

        Example
        -------
        ::

            team = client.get_team("Bayern München")
            print(team["team"]["currentElo"])
        """
        encoded = quote(team_name, safe="")
        return self._get(f"/api/teams/{encoded}")

    def get_head_to_head(self, team1: str, team2: str) -> list[dict[str, Any]]:
        """Return head-to-head match history between two teams.

        Parameters
        ----------
        team1:
            Name of the first team.
        team2:
            Name of the second team.

        Returns
        -------
        list of dict
            Each entry has ``date``, ``result``, ``winner``, ``homeGoals``,
            ``awayGoals``.

        Example
        -------
        ::

            h2h = client.get_head_to_head("Bayern München", "Borussia Dortmund")
            for match in h2h:
                print(match["date"], match["result"])
        """
        e1 = quote(team1, safe="")
        e2 = quote(team2, safe="")
        return self._get(f"/api/teams/{e1}/head-to-head/{e2}")

    # ------------------------------------------------------------------
    # Matches
    # ------------------------------------------------------------------

    def get_matches(
        self,
        season: str | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        limit: int | None = None,
        offset: int | None = None,
    ) -> list[dict[str, Any]]:
        """Return a paginated list of historical matches.

        Parameters
        ----------
        season:
            Filter by season string, e.g. ``"2024/2025"``.
        from_date:
            ISO-8601 start date (``"YYYY-MM-DD"``).
        to_date:
            ISO-8601 end date (``"YYYY-MM-DD"``).
        limit:
            Maximum number of results (default 100, max 500).
        offset:
            Pagination offset (default 0).

        Returns
        -------
        list of dict
            Each dict contains match details including ELO values.

        Example
        -------
        ::

            matches = client.get_matches(season="2024/2025", limit=20)
            for m in matches:
                print(m["homeTeam"], "vs", m["awayTeam"], m["date"])
        """
        return self._get(
            "/api/matches",
            params={
                "season": season,
                "from": from_date,
                "to": to_date,
                "limit": limit,
                "offset": offset,
            },
        )

    def get_upcoming_matches(
        self,
        league: str | None = None,
        sort: str | None = None,
        limit: int | None = None,
    ) -> list[dict[str, Any]]:
        """Return upcoming scheduled matches with ELO difference predictions.

        Parameters
        ----------
        league:
            Filter by league code (e.g. ``"BL1"``).
        sort:
            Sort order (default ``"date"``).
        limit:
            Maximum number of results (default 50, max 200).

        Returns
        -------
        list of dict
            Each dict contains ``date``, ``homeTeam``, ``awayTeam``,
            ``homeElo``, ``awayElo``, ``eloDiff``, ``competition``,
            ``matchDay``.

        Example
        -------
        ::

            upcoming = client.get_upcoming_matches(league="BL1")
            for m in upcoming:
                print(m["homeTeam"], "vs", m["awayTeam"])
        """
        return self._get(
            "/api/matches/upcoming",
            params={"league": league, "sort": sort, "limit": limit},
        )

    def get_match(self, match_id: int | str) -> dict[str, Any]:
        """Return full details for a single match.

        Parameters
        ----------
        match_id:
            The numeric match identifier.

        Returns
        -------
        dict
            Contains ``match``, ``homeRecentForm``, ``awayRecentForm``,
            ``homeStats``, ``awayStats``, and ``headToHead``.

        Example
        -------
        ::

            match = client.get_match(12345)
            print(match["match"]["homeTeam"], match["match"]["homeElo"])
        """
        return self._get(f"/api/matches/{match_id}")

    # ------------------------------------------------------------------
    # Seasons
    # ------------------------------------------------------------------

    def get_seasons(self) -> dict[str, Any]:
        """Return all available seasons.

        Returns
        -------
        dict
            Contains ``seasons`` (list of season strings) and
            ``latestSeason`` (most recent season string).

        Example
        -------
        ::

            data = client.get_seasons()
            print(data["latestSeason"])
        """
        return self._get("/api/seasons")

    def get_season(
        self, season: str, league: str | None = None
    ) -> list[dict[str, Any]]:
        """Return per-team ELO change statistics for a specific season.

        Parameters
        ----------
        season:
            Season string, e.g. ``"2024/2025"``.
        league:
            Optional league filter, e.g. ``"BL1"``.

        Returns
        -------
        list of dict
            Each entry has ``teamName``, ``season``, ``startElo``,
            ``endElo``, ``change``, ``league``, ``totalMatches``.

        Example
        -------
        ::

            season_data = client.get_season("2024/2025", league="BL1")
            for entry in season_data:
                print(entry["teamName"], entry["change"])
        """
        encoded = quote(season, safe="")
        return self._get(f"/api/seasons/{encoded}", params={"league": league})

    # ------------------------------------------------------------------
    # Comparison
    # ------------------------------------------------------------------

    def get_comparison_history(self, teams: list[str]) -> dict[str, Any]:
        """Return historical ELO time-series for multiple teams side-by-side.

        Parameters
        ----------
        teams:
            List of team name strings to compare (at least two).

        Returns
        -------
        dict
            Keys are team names; values are lists of ``{"Date": epoch_ms, "ELO": number}``.

        Example
        -------
        ::

            history = client.get_comparison_history(
                ["Bayern München", "Borussia Dortmund"]
            )
            for team, datapoints in history.items():
                print(team, datapoints[-1]["ELO"])
        """
        if not isinstance(teams, list) or len(teams) < 2:
            raise ValueError("At least two team names are required for comparison")
        return self._get(
            "/api/comparison/history",
            params={"teams": ",".join(teams)},
        )
