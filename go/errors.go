package elevenelo

import "fmt"

// ElevenEloError is the base type for all errors returned by this client.
// All specific error types embed ElevenEloError.
type ElevenEloError struct {
	Message string
}

func (e *ElevenEloError) Error() string { return e.Message }

// AuthenticationError is returned when the API key is missing or invalid (HTTP 401).
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string { return e.Message }

// RateLimitError is returned when the daily rate limit for the API key tier has
// been exceeded (HTTP 429).
// ResetAt contains the value of the X-RateLimit-Reset response header, if set.
type RateLimitError struct {
	Message string
	ResetAt string
}

func (e *RateLimitError) Error() string {
	if e.ResetAt != "" {
		return fmt.Sprintf("%s (resets at %s)", e.Message, e.ResetAt)
	}
	return e.Message
}

// NotFoundError is returned when the requested resource does not exist (HTTP 404).
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string { return e.Message }

// APIError is returned for any unexpected non-2xx response not covered by the
// more specific error types.
// StatusCode holds the HTTP status code.
type APIError struct {
	Message    string
	StatusCode int
}

func (e *APIError) Error() string { return e.Message }
