// Package elevenelo provides a Go client for the 11elo Soccer ELO API.
//
// Basic usage:
//
//	client, err := elevenelo.NewClient("11e_fre_your_key_here")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	teams, err := client.GetTeams(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, t := range teams {
//	    fmt.Println(t["teamName"], t["currentElo"])
//	}
package elevenelo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.11elo.com"
	defaultTimeout = 30 * time.Second
)

// Client is a synchronous HTTP client for the 11elo public API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
// Useful for self-hosted instances or local development.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithTimeout sets the HTTP request timeout (default: 30s).
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.http.Timeout = timeout
	}
}

// WithHTTPClient replaces the underlying *http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.http = httpClient
	}
}

// NewClient creates a new Client with the given API key and optional options.
// Returns an error if apiKey is empty.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("elevenelo: apiKey must not be empty")
	}
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (c *Client) get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	rawURL := c.baseURL + path
	if len(params) > 0 {
		rawURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("elevenelo: build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("elevenelo: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenelo: read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, &AuthenticationError{
			Message: "invalid or missing API key – obtain one at https://www.11elo.com/docs",
		}
	case http.StatusTooManyRequests:
		resetAt := resp.Header.Get("X-RateLimit-Reset")
		return nil, &RateLimitError{
			Message: "daily rate limit exceeded – upgrade your plan at https://www.11elo.com/docs",
			ResetAt: resetAt,
		}
	case http.StatusNotFound:
		return nil, &NotFoundError{
			Message: fmt.Sprintf("resource not found: %s", rawURL),
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			Message:    fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return json.RawMessage(body), nil
}

// decode unmarshals the JSON body into dst.
func decode(data json.RawMessage, dst any) error {
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("elevenelo: decode JSON: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Teams
// ---------------------------------------------------------------------------

// GetTeams returns all teams with their current ELO stats and league info.
// Each map contains keys such as "teamName", "currentElo", "league", etc.
func (c *Client) GetTeams(ctx context.Context) ([]map[string]any, error) {
	data, err := c.get(ctx, "/api/teams", nil)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	return result, decode(data, &result)
}

// GetTeam returns detailed information for a single team.
// teamName is the canonical team name, e.g. "Bayern München".
// The result contains "team", "eloHistory", "recentForm", "significantMatches",
// "stats", and "upcomingMatches".
func (c *Client) GetTeam(ctx context.Context, teamName string) (map[string]any, error) {
	path := "/api/teams/" + url.PathEscape(teamName)
	data, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	return result, decode(data, &result)
}

// GetHeadToHead returns the head-to-head match history between two teams.
// Each entry contains "date", "result", "winner", "homeGoals", "awayGoals".
func (c *Client) GetHeadToHead(ctx context.Context, team1, team2 string) ([]map[string]any, error) {
	path := "/api/teams/" + url.PathEscape(team1) + "/head-to-head/" + url.PathEscape(team2)
	data, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	return result, decode(data, &result)
}

// ---------------------------------------------------------------------------
// Matches
// ---------------------------------------------------------------------------

// MatchesOptions holds the optional query parameters for GetMatches.
type MatchesOptions struct {
	Season   string // e.g. "2024/2025"
	FromDate string // ISO-8601 start date "YYYY-MM-DD"
	ToDate   string // ISO-8601 end date "YYYY-MM-DD"
	Limit    int    // max results (default 100, max 500); 0 = use server default
	Offset   int    // pagination offset; -1 = use server default (0)
}

// GetMatches returns a paginated list of historical matches.
func (c *Client) GetMatches(ctx context.Context, opts *MatchesOptions) ([]map[string]any, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Season != "" {
			params.Set("season", opts.Season)
		}
		if opts.FromDate != "" {
			params.Set("from", opts.FromDate)
		}
		if opts.ToDate != "" {
			params.Set("to", opts.ToDate)
		}
		if opts.Limit > 0 {
			params.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", strconv.Itoa(opts.Offset))
		}
	}
	data, err := c.get(ctx, "/api/matches", params)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	return result, decode(data, &result)
}

// UpcomingMatchesOptions holds the optional query parameters for GetUpcomingMatches.
type UpcomingMatchesOptions struct {
	League string // league code, e.g. "BL1"
	Sort   string // sort order, default "date"
	Limit  int    // max results (default 50, max 200); 0 = use server default
}

// GetUpcomingMatches returns upcoming fixtures with ELO-difference predictions.
// Each entry contains "date", "homeTeam", "awayTeam", "homeElo", "awayElo",
// "eloDiff", "competition", "matchDay".
func (c *Client) GetUpcomingMatches(ctx context.Context, opts *UpcomingMatchesOptions) ([]map[string]any, error) {
	params := url.Values{}
	if opts != nil {
		if opts.League != "" {
			params.Set("league", opts.League)
		}
		if opts.Sort != "" {
			params.Set("sort", opts.Sort)
		}
		if opts.Limit > 0 {
			params.Set("limit", strconv.Itoa(opts.Limit))
		}
	}
	data, err := c.get(ctx, "/api/matches/upcoming", params)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	return result, decode(data, &result)
}

// GetMatch returns full details for a single match identified by matchID.
// The result contains "match", "homeRecentForm", "awayRecentForm",
// "homeStats", "awayStats", and "headToHead".
func (c *Client) GetMatch(ctx context.Context, matchID int64) (map[string]any, error) {
	path := "/api/matches/" + strconv.FormatInt(matchID, 10)
	data, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	return result, decode(data, &result)
}

// ---------------------------------------------------------------------------
// Seasons
// ---------------------------------------------------------------------------

// SeasonsResponse is the payload returned by GetSeasons.
type SeasonsResponse struct {
	Seasons      []string `json:"seasons"`
	LatestSeason string   `json:"latestSeason"`
}

// GetSeasons returns all available seasons and the most recent one.
func (c *Client) GetSeasons(ctx context.Context) (*SeasonsResponse, error) {
	data, err := c.get(ctx, "/api/seasons", nil)
	if err != nil {
		return nil, err
	}
	var result SeasonsResponse
	return &result, decode(data, &result)
}

// GetSeason returns per-team ELO change statistics for a given season.
// Pass an empty league string to retrieve data for all leagues.
// Each entry contains "teamName", "season", "startElo", "endElo", "change",
// "league", and "totalMatches".
func (c *Client) GetSeason(ctx context.Context, season, league string) ([]map[string]any, error) {
	path := "/api/seasons/" + url.PathEscape(season)
	params := url.Values{}
	if league != "" {
		params.Set("league", league)
	}
	data, err := c.get(ctx, path, params)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	return result, decode(data, &result)
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

// GetComparisonHistory returns historical ELO time-series for multiple teams.
// teams must contain at least two team names.
// The result is a map from team name to a list of data points, each containing
// "Date" (epoch ms) and "ELO".
func (c *Client) GetComparisonHistory(ctx context.Context, teams []string) (map[string]any, error) {
	if len(teams) < 2 {
		return nil, fmt.Errorf("elevenelo: at least two team names are required for comparison")
	}
	params := url.Values{}
	params.Set("teams", strings.Join(teams, ","))
	data, err := c.get(ctx, "/api/comparison/history", params)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	return result, decode(data, &result)
}
