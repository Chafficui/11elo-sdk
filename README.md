# 11elo API Clients

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official client libraries for the **[11elo](https://11elo.com) Soccer ELO API** — live and historical Elo ratings for German football (Bundesliga, 2. Bundesliga, 3. Liga).

## Available clients

| Language | Package | Install |
|----------|---------|---------|
| **JavaScript / Node.js** | [![npm](https://img.shields.io/npm/v/11elo)](https://www.npmjs.com/package/11elo) | `npm install 11elo` |
| **Python** | [![PyPI](https://img.shields.io/pypi/v/elevenelo)](https://pypi.org/project/elevenelo/) | `pip install elevenelo` |
| **Go** | [![Go Reference](https://pkg.go.dev/badge/github.com/Chafficui/11elo-clients/go.svg)](https://pkg.go.dev/github.com/Chafficui/11elo-clients/go) | `go get github.com/Chafficui/11elo-clients/go` |
| **PHP** | [![Packagist](https://img.shields.io/packagist/v/chafficui/elevenelo)](https://packagist.org/packages/chafficui/elevenelo) | `composer require chafficui/elevenelo` |

Each client lives in its own directory with a dedicated README, tests, and package manifest ready for publishing.

## Getting an API key

Register for free at [11elo.com/developer](https://11elo.com/developer).

Keys follow the format `11e_<tier>_<hex>` and are sent via the `X-API-Key` request header (handled automatically by every client).

**Rate limits by tier:**

| Tier | Requests / day |
|------|---------------|
| Free | 100 |
| Basic | 1,000 |
| Pro | 10,000 |

## API endpoints

All clients wrap the same public API:

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

## Publishing

### JavaScript (npm)

```bash
cd javascript
npm publish
```

### Python (PyPI)

```bash
cd python
python -m build
twine upload dist/*
```

### Go

Go modules are published automatically when you tag a release:

```bash
git tag go/v0.1.0
git push origin go/v0.1.0
```

### PHP (Packagist)

Register the package on [packagist.org](https://packagist.org/) pointing to this repository's `php/` directory. New versions are published via git tags.

## License

MIT
