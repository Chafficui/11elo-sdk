# elevenelo (Go)

[![Go Reference](https://pkg.go.dev/badge/github.com/Chafficui/11elo-clients/go.svg)](https://pkg.go.dev/github.com/Chafficui/11elo-clients/go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go client for the **[11elo](https://11elo.com) Soccer ELO API** — live and historical Elo ratings for German football (Bundesliga, 2. Bundesliga, 3. Liga).

No external dependencies — uses the Go standard library only (`net/http`).

## Installation

```bash
go get github.com/Chafficui/11elo-clients/go@latest
```

Requires **Go 1.21+**.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    elevenelo "github.com/Chafficui/11elo-clients/go"
)

func main() {
    client, err := elevenelo.NewClient("11e_fre_your_key_here")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List all teams
    teams, err := client.GetTeams(ctx)
    if err != nil {
        log.Fatal(err)
    }
    for _, t := range teams {
        fmt.Println(t["teamName"], t["currentElo"])
    }

    // Get detailed info for one team
    team, err := client.GetTeam(ctx, "Bayern München")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(team["team"])

    // Upcoming matches
    upcoming, err := client.GetUpcomingMatches(ctx, &elevenelo.UpcomingMatchesOptions{
        League: "BL1",
        Limit:  10,
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, m := range upcoming {
        fmt.Println(m["homeTeam"], "vs", m["awayTeam"])
    }
}
```

## Getting an API key

Register for free at <https://11elo.com/developer>.  
Keys follow the format `11e_<tier>_<hex>` and are sent via the `X-API-Key` request header (handled automatically by this client).

**Rate limits by tier:**

| Tier  | Requests / day |
|-------|---------------|
| Free  | 100           |
| Basic | 1,000         |
| Pro   | 10,000        |

## API reference

### `NewClient(apiKey string, opts ...Option) (*Client, error)`

| Option | Default | Description |
|--------|---------|-------------|
| `WithBaseURL(url)` | `https://11elo.com` | Override for self-hosted / local dev |
| `WithTimeout(d)` | `30s` | HTTP request timeout |
| `WithHTTPClient(c)` | auto-created | Custom `*http.Client` |

---

### Teams

#### `GetTeams(ctx) ([]map[string]any, error)`

Returns all teams with current ELO stats, league info and trend data.

```go
teams, err := client.GetTeams(ctx)
// [{"teamName": "Bayern München", "currentElo": 1847, "league": "BL1", ...}, ...]
```

#### `GetTeam(ctx, teamName) (map[string]any, error)`

Full detail for one team — ELO history, recent form, upcoming matches, career stats.

```go
team, err := client.GetTeam(ctx, "Borussia Dortmund")
// {"team": {...}, "eloHistory": [...], "recentForm": [...], "upcomingMatches": [...]}
```

#### `GetHeadToHead(ctx, team1, team2) ([]map[string]any, error)`

Historical head-to-head match results between two teams.

```go
h2h, err := client.GetHeadToHead(ctx, "Bayern München", "Borussia Dortmund")
// [{"date": "2026-02-15", "result": "2:1", "winner": "Bayern München", ...}, ...]
```

---

### Matches

#### `GetMatches(ctx, *MatchesOptions) ([]map[string]any, error)`

Paginated match history.  All options are optional — pass `nil` for defaults.

```go
matches, err := client.GetMatches(ctx, &elevenelo.MatchesOptions{
    Season: "2024/2025",
    Limit:  50,
})
```

`MatchesOptions` fields:

| Field      | Type   | Description                                      |
|------------|--------|--------------------------------------------------|
| `Season`   | string | Filter by season, e.g. `"2024/2025"`             |
| `FromDate` | string | ISO-8601 start date `"YYYY-MM-DD"`               |
| `ToDate`   | string | ISO-8601 end date `"YYYY-MM-DD"`                 |
| `Limit`    | int    | Max results (default 100, max 500); `0` = default|
| `Offset`   | int    | Pagination offset; `0` = default                 |

#### `GetUpcomingMatches(ctx, *UpcomingMatchesOptions) ([]map[string]any, error)`

Upcoming fixtures with ELO-difference predictions.

```go
upcoming, err := client.GetUpcomingMatches(ctx, &elevenelo.UpcomingMatchesOptions{
    League: "BL1",
})
```

#### `GetMatch(ctx, matchID int64) (map[string]any, error)`

Full detail for a single match.

```go
match, err := client.GetMatch(ctx, 12345)
// {"match": {...}, "homeRecentForm": [...], "awayRecentForm": [...], ...}
```

---

### Seasons

#### `GetSeasons(ctx) (*SeasonsResponse, error)`

List all available seasons and the most recent one.

```go
data, err := client.GetSeasons(ctx)
fmt.Println(data.LatestSeason) // "2025/2026"
```

#### `GetSeason(ctx, season, league string) ([]map[string]any, error)`

Per-team ELO change statistics for a given season.  Pass `""` for `league` to retrieve all leagues.

```go
entries, err := client.GetSeason(ctx, "2024/2025", "BL1")
// [{"teamName": "Bayern München", "startElo": 1820, "endElo": 1847, "change": 27, ...}, ...]
```

---

### Comparison

#### `GetComparisonHistory(ctx, teams []string) (map[string]any, error)`

Time-series ELO data for multiple teams in one call.  `teams` must contain at least two names.

```go
history, err := client.GetComparisonHistory(ctx, []string{
    "Bayern München",
    "Borussia Dortmund",
})
// map[string]any{
//   "Bayern München":    []any{{"Date": 1709856000000, "ELO": 1847}, ...},
//   "Borussia Dortmund": []any{{"Date": 1709856000000, "ELO": 1720}, ...},
// }
```

---

## Error handling

All errors are returned as Go `error` values.  You can type-assert to one of the specific error types:

| Type                   | When returned                                      |
|------------------------|----------------------------------------------------|
| `*AuthenticationError` | API key is missing or invalid (HTTP 401)           |
| `*RateLimitError`      | Daily quota exceeded (HTTP 429)                    |
| `*NotFoundError`       | Resource does not exist (HTTP 404)                 |
| `*APIError`            | Any other non-2xx response; `.StatusCode` is set   |

```go
import (
    "errors"
    elevenelo "github.com/Chafficui/11elo-clients/go"
)

team, err := client.GetTeam(ctx, "Unknown FC")
var notFound *elevenelo.NotFoundError
var rateLimit *elevenelo.RateLimitError
var authErr *elevenelo.AuthenticationError

switch {
case errors.As(err, &notFound):
    fmt.Println("Team not found")
case errors.As(err, &rateLimit):
    fmt.Printf("Rate limit hit, resets at %s\n", rateLimit.ResetAt)
case errors.As(err, &authErr):
    fmt.Println("Bad API key")
case err != nil:
    fmt.Println("Other error:", err)
}
```

## License

MIT
