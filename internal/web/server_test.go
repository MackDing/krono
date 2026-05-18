package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/krono-sh/krono/internal/store"
)

func newTestServer(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "krono.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return NewServer(st).Handler(), st
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func TestListJobsEmpty(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(h, "GET", "/api/jobs", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var jobs []store.Job
	if err := json.Unmarshal(rec.Body.Bytes(), &jobs); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, rec.Body.String())
	}
	if len(jobs) != 0 {
		t.Fatalf("got %d jobs, want 0", len(jobs))
	}
}

func TestCreateThenListJob(t *testing.T) {
	h, _ := newTestServer(t)
	body := `{"name":"backup","schedule":"0 3 * * *","type":"shell","command":"backup.sh","enabled":true}`
	rec := do(h, "POST", "/api/jobs", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 (body: %s)", rec.Code, rec.Body.String())
	}
	var created store.Job
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID == 0 || created.Name != "backup" {
		t.Fatalf("created job wrong: %+v", created)
	}

	rec = do(h, "GET", "/api/jobs", "")
	var jobs []store.Job
	if err := json.Unmarshal(rec.Body.Bytes(), &jobs); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(jobs))
	}
}

func TestCreateJobRejectsBadSchedule(t *testing.T) {
	h, _ := newTestServer(t)
	body := `{"name":"x","schedule":"not-a-cron","type":"shell","command":"echo x","enabled":true}`
	rec := do(h, "POST", "/api/jobs", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for an unparseable schedule (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestGetUpdateDeleteJob(t *testing.T) {
	h, st := newTestServer(t)
	job := &store.Job{Name: "j", Schedule: "@every 1m", Type: "shell", Command: "echo a", Enabled: true}
	if err := st.CreateJob(job); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	id := strconv.FormatInt(job.ID, 10)

	if rec := do(h, "GET", "/api/jobs/"+id, ""); rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", rec.Code)
	}

	upd := `{"name":"j2","schedule":"@every 2m","type":"shell","command":"echo b","enabled":false}`
	if rec := do(h, "PUT", "/api/jobs/"+id, upd); rec.Code != http.StatusOK {
		t.Fatalf("put status = %d, want 200", rec.Code)
	}
	got, _ := st.GetJob(job.ID)
	if got.Name != "j2" || got.Enabled {
		t.Fatalf("update not applied: %+v", got)
	}

	if rec := do(h, "DELETE", "/api/jobs/"+id, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", rec.Code)
	}
	if _, err := st.GetJob(job.ID); err == nil {
		t.Fatal("job still exists after delete")
	}
}

func TestGetMissingJob(t *testing.T) {
	h, _ := newTestServer(t)
	if rec := do(h, "GET", "/api/jobs/9999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestListRuns(t *testing.T) {
	h, st := newTestServer(t)
	job := &store.Job{Name: "j", Schedule: "@every 1m", Type: "shell", Command: "echo a", Enabled: true}
	if err := st.CreateJob(job); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if err := st.CreateRun(&store.Run{JobID: job.ID, StartedAt: time.Now(), EndedAt: time.Now(), Status: "success"}); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	rec := do(h, "GET", "/api/jobs/"+strconv.FormatInt(job.ID, 10)+"/runs", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var runs []store.Run
	if err := json.Unmarshal(rec.Body.Bytes(), &runs); err != nil {
		t.Fatalf("decode runs: %v", err)
	}
	if len(runs) != 1 || runs[0].Status != "success" {
		t.Fatalf("runs = %+v, want 1 with status success", runs)
	}
}

func TestCreateJobValidationErrors(t *testing.T) {
	h, _ := newTestServer(t)
	cases := map[string]string{
		"missing name":    `{"name":"","schedule":"@every 1m","command":"echo x"}`,
		"missing command": `{"name":"x","schedule":"@every 1m","command":""}`,
		"bad type":        `{"name":"x","schedule":"@every 1m","command":"echo x","type":"ftp"}`,
	}
	for name, body := range cases {
		if rec := do(h, "POST", "/api/jobs", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
}

func TestInvalidJobID(t *testing.T) {
	h, _ := newTestServer(t)
	if rec := do(h, "GET", "/api/jobs/not-a-number", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for a non-numeric id", rec.Code)
	}
}

func TestMalformedJSONBody(t *testing.T) {
	h, _ := newTestServer(t)
	if rec := do(h, "POST", "/api/jobs", `{not valid json`); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for a malformed JSON body", rec.Code)
	}
}

func TestUpdateMissingJob(t *testing.T) {
	h, _ := newTestServer(t)
	body := `{"name":"x","schedule":"@every 1m","type":"shell","command":"echo x","enabled":true}`
	if rec := do(h, "PUT", "/api/jobs/9999", body); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for PUT on a missing job", rec.Code)
	}
}

func TestServesDashboardHTML(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(h, "GET", "/", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("<html")) {
		t.Fatal("response is not an HTML document")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Krono")) {
		t.Fatal("dashboard HTML does not mention Krono")
	}
}
