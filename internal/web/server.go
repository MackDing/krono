// Package web serves the Krono dashboard: an embedded HTML UI and a JSON API
// over the job store.
package web

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/MackDing/krono/internal/schedule"
	"github.com/MackDing/krono/internal/store"
)

//go:embed index.html
var indexHTML []byte

// Server serves the Krono dashboard HTTP API, backed by the job store.
type Server struct {
	store *store.Store
}

// NewServer returns a Server backed by st.
func NewServer(st *store.Store) *Server {
	return &Server{store: st}
}

// Handler returns the HTTP handler for the dashboard UI and its JSON API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("GET /api/jobs", s.listJobs)
	mux.HandleFunc("POST /api/jobs", s.createJob)
	mux.HandleFunc("GET /api/jobs/{id}", s.getJob)
	mux.HandleFunc("PUT /api/jobs/{id}", s.updateJob)
	mux.HandleFunc("DELETE /api/jobs/{id}", s.deleteJob)
	mux.HandleFunc("GET /api/jobs/{id}/runs", s.listRuns)
	return mux
}

// index serves the embedded dashboard page.
func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

// jobInput is the create/update request body.
type jobInput struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Type     string `json:"type"`
	Command  string `json:"command"`
	Enabled  bool   `json:"enabled"`
}

// validate reports whether the input describes a usable job.
func (in jobInput) validate() error {
	if in.Name == "" {
		return errors.New("name is required")
	}
	if in.Command == "" {
		return errors.New("command is required")
	}
	if in.Type != "" && in.Type != "shell" && in.Type != "http" {
		return errors.New(`type must be "shell" or "http"`)
	}
	if _, err := schedule.Parse(in.Schedule); err != nil {
		return err
	}
	return nil
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.store.ListJobs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if jobs == nil {
		jobs = []*store.Job{}
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	in, err := decodeJobInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job := &store.Job{
		Name: in.Name, Schedule: in.Schedule, Type: jobType(in.Type),
		Command: in.Command, Enabled: in.Enabled,
	}
	if err := s.store.CreateJob(job); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job, err := s.store.GetJob(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) updateJob(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job, err := s.store.GetJob(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	in, err := decodeJobInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job.Name, job.Schedule, job.Command, job.Enabled = in.Name, in.Schedule, in.Command, in.Enabled
	job.Type = jobType(in.Type)
	if err := s.store.UpdateJob(job); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.store.DeleteJob(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	runs, err := s.store.ListRuns(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if runs == nil {
		runs = []*store.Run{}
	}
	writeJSON(w, http.StatusOK, runs)
}

// --- helpers ---------------------------------------------------------------

func decodeJobInput(r *http.Request) (jobInput, error) {
	var in jobInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return in, errors.New("invalid JSON body")
	}
	if err := in.validate(); err != nil {
		return in, err
	}
	return in, nil
}

func jobType(t string) string {
	if t == "" {
		return "shell"
	}
	return t
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid job id")
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
