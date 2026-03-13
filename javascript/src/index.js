/**
 * JavaScript/Node.js client for the 11elo Soccer ELO API.
 *
 * Requires Node.js 18+ (uses native `fetch`) or any environment that
 * exposes the global `fetch` function (modern browsers, Deno, Bun, etc.).
 *
 * @example
 * ```js
 * import { Client } from '11elo';
 *
 * const client = new Client({ apiKey: '11e_fre_your_key_here' });
 *
 * const teams = await client.getTeams();
 * console.log(teams[0].teamName, teams[0].currentElo);
 * ```
 * @module
 */

const DEFAULT_BASE_URL = 'https://api.11elo.com';
const DEFAULT_TIMEOUT_MS = 30_000;

// ---------------------------------------------------------------------------
// Custom errors
// ---------------------------------------------------------------------------

/**
 * Base error class for all 11elo client errors.
 */
export class ElevenEloError extends Error {
  constructor(message) {
    super(message);
    this.name = 'ElevenEloError';
  }
}

/**
 * Thrown when the API key is missing or invalid (HTTP 401).
 */
export class AuthenticationError extends ElevenEloError {
  constructor(message = 'Invalid or missing API key. Obtain one at https://www.11elo.com/docs') {
    super(message);
    this.name = 'AuthenticationError';
  }
}

/**
 * Thrown when the daily rate limit for the API key tier has been exceeded (HTTP 429).
 * @property {string|null} resetAt - ISO-8601 timestamp when the limit resets.
 */
export class RateLimitError extends ElevenEloError {
  constructor(message = 'Daily rate limit exceeded. Upgrade your plan at https://www.11elo.com/docs', resetAt = null) {
    super(message);
    this.name = 'RateLimitError';
    this.resetAt = resetAt;
  }
}

/**
 * Thrown when the requested resource does not exist (HTTP 404).
 */
export class NotFoundError extends ElevenEloError {
  constructor(message) {
    super(message);
    this.name = 'NotFoundError';
  }
}

/**
 * Thrown for unexpected API errors (any non-2xx response not covered above).
 * @property {number} statusCode - The HTTP status code.
 */
export class ApiError extends ElevenEloError {
  constructor(message, statusCode) {
    super(message);
    this.name = 'ApiError';
    this.statusCode = statusCode;
  }
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

/**
 * Synchronous-style async client for the 11elo public API.
 *
 * @example
 * ```js
 * const client = new Client({ apiKey: '11e_fre_your_key_here' });
 * const team = await client.getTeam('Bayern München');
 * console.log(team.team.currentElo);
 * ```
 */
export class Client {
  /**
   * @param {object} options
   * @param {string} options.apiKey - Your 11elo API key (required).
   * @param {string} [options.baseUrl='https://api.11elo.com'] - Override the base URL.
   * @param {number} [options.timeoutMs=30000] - Request timeout in milliseconds.
   */
  constructor({ apiKey, baseUrl = DEFAULT_BASE_URL, timeoutMs = DEFAULT_TIMEOUT_MS } = {}) {
    if (!apiKey) {
      throw new Error('apiKey is required');
    }
    this._apiKey = apiKey;
    this._baseUrl = baseUrl.replace(/\/$/, '');
    this._timeoutMs = timeoutMs;
  }

  // -------------------------------------------------------------------------
  // Internal helpers
  // -------------------------------------------------------------------------

  _url(path) {
    return `${this._baseUrl}${path}`;
  }

  _buildQuery(params) {
    const filtered = Object.fromEntries(
      Object.entries(params).filter(([, v]) => v !== undefined && v !== null),
    );
    const qs = new URLSearchParams(filtered).toString();
    return qs ? `?${qs}` : '';
  }

  async _get(path, params = {}) {
    const url = this._url(path) + this._buildQuery(params);
    const controller = new AbortController();
    const timerId = setTimeout(() => controller.abort(), this._timeoutMs);

    let response;
    try {
      response = await fetch(url, {
        method: 'GET',
        headers: {
          'X-API-Key': this._apiKey,
          Accept: 'application/json',
        },
        signal: controller.signal,
      });
    } catch (err) {
      if (err.name === 'AbortError') {
        throw new ElevenEloError(`Request timed out after ${this._timeoutMs}ms`);
      }
      throw new ElevenEloError(`Network error: ${err.message}`);
    } finally {
      clearTimeout(timerId);
    }

    return this._handleResponse(response);
  }

  async _handleResponse(response) {
    if (response.status === 401) {
      throw new AuthenticationError();
    }
    if (response.status === 429) {
      const resetAt = response.headers.get('X-RateLimit-Reset');
      throw new RateLimitError(undefined, resetAt);
    }
    if (response.status === 404) {
      throw new NotFoundError(`Resource not found: ${response.url}`);
    }
    if (!response.ok) {
      const body = await response.text().catch(() => '');
      throw new ApiError(
        `API request failed with status ${response.status}: ${body}`,
        response.status,
      );
    }

    try {
      return await response.json();
    } catch {
      const text = await response.text().catch(() => '');
      throw new ElevenEloError(`Failed to parse JSON response: ${text}`);
    }
  }

  // -------------------------------------------------------------------------
  // Teams
  // -------------------------------------------------------------------------

  /**
   * Return all teams with their current ELO stats and league info.
   *
   * @returns {Promise<Array<object>>} Array of team objects.
   *
   * @example
   * ```js
   * const teams = await client.getTeams();
   * teams.forEach(t => console.log(t.teamName, t.currentElo));
   * ```
   */
  getTeams() {
    return this._get('/api/teams');
  }

