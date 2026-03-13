package elevenelo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	elevenelo "github.com/Chafficui/11elo-clients/go"
)

// newTestClient builds a Client pointed at the given test server URL.
func newTestClient(t *testing.T, baseURL string) *elevenelo.Client {
	t.Helper()
	c, err := elevenelo.NewClient("11e_fre_testkey", elevenelo.WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// serve is a tiny helper that creates an httptest.Server responding with the
// given status code and JSON body.
func serve(t *testing.T, status int, body any, headers ...map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the API key header is forwarded.
		if r.Header.Get("X-API-Key") == "" {
			t.Error("expected X-API-Key header to be set")
		}
		for _, h := range headers {
			for k, v := range h {
				w.Header().Set(k, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewClient_EmptyKey(t *testing.T) {
	_, err := elevenelo.NewClient("")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewClient_Options(t *testing.T) {
	c, err := elevenelo.NewClient("key", elevenelo.WithBaseURL("http://localhost:3001"))
	if err != nil {
		t.Fatal(err)
	}
	// Smoke-check that the custom base URL is stored by hitting an unreachable
	// endpoint and verifying the error (not an auth error).
	_, clientErr := c.GetTeams(context.Background())
	if clientErr == nil {
		t.Fatal("expected network error for localhost:3001")
	}
}

// ---------------------------------------------------------------------------
// Teams
// ---------------------------------------------------------------------------

func TestGetTeams(t *testing.T) {
	payload := []map[string]any{{"teamName": "Bayern München", "currentElo": float64(1847)}}
	srv := serve(t, http.StatusOK, payload)
	defer srv.Close()

	result, err := newTestClient(t, srv.URL).GetTeams(context.Background())
	if err != nil {
		t.Fatalf("GetTeams: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 team, got %d", len(result))
	}
	if result[0]["teamName"] != "Bayern München" {
		t.Errorf("unexpected teamName: %v", result[0]["teamName"])
	}
}

func TestGetTeam(t *testing.T) {
	payload := map[string]any{"team": map[string]any{"name": "Bayern München"}}
	var capturedURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetTeam(context.Background(), "Bayern München")
	if err != nil {
		t.Fatalf("GetTeam: %v", err)
	}
	want := "/api/teams/Bayern%20M%C3%BCnchen"
	if capturedURI != want {
		t.Errorf("URI = %q, want %q", capturedURI, want)
	}
}

func TestGetHeadToHead(t *testing.T) {
	payload := []map[string]any{{"result": "2:1"}}
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetHeadToHead(context.Background(), "Bayern München", "Borussia Dortmund")
	if err != nil {
		t.Fatalf("GetHeadToHead: %v", err)
	}
	if capturedPath == "" || capturedPath == "/api/teams" {
		t.Errorf("unexpected path: %q", capturedPath)
	}
}

// ---------------------------------------------------------------------------
// Matches
// ---------------------------------------------------------------------------

func TestGetMatches_WithParams(t *testing.T) {
	payload := []map[string]any{{"id": float64(1)}}
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	opts := &elevenelo.MatchesOptions{Season: "2024/2025", Limit: 10}
	_, err := newTestClient(t, srv.URL).GetMatches(context.Background(), opts)
	if err != nil {
		t.Fatalf("GetMatches: %v", err)
	}
	if capturedQuery == "" {
		t.Error("expected query string to be set")
	}
}

func TestGetUpcomingMatches(t *testing.T) {
	payload := []map[string]any{{"homeTeam": "BVB"}}
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	opts := &elevenelo.UpcomingMatchesOptions{League: "BL1", Limit: 5}
	_, err := newTestClient(t, srv.URL).GetUpcomingMatches(context.Background(), opts)
	if err != nil {
		t.Fatalf("GetUpcomingMatches: %v", err)
	}
	if capturedQuery == "" {
		t.Error("expected query string with league param")
	}
}

func TestGetMatch(t *testing.T) {
	payload := map[string]any{"match": map[string]any{"id": float64(12345)}}
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetMatch(context.Background(), 12345)
	if err != nil {
		t.Fatalf("GetMatch: %v", err)
	}
	if capturedPath != "/api/matches/12345" {
		t.Errorf("path = %q, want /api/matches/12345", capturedPath)
	}
}

// ---------------------------------------------------------------------------
// Seasons
// ---------------------------------------------------------------------------

func TestGetSeasons(t *testing.T) {
	payload := map[string]any{
		"seasons":      []string{"2025/2026", "2024/2025"},
		"latestSeason": "2025/2026",
	}
	srv := serve(t, http.StatusOK, payload)
	defer srv.Close()

	result, err := newTestClient(t, srv.URL).GetSeasons(context.Background())
	if err != nil {
		t.Fatalf("GetSeasons: %v", err)
	}
	if result.LatestSeason != "2025/2026" {
		t.Errorf("latestSeason = %q, want 2025/2026", result.LatestSeason)
	}
}

func TestGetSeason(t *testing.T) {
	payload := []map[string]any{{"teamName": "Bayern München", "change": float64(27)}}
	var capturedPath, capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetSeason(context.Background(), "2024/2025", "BL1")
	if err != nil {
		t.Fatalf("GetSeason: %v", err)
	}
	if capturedPath == "" {
		t.Error("expected path to be set")
	}
	if capturedQuery == "" {
		t.Error("expected league query param")
	}
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

func TestGetComparisonHistory(t *testing.T) {
	payload := map[string]any{
		"Bayern München":    []any{map[string]any{"Date": float64(1709856000000), "ELO": float64(1847)}},
		"Borussia Dortmund": []any{map[string]any{"Date": float64(1709856000000), "ELO": float64(1720)}},
	}
	srv := serve(t, http.StatusOK, payload)
	defer srv.Close()

	result, err := newTestClient(t, srv.URL).GetComparisonHistory(
		context.Background(),
		[]string{"Bayern München", "Borussia Dortmund"},
	)
	if err != nil {
		t.Fatalf("GetComparisonHistory: %v", err)
	}
	if _, ok := result["Bayern München"]; !ok {
		t.Error("expected Bayern München in result")
	}
}

func TestGetComparisonHistory_TooFewTeams(t *testing.T) {
	c, _ := elevenelo.NewClient("key")
	_, err := c.GetComparisonHistory(context.Background(), []string{"Bayern München"})
	if err == nil {
		t.Fatal("expected error for fewer than two teams")
	}
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func TestAuthenticationError(t *testing.T) {
	srv := serve(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetTeams(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*elevenelo.AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestRateLimitError(t *testing.T) {
	srv := serve(t, http.StatusTooManyRequests, nil, map[string]string{
		"X-RateLimit-Reset": "2026-03-13T00:00:00Z",
	})
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetTeams(context.Background())
	rle, ok := err.(*elevenelo.RateLimitError)
	if !ok {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if rle.ResetAt != "2026-03-13T00:00:00Z" {
		t.Errorf("ResetAt = %q, want 2026-03-13T00:00:00Z", rle.ResetAt)
	}
}

func TestNotFoundError(t *testing.T) {
	srv := serve(t, http.StatusNotFound, nil)
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetTeam(context.Background(), "Unknown FC")
	if _, ok := err.(*elevenelo.NotFoundError); !ok {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestAPIError(t *testing.T) {
	srv := serve(t, http.StatusInternalServerError, map[string]string{"error": "internal"})
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetTeams(context.Background())
	ae, ok := err.(*elevenelo.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if ae.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", ae.StatusCode)
	}
}
