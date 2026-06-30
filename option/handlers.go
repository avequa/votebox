package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	store  *Store
	worker *Worker
}

func NewHandler(store *Store, worker *Worker) *Handler {
	return &Handler{store: store, worker: worker}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/polls", h.createPoll)
	mux.HandleFunc("GET /api/polls", h.listPolls)
	mux.HandleFunc("GET /api/polls/{id}", h.getPoll)
	mux.HandleFunc("POST /api/polls/{id}/vote", h.vote)
}


// METHODS

// DTO запроса 
type createPollRequest struct {
	Question string `json:"question"`
	Options  []string `json:"options"`
}

func (h *Handler) createPoll(rw http.ResponseWriter, r *http.Request) {

	var req createPollRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(rw, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Question == "" || len(req.Options) < 2 {
		writeError(rw, http.StatusBadRequest, "опрос бессмысленный")
		return
	}

	poll := h.store.CreatePoll(req.Question, req.Options)
	writeJSON(rw, http.StatusCreated, poll)
}

func (h *Handler) listPolls(rw http.ResponseWriter, r *http.Request) {
	
	polls := h.store.ListPolls()
	writeJSON(rw, http.StatusOK, polls)
}

func (h *Handler) getPoll(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	poll, err := h.store.GetPoll(id)
	if errors.Is(err, ErrPollNotFound) {
		writeError(rw, http.StatusNotFound, "опрос не найден")
		return
	}
	writeJSON(rw, http.StatusOK, poll)
}

// DTO запроса 
type voteRequest struct {
	OptionID string `json:"option_id"`
}

func (h *Handler) vote(rw http.ResponseWriter, r *http.Request) {
	pollID := r.PathValue("id")
	var req voteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(rw, http.StatusBadRequest, "invalid json")
		return
	}

	job := VoteJob{PollID: pollID, OptionID: req.OptionID}
	ok := h.worker.Submit(job)
	if !ok {
		writeError(rw, http.StatusServiceUnavailable, "queue full")
		return
	}
	writeJSON(rw, http.StatusAccepted, map[string]string{"status": "accepted"})
}


// json хелперы

// writeJSON кодирует v в json + ставит заголовок и статус
func writeJSON(rw http.ResponseWriter, status int, v any) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	_ = json.NewEncoder(rw).Encode(v)
}

// writeError формат ошибки: {"error": "..."}
func writeError(rw http.ResponseWriter, status int, msg string) {
	writeJSON(rw, status, map[string]string{"error": msg})
}