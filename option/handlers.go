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

func (h *Handler) createPoll(w http.ResponseWriter, r *http.Request) {

	var req createPollRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Question == "" || len(req.Options) < 2 {
		writeError(w, http.StatusBadRequest, "опрос бессмысленный")
		return
	}

	poll := h.store.CreatePoll(req.Question, req.Options)
	writeJSON(w, http.StatusCreated, poll)
}
