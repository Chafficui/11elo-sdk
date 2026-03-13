# elevenelo

[![PyPI](https://img.shields.io/pypi/v/elevenelo)](https://pypi.org/project/elevenelo/)
[![Python](https://img.shields.io/pypi/pyversions/elevenelo)](https://pypi.org/project/elevenelo/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Python client for the **[11elo](https://11elo.com) Soccer ELO API** — live and historical Elo ratings for German football (Bundesliga, 2. Bundesliga, 3. Liga).

## Installation

```bash
pip install elevenelo
```

Requires Python 3.10+ and [`requests`](https://pypi.org/project/requests/).

## Quick start

```python
from elevenelo import Client

client = Client(api_key="11e_fre_your_key_here")

# List all teams
teams = client.get_teams()
for team in teams:
    print(team["teamName"], team["currentElo"])

# Get detailed info for one team
team = client.get_team("Bayern München")
print(team["team"]["currentElo"])

# Upcoming matches
upcoming = client.get_upcoming_matches(league="BL1", limit=10)
for match in upcoming:
    print(match["homeTeam"], "vs", match["awayTeam"])
```

## Getting an API key

Register for free at <https://11elo.com/developer>.  
Keys follow the format `11e_<tier>_<hex>` and are passed via the `X-API-Key` request header (handled automatically by this client).

**Rate limits by tier:**

| Tier  | Requests / day |
|-------|---------------|
| Free  | 100           |
| Basic | 1,000         |
| Pro   | 10,000        |

## API reference

### `Client(api_key, *, base_url, timeout, session)`

| Parameter  | Default                 | Description                                    |
|------------|-------------------------|------------------------------------------------|
| `api_key`  | —                       | Your 11elo API key (**required**)              |
| `base_url` | `https://11elo.com`     | Override for self-hosted / local dev           |
| `timeout`  | `30`                    | HTTP request timeout in seconds                |
| `session`  | `None` (auto-created)   | Custom `requests.Session`                      |

---

### Teams

#### `client.get_teams() → list[dict]`

Returns all teams with current ELO stats, league info and trend data.

```python
teams = client.get_teams()
# [{"teamName": "Bayern München", "currentElo": 1847, "league": "BL1", ...}, ...]
```

#### `client.get_team(team_name) → dict`

Full detail for one team — ELO history, recent form, upcoming matches, career stats.

```python
team = client.get_team("Borussia Dortmund")
# {
#   "team": {"name": "Borussia Dortmund", "currentElo": 1720, ...},
#   "eloHistory": [...],
#   "recentForm": [...],
#   "upcomingMatches": [...]
# }
```

#### `client.get_head_to_head(team1, team2) → list[dict]`

Historical head-to-head match results between two teams.

```python
h2h = client.get_head_to_head("Bayern München", "Borussia Dortmund")
# [{"date": "2026-02-15", "result": "2:1", "winner": "Bayern München", ...}, ...]
```

---

### Matches

#### `client.get_matches(*, season, from_date, to_date, limit, offset) → list[dict]`

Paginated match history.  All parameters are optional.

```python
matches = client.get_matches(season="2024/2025", limit=50)
# [{"homeTeam": "Bayern München", "awayTeam": "BVB", "homeElo": 1835, ...}, ...]
```

#### `client.get_upcoming_matches(*, league, sort, limit) → list[dict]`

Upcoming fixtures with ELO-difference predictions.

```python
upcoming = client.get_upcoming_matches(league="BL1", limit=20)
# [{"homeTeam": "...", "awayTeam": "...", "eloDiff": 40, ...}, ...]
```

#### `client.get_match(match_id) → dict`

Full detail for a single match: teams, ELO at time of match, head-to-head, recent form.

```python
match = client.get_match(12345)
# {"match": {...}, "homeRecentForm": [...], "awayRecentForm": [...], ...}
```

---

### Seasons

#### `client.get_seasons() → dict`

List all available seasons and the latest one.

```python
data = client.get_seasons()
# {"seasons": ["2025/2026", "2024/2025", ...], "latestSeason": "2025/2026"}
```

#### `client.get_season(season, *, league) → list[dict]`

Per-team ELO change statistics for a given season.

```python
entries = client.get_season("2024/2025", league="BL1")
# [{"teamName": "Bayern München", "startElo": 1820, "endElo": 1847, "change": 27, ...}, ...]
```

---

### Comparison

#### `client.get_comparison_history(teams) → dict`

Time-series ELO data for multiple teams in one call.

```python
history = client.get_comparison_history(["Bayern München", "Borussia Dortmund"])
# {
#   "Bayern München": [{"Date": 1709856000000, "ELO": 1847}, ...],
#   "Borussia Dortmund": [{"Date": 1709856000000, "ELO": 1720}, ...]
# }
```

---

## Error handling

All exceptions inherit from `elevenelo.ElevenEloError`.

| Exception              | When raised                                      |
|------------------------|--------------------------------------------------|
| `AuthenticationError`  | API key is missing or invalid (HTTP 401)         |
| `RateLimitError`       | Daily quota exceeded (HTTP 429)                  |
| `NotFoundError`        | Resource does not exist (HTTP 404)               |
| `ApiError`             | Any other non-2xx response                       |

```python
from elevenelo import Client, AuthenticationError, RateLimitError, NotFoundError

client = Client(api_key="11e_fre_your_key_here")

try:
    team = client.get_team("Unknown FC")
except NotFoundError:
    print("Team not found")
except RateLimitError as e:
    print(f"Rate limit hit, resets at {e.reset_at}")
except AuthenticationError:
    print("Bad API key")
```

## License

MIT
