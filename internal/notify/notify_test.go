package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MackDing/krono/internal/store"
)

func TestWebhookPostsRunDetails(t *testing.T) {
	var (
		mu   sync.Mutex
		body []byte
		ct   string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		body, _ = io.ReadAll(r.Body)
		ct = r.Header.Get("Content-Type")
	}))
	defer srv.Close()

	job := &store.Job{Name: "backup", Command: "backup.sh"}
	run := &store.Run{Status: "failure", ExitCode: 7, StartedAt: time.Now(), Output: "disk full"}

	if err := Webhook(context.Background(), srv.URL, job, run); err != nil {
		t.Fatalf("Webhook: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode payload: %v (body: %s)", err, body)
	}
	if got["job"] != "backup" || got["status"] != "failure" {
		t.Fatalf("payload = %v, want job=backup status=failure", got)
	}
}

func TestWebhookEmptyURLIsNoop(t *testing.T) {
	if err := Webhook(context.Background(), "", &store.Job{Name: "x"}, &store.Run{Status: "failure"}); err != nil {
		t.Fatalf("Webhook with an empty URL should be a no-op, got: %v", err)
	}
}

func TestWebhookReportsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := Webhook(context.Background(), srv.URL, &store.Job{Name: "x"}, &store.Run{Status: "failure"})
	if err == nil {
		t.Fatal("Webhook should return an error when the endpoint responds 500")
	}
}
