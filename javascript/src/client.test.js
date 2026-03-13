/**
 * Tests for the 11elo JavaScript client.
 * Uses the built-in Node.js test runner (node:test) – no extra dependencies.
 *
 * Run: node --test src/client.test.js
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';

import {
  Client,
  ElevenEloError,
  AuthenticationError,
  RateLimitError,
  NotFoundError,
  ApiError,
} from './index.js';

const BASE = 'https://11elo.com';

// ---------------------------------------------------------------------------
// Minimal fetch mock helpers
// ---------------------------------------------------------------------------

function mockFetch(status, body, headers = {}) {
  return async () =>
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json', ...headers },
    });
}

function makeMockClient(fetchFn) {
  const client = new Client({ apiKey: '11e_fre_testkey' });
  // Replace global fetch for the duration of the test
  client._fetch = fetchFn;
  // Patch internal _get to use _fetch
  const originalGet = client._get.bind(client);
  client._get = async (path, params = {}) => {
    const url = client._url(path) + client._buildQuery(params);
    const controller = new AbortController();
    const timerId = setTimeout(() => controller.abort(), client._timeoutMs);
    let response;
    try {
      response = await client._fetch(url, {
        method: 'GET',
        headers: { 'X-API-Key': client._apiKey, Accept: 'application/json' },
        signal: controller.signal,
      });
    } finally {
      clearTimeout(timerId);
    }
    return client._handleResponse(response);
  };
  return client;
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

test('throws when apiKey is missing', () => {
  assert.throws(() => new Client({}), /apiKey is required/);
});

test('accepts custom baseUrl and timeoutMs', () => {
  const c = new Client({ apiKey: 'k', baseUrl: 'http://localhost:3001', timeoutMs: 5000 });
  assert.equal(c._baseUrl, 'http://localhost:3001');
  assert.equal(c._timeoutMs, 5000);
});

// ---------------------------------------------------------------------------
// Successful responses
// ---------------------------------------------------------------------------

test('getTeams returns parsed JSON', async () => {
  const payload = [{ teamName: 'Bayern München', currentElo: 1847 }];
  const client = makeMockClient(mockFetch(200, payload));
  const result = await client.getTeams();
  assert.deepEqual(result, payload);
});

test('getTeam encodes team name in URL', async () => {
  const payload = { team: { name: 'Bayern München' } };
  let capturedUrl;
  const client = makeMockClient(async (url, opts) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getTeam('Bayern München');
  assert.ok(capturedUrl.includes('Bayern%20M%C3%BCnchen'));
});

test('getHeadToHead encodes both team names', async () => {
  const payload = [{ result: '2:1' }];
  let capturedUrl;
  const client = makeMockClient(async (url) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getHeadToHead('Bayern München', 'Borussia Dortmund');
  assert.ok(capturedUrl.includes('head-to-head'));
  assert.ok(capturedUrl.includes('Bayern%20M%C3%BCnchen'));
  assert.ok(capturedUrl.includes('Borussia%20Dortmund'));
});

test('getMatches passes query params', async () => {
  const payload = [{ id: 1 }];
  let capturedUrl;
  const client = makeMockClient(async (url) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getMatches({ season: '2024/2025', limit: 10, offset: 0 });
  assert.ok(capturedUrl.includes('limit=10'));
  assert.ok(capturedUrl.includes('offset=0'));
  assert.ok(capturedUrl.includes('season='));
});

test('getUpcomingMatches passes league param', async () => {
  const payload = [{ id: 2 }];
  let capturedUrl;
  const client = makeMockClient(async (url) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getUpcomingMatches({ league: 'BL1' });
  assert.ok(capturedUrl.includes('league=BL1'));
});

test('getMatch appends matchId to path', async () => {
  const payload = { match: { id: 12345 } };
  let capturedUrl;
  const client = makeMockClient(async (url) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getMatch(12345);
  assert.ok(capturedUrl.endsWith('/api/matches/12345'));
});

test('getSeasons returns seasons object', async () => {
  const payload = { seasons: ['2025/2026'], latestSeason: '2025/2026' };
  const client = makeMockClient(mockFetch(200, payload));
  const result = await client.getSeasons();
  assert.deepEqual(result, payload);
});

test('getSeason encodes season in URL', async () => {
  const payload = [{ teamName: 'Bayern München', change: 27 }];
  let capturedUrl;
  const client = makeMockClient(async (url) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getSeason('2024/2025', { league: 'BL1' });
  assert.ok(capturedUrl.includes('2024'));
  assert.ok(capturedUrl.includes('league=BL1'));
});

test('getComparisonHistory joins teams with comma', async () => {
  const payload = { 'Bayern München': [], 'Borussia Dortmund': [] };
  let capturedUrl;
  const client = makeMockClient(async (url) => {
    capturedUrl = url;
    return new Response(JSON.stringify(payload), { status: 200 });
  });
  await client.getComparisonHistory(['Bayern München', 'Borussia Dortmund']);
  assert.ok(capturedUrl.includes('teams='));
});

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

test('getTeams throws AuthenticationError on 401', async () => {
  const client = makeMockClient(mockFetch(401, { error: 'Unauthorized' }));
  await assert.rejects(() => client.getTeams(), AuthenticationError);
});

test('getTeams throws RateLimitError on 429 with resetAt', async () => {
  const client = makeMockClient(
    mockFetch(429, {}, { 'X-RateLimit-Reset': '2026-03-13T00:00:00Z' }),
  );
  try {
    await client.getTeams();
    assert.fail('expected RateLimitError');
  } catch (err) {
    assert.ok(err instanceof RateLimitError);
    assert.equal(err.resetAt, '2026-03-13T00:00:00Z');
  }
});

test('getTeam throws NotFoundError on 404', async () => {
  const client = makeMockClient(mockFetch(404, {}));
  await assert.rejects(() => client.getTeam('Unknown FC'), NotFoundError);
});

test('getTeams throws ApiError on 500', async () => {
  const client = makeMockClient(mockFetch(500, {}));
  try {
    await client.getTeams();
    assert.fail('expected ApiError');
  } catch (err) {
    assert.ok(err instanceof ApiError);
    assert.equal(err.statusCode, 500);
  }
});

test('getComparisonHistory throws when fewer than two teams given', () => {
  const client = new Client({ apiKey: 'k' });
  assert.throws(() => client.getComparisonHistory(['Bayern München']), /at least two/);
});
