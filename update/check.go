package update

import (
	"context"
	"log/slog"
	"time"

	"kanba/pkg/semver"
)

const maxAttempts = 5

var (
	backoffSchedule = []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
	}

	// fetchLatestTagFn is swapped out in tests to avoid real network calls.
	fetchLatestTagFn = fetchLatestTag
)

// CheckLatest checks GitHub for the latest release and reports whether it is
// newer than currentVersion. On persistent failure (all attempts exhausted,
// or a malformed response) it returns available=false, err=nil: callers
// should treat that as "nothing to report", not as an error to surface.
func CheckLatest(ctx context.Context, currentVersion string) (latest string, available bool, err error) {
	var tag string
	var lastErr error

	for attempt := range maxAttempts {
		tag, lastErr = fetchLatestTagFn(ctx)
		if lastErr == nil {
			break
		}

		if isNonRetryable(lastErr) {
			slog.Debug("update check: permanent error fetching latest release, not retrying", "error", lastErr)
			return "", false, nil
		}

		select {
		case <-time.After(backoffSchedule[attempt]):
		case <-ctx.Done():
			return "", false, nil
		}
	}

	if lastErr != nil {
		slog.Debug("update check: giving up after exhausting retries", "error", lastErr, "attempts", maxAttempts)
		return "", false, nil
	}

	cmp, err := semver.Compare(tag, currentVersion)
	if err != nil {
		slog.Debug("update check: malformed release tag from GitHub", "tag", tag, "error", err)
		return "", false, nil
	}

	return tag, cmp > 0, nil
}
