// Package notify sends webhook notifications about job runs.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MackDing/krono/internal/store"
)

// payload is the JSON body POSTed to the webhook.
type payload struct {
	Job       string    `json:"job"`
	Status    string    `json:"status"`
	ExitCode  int       `json:"exitCode"`
	StartedAt time.Time `json:"startedAt"`
	Output    string    `json:"output"`
}

// Webhook POSTs a JSON notification about run to url. It is a no-op (returns
// nil) when url is empty, so callers may pass an unconfigured URL freely.
func Webhook(ctx context.Context, url string, job *store.Job, run *store.Run) error {
	if url == "" {
		return nil
	}
	body, err := json.Marshal(payload{
		Job:       job.Name,
		Status:    run.Status,
		ExitCode:  run.ExitCode,
		StartedAt: run.StartedAt,
		Output:    run.Output,
	})
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("notify: post to %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook %s returned %s", url, resp.Status)
	}
	return nil
}
