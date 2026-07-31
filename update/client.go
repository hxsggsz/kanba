package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	releasesURL = "https://api.github.com/repos/hxsggsz/kanba/releases/latest"
	httpClient  = &http.Client{Timeout: 3 * time.Second}
)

type releaseResponse struct {
	TagName string `json:"tag_name"`
}

// statusError wraps an HTTP error response that carries a status code, so
// callers can distinguish permanent client errors (4xx: not found, rate
// limited, forbidden) from transient ones worth retrying.
type statusError struct {
	statusCode int
	err        error
}

func (e *statusError) Error() string { return e.err.Error() }
func (e *statusError) Unwrap() error { return e.err }

// isNonRetryable reports whether err represents a permanent failure (an
// HTTP 4xx response) that will not be fixed by retrying, as opposed to a
// network error or 5xx response which may succeed on a later attempt.
func isNonRetryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		return se.statusCode >= 400 && se.statusCode < 500
	}
	return false
}

func fetchLatestTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "kanba-update-checker")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("github api returned status %d", resp.StatusCode)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return "", &statusError{statusCode: resp.StatusCode, err: statusErr}
		}
		return "", statusErr
	}

	var rel releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("release response missing tag_name")
	}

	return rel.TagName, nil
}
