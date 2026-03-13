# 11elo

[![npm](https://img.shields.io/npm/v/11elo)](https://www.npmjs.com/package/11elo)
[![Node.js](https://img.shields.io/node/v/11elo)](https://www.npmjs.com/package/11elo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

JavaScript/Node.js client for the **[11elo](https://11elo.com) Soccer ELO API** — live and historical Elo ratings for German football (Bundesliga, 2. Bundesliga, 3. Liga).

Works in **Node.js 18+**, modern browsers, Deno and Bun — uses the native `fetch` API with no extra dependencies.

## Installation

```bash
npm install 11elo
```

## Quick start

```js
import { Client } from '11elo';

const client = new Client({ apiKey: '11e_fre_your_key_here' });

// List all teams
const teams = await client.getTeams();
teams.forEach(t => console.log(t.teamName, t.currentElo));

// Get detailed info for one team
const { team, eloHistory } = await client.getTeam('Bayern München');
console.log(team.currentElo);

// Upcoming matches
const upcoming = await client.getUpcomingMatches({ league: 'BL1', limit: 10 });
upcoming.forEach(m => console.log(m.homeTeam, 'vs', m.awayTeam));
```

## Getting an API key

Register for free at <https://www.11elo.com/docs>.  
Keys follow the format `11e_<tier>_<hex>` and are sent via the `X-API-Key` request header (handled automatically by this client).

**Rate limits by tier:**

| Tier  | Requests / day |
|-------|---------------|
| Free  | 100           |
| Basic | 1,000         |
| Pro   | 10,000        |

## API reference

### `new Client(options)`

| Option      | Default               | Description                                    |
|-------------|-----------------------|------------------------------------------------|
| `apiKey`    | —                     | Your 11elo API key (**required**)              |
| `baseUrl`   | `https://api.11elo.com`   | Override for self-hosted / local dev           |
| `timeoutMs` | `30000`               | Request timeout in milliseconds                |

---

### Teams

#### `client.getTeams() → Promise<object[]>`

Returns all teams with current ELO stats, league info and trend data.

```js
const teams = await client.getTeams();
// [{ teamName: 'Bayern München', currentElo: 1847, league: 'BL1', ... }, ...]
```

#### `client.getTeam(teamName) → Promise<object>`

Full detail for one team — ELO history, recent form, upcoming matches, career stats.

```js
const { team, eloHistory, recentForm, upcomingMatches } =
  await client.getTeam('Borussia Dortmund');
```

#### `client.getHeadToHead(team1, team2) → Promise<object[]>`

Historical head-to-head match results between two teams.

```js
const h2h = await client.getHeadToHead('Bayern München', 'Borussia Dortmund');
// [{ date: '2026-02-15', result: '2:1', winner: 'Bayern München', ... }, ...]
```

---

### Matches

#### `client.getMatches(options?) → Promise<object[]>`

Paginated match history.  All options are optional.

| Option   | Description                             |
|----------|-----------------------------------------|
| `season` | Filter by season, e.g. `"2024/2025"`   |
| `from`   | ISO-8601 start date (`"YYYY-MM-DD"`)   |
| `to`     | ISO-8601 end date (`"YYYY-MM-DD"`)     |
| `limit`  | Max results (default 100, max 500)      |
| `offset` | Pagination offset (default 0)           |

```js
const matches = await client.getMatches({ season: '2024/2025', limit: 50 });
```

#### `client.getUpcomingMatches(options?) → Promise<object[]>`

Upcoming fixtures with ELO-difference predictions.

| Option   | Description                              |
|----------|------------------------------------------|
| `league` | League code filter, e.g. `"BL1"`        |
| `sort`   | Sort order (default `"date"`)            |
| `limit`  | Max results (default 50, max 200)        |

```js
const upcoming = await client.getUpcomingMatches({ league: 'BL1' });
```

#### `client.getMatch(matchId) → Promise<object>`

Full detail for a single match.

```js
const { match, homeRecentForm, headToHead } = await client.getMatch(12345);
```

---

### Seasons

#### `client.getSeasons() → Promise<{ seasons: string[], latestSeason: string }>`

List all available seasons.

```js
const { seasons, latestSeason } = await client.getSeasons();
```

#### `client.getSeason(season, options?) → Promise<object[]>`

Per-team ELO change statistics for a given season.

| Option   | Description                             |
|----------|-----------------------------------------|
| `league` | Optional league filter, e.g. `"BL1"`   |

```js
const entries = await client.getSeason('2024/2025', { league: 'BL1' });
// [{ teamName: 'Bayern München', startElo: 1820, endElo: 1847, change: 27, ... }, ...]
```

---

### Comparison

#### `client.getComparisonHistory(teams) → Promise<object>`

Time-series ELO data for multiple teams in one call.

```js
const history = await client.getComparisonHistory([
  'Bayern München',
  'Borussia Dortmund',
]);
// {
//   'Bayern München': [{ Date: 1709856000000, ELO: 1847 }, ...],
//   'Borussia Dortmund': [{ Date: 1709856000000, ELO: 1720 }, ...]
// }
```

---

## Error handling

All exceptions extend `ElevenEloError`.

| Class                | When thrown                                      |
|----------------------|--------------------------------------------------|
| `AuthenticationError`| API key is missing or invalid (HTTP 401)         |
| `RateLimitError`     | Daily quota exceeded (HTTP 429)                  |
| `NotFoundError`      | Resource does not exist (HTTP 404)               |
| `ApiError`           | Any other non-2xx response                       |

```js
import { Client, AuthenticationError, RateLimitError, NotFoundError } from '11elo';

const client = new Client({ apiKey: '11e_fre_your_key_here' });

try {
  const team = await client.getTeam('Unknown FC');
} catch (err) {
  if (err instanceof NotFoundError) {
    console.error('Team not found');
  } else if (err instanceof RateLimitError) {
    console.error('Rate limit hit, resets at', err.resetAt);
  } else if (err instanceof AuthenticationError) {
    console.error('Bad API key');
  } else {
    throw err;
  }
}
```

## CommonJS usage

The package ships as ES modules (`"type": "module"`).  For CommonJS projects, use dynamic `import()`:

```js
const { Client } = await import('11elo');
```

Or use a bundler (webpack, Rollup, esbuild) which handles the interop automatically.

## License

MIT