  /**
   * Return detailed information for a single team.
   *
   * @param {string} teamName - The canonical team name, e.g. `"Bayern München"`.
   * @returns {Promise<object>} Object with `team`, `eloHistory`, `recentForm`,
   *   `significantMatches`, `stats`, and `upcomingMatches`.
   *
   * @example
   * ```js
   * const { team, eloHistory } = await client.getTeam('Bayern München');
   * console.log(team.currentElo);
   * ```
   */
  getTeam(teamName) {
    return this._get(`/api/teams/${encodeURIComponent(teamName)}`);
  }

  /**
   * Return head-to-head match history between two teams.
   *
   * @param {string} team1 - Name of the first team.
   * @param {string} team2 - Name of the second team.
   * @returns {Promise<Array<object>>} Array of match result objects.
   *
   * @example
   * ```js
   * const h2h = await client.getHeadToHead('Bayern München', 'Borussia Dortmund');
   * console.log(h2h[0].result);
   * ```
   */
  getHeadToHead(team1, team2) {
    return this._get(
      `/api/teams/${encodeURIComponent(team1)}/head-to-head/${encodeURIComponent(team2)}`,
    );
  }

  // -------------------------------------------------------------------------
  // Matches
  // -------------------------------------------------------------------------

  /**
   * Return a paginated list of historical matches.
   *
   * @param {object} [options]
   * @param {string} [options.season] - Season string, e.g. `"2024/2025"`.
   * @param {string} [options.from] - ISO-8601 start date (`"YYYY-MM-DD"`).
   * @param {string} [options.to] - ISO-8601 end date (`"YYYY-MM-DD"`).
   * @param {number} [options.limit] - Max results (default 100, max 500).
   * @param {number} [options.offset] - Pagination offset (default 0).
   * @returns {Promise<Array<object>>} Array of match objects.
   *
   * @example
   * ```js
   * const matches = await client.getMatches({ season: '2024/2025', limit: 20 });
   * ```
   */
  getMatches({ season, from, to, limit, offset } = {}) {
    return this._get('/api/matches', { season, from, to, limit, offset });
  }

  /**
   * Return upcoming fixtures with ELO-difference predictions.
   *
   * @param {object} [options]
   * @param {string} [options.league] - League code filter, e.g. `"BL1"`.
   * @param {string} [options.sort] - Sort order (default `"date"`).
   * @param {number} [options.limit] - Max results (default 50, max 200).
   * @returns {Promise<Array<object>>}
   *
   * @example
   * ```js
   * const upcoming = await client.getUpcomingMatches({ league: 'BL1', limit: 10 });
   * ```
   */
  getUpcomingMatches({ league, sort, limit } = {}) {
    return this._get('/api/matches/upcoming', { league, sort, limit });
  }

  /**
   * Return full details for a single match.
   *
   * @param {number|string} matchId - The numeric match identifier.
   * @returns {Promise<object>} Object with `match`, `homeRecentForm`, `awayRecentForm`,
   *   `homeStats`, `awayStats`, and `headToHead`.
   *
   * @example
   * ```js
   * const { match, headToHead } = await client.getMatch(12345);
   * console.log(match.homeTeam, match.homeElo);
   * ```
   */
  getMatch(matchId) {
    return this._get(`/api/matches/${matchId}`);
  }

  // -------------------------------------------------------------------------
  // Seasons
  // -------------------------------------------------------------------------

  /**
   * Return all available seasons and the latest one.
   *
   * @returns {Promise<{seasons: string[], latestSeason: string}>}
   *
   * @example
   * ```js
   * const { seasons, latestSeason } = await client.getSeasons();
   * console.log(latestSeason);
   * ```
   */
  getSeasons() {
    return this._get('/api/seasons');
  }

  /**
   * Return per-team ELO change statistics for a specific season.
   *
   * @param {string} season - Season string, e.g. `"2024/2025"`.
   * @param {object} [options]
   * @param {string} [options.league] - Optional league filter, e.g. `"BL1"`.
   * @returns {Promise<Array<object>>}
   *
   * @example
   * ```js
   * const data = await client.getSeason('2024/2025', { league: 'BL1' });
   * data.forEach(e => console.log(e.teamName, e.change));
   * ```
   */
  getSeason(season, { league } = {}) {
    return this._get(`/api/seasons/${encodeURIComponent(season)}`, { league });
  }

  // -------------------------------------------------------------------------
  // Comparison
  // -------------------------------------------------------------------------

  /**
   * Return historical ELO time-series for multiple teams side-by-side.
   *
   * @param {string[]} teams - Array of team names (minimum two).
   * @returns {Promise<object>} Keys are team names; values are arrays of
   *   `{Date: number, ELO: number}` data points.
   *
   * @example
   * ```js
   * const history = await client.getComparisonHistory([
   *   'Bayern München',
   *   'Borussia Dortmund',
   * ]);
   * Object.entries(history).forEach(([team, points]) => {
   *   console.log(team, points.at(-1).ELO);
   * });
   * ```
   */
  getComparisonHistory(teams) {
    if (!Array.isArray(teams) || teams.length < 2) {
      throw new Error('getComparisonHistory requires an array of at least two team names');
    }
    return this._get('/api/comparison/history', { teams: teams.join(',') });
  }
}
