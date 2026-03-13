# 11elo SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/Chafficui/11elo-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/Chafficui/11elo-sdk/actions/workflows/ci.yml)

Official SDKs for the **[11elo](https://11elo.com) Soccer ELO API** — live and historical Elo ratings for German football (Bundesliga, 2. Bundesliga, 3. Liga).

## Installation

| Language | Package | Install |
|----------|---------|---------|
| **JavaScript / Node.js** | [![npm](https://img.shields.io/npm/v/11elo)](https://www.npmjs.com/package/11elo) | `npm install 11elo` |
| **Python** | [![PyPI](https://img.shields.io/pypi/v/elevenelo)](https://pypi.org/project/elevenelo/) | `pip install elevenelo` |
| **Go** | [![Go Reference](https://pkg.go.dev/badge/github.com/Chafficui/11elo-sdk/go.svg)](https://pkg.go.dev/github.com/Chafficui/11elo-sdk/go) | `go get github.com/Chafficui/11elo-sdk/go` |
| **PHP** | [![Packagist](https://img.shields.io/packagist/v/chafficui/elevenelo)](https://packagist.org/packages/chafficui/elevenelo) | `composer require chafficui/elevenelo` |

Each SDK lives in its own directory with a dedicated README, tests, and examples.

## Quick start

### JavaScript

```js
import { Client } from '11elo';

const client = new Client({ apiKey: '11e_fre_your_key_here' });

const teams = await client.getTeams();
teams.forEach(t => console.log(t.teamName, t.currentElo));
```

### Python

```python
from elevenelo import Client

client = Client(api_key="11e_fre_your_key_here")

teams = client.get_teams()
for team in teams:
    print(team["teamName"], team["currentElo"])
```

### Go

```go
client, _ := elevenelo.NewClient("11e_fre_your_key_here")

teams, _ := client.GetTeams(context.Background())
for _, t := range teams {
    fmt.Println(t["teamName"], t["currentElo"])
}
```

### PHP

```php
$client = new \ElevenElo\Client('11e_fre_your_key_here');

$teams = $client->getTeams();
foreach ($teams as $team) {
    echo $team['teamName'] . ' ' . $team['currentElo'] . PHP_EOL;
}
```

## Getting an API key

Register for free at [11elo.com/docs](https://www.11elo.com/docs).

Keys follow the format `11e_<tier>_<hex>` and are sent via the `X-API-Key` request header (handled automatically by every SDK).

**Rate limits by tier:**

| Tier | Requests / day |
|------|---------------|
| Free | 100 |
| Basic | 1,000 |
| Pro | 10,000 |

## API endpoints

All SDKs wrap the same REST API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| Teams | `GET /api/teams` | All teams with current ELO stats |
| Team detail | `GET /api/teams/:name` | Full team info, ELO history, form |
| Head-to-head | `GET /api/teams/:a/head-to-head/:b` | H2H match history |
| Matches | `GET /api/matches` | Paginated match history |
| Upcoming | `GET /api/matches/upcoming` | Future fixtures with ELO predictions |
| Match detail | `GET /api/matches/:id` | Single match details |
| Seasons | `GET /api/seasons` | Available seasons |
| Season stats | `GET /api/seasons/:season` | Per-team ELO changes |
| Comparison | `GET /api/comparison/history` | Multi-team ELO time-series |

For detailed API reference and error handling, see the README in each SDK directory:
- [JavaScript SDK](./javascript/README.md)
- [Python SDK](./python/README.md)
- [Go SDK](./go/README.md)
- [PHP SDK](./php/README.md)

## Contributing

Contributions are welcome! To get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-change`)
3. Make your changes and add tests
4. Run the test suite for the SDK you modified (see below)
5. Open a pull request

### Running tests

```bash
# JavaScript
cd javascript && node --test src/client.test.js

# Python
cd python && pip install -e ".[dev]" && pytest

# Go
cd go && go test -v ./...

# PHP
cd php && composer install && vendor/bin/phpunit
```

### Updating SDKs for API changes

When the 11elo API adds or changes endpoints:

1. Implement the new method in each SDK's client (`javascript/src/index.js`, `python/elevenelo/client.py`, `go/elevenelo.go`, `php/src/Client.php`)
2. Add corresponding tests
3. Update the README in each affected SDK directory
4. Bump the version in the respective package manifest (`package.json`, `pyproject.toml`, `go.mod`, `composer.json`)

### Versioning

All SDKs follow [Semantic Versioning](https://semver.org/). Bump versions as follows:

- **Patch** (0.1.x): Bug fixes, documentation updates
- **Minor** (0.x.0): New API methods, backwards-compatible additions
- **Major** (x.0.0): Breaking changes to the SDK interface

## License

MIT
