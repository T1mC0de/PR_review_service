package handlers

import (
	"encoding/json"
	"net/http"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
	"pr-review-service/internal/storage"
)

type PullRequestHandler struct {
	*BaseHandler
}

func NewPullRequestHandler(storage *storage.Storage) *PullRequestHandler {
	return &PullRequestHandler{
		BaseHandler: NewBaseHandler(storage),
	}
}

func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		http.Error(
			w,
			"pull_request_id, pull_request_name and author_id are required",
			http.StatusBadRequest,
		)

		return
	}

	err := h.storage.CreatePullRequest(
		r.Context(),
		req.PullRequestID,
		req.PullRequestName,
		req.AuthorID,
	)
	if err != nil {
		switch err {
		case constants.ErrPRExists:
			h.sendError(w, http.StatusBadRequest, "PR_EXISTS", "pull_request_id already exists")
		case constants.ErrUserTeamNotFound:
			h.sendError(
				w,
				http.StatusBadRequest,
				"USER_TEAM_NOT_FOUND",
				"author_id does not belong to any team",
			)
		default:
			h.sendError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Internal server error",
			)
		}
	}

	resp := models.CreatePRRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}

	h.sendJSON(w, http.StatusCreated, resp)
}

func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if req.PullRequestID == "" {
		http.Error(
			w,
			"pull_request_id is required",
			http.StatusBadRequest,
		)

		return
	}

	err := h.storage.MergePullRequest(
		r.Context(),
		req.PullRequestID,
	)
	if err != nil {
		switch err {
		case constants.ErrPRNotFound:
			h.sendError(w, http.StatusNotFound, "PR_NOT_FOUND", "pull_request_id not found")
		default:
			h.sendError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Internal server error",
			)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PullRequestHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ReassignRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		http.Error(
			w,
			"pull_request_id and old_user_id are required",
			http.StatusBadRequest,
		)

		return
	}

	newReviewerID, err := h.storage.ReassignPullRequest(
		r.Context(),
		req.PullRequestID,
		req.OldUserID,
	)
	if err != nil {
		switch err {
		case constants.ErrPRNotFound:
			h.sendError(w, http.StatusNotFound, "PR_NOT_FOUND", "pull_request_id not found")
		case constants.ErrPRMerged:
			h.sendError(w, http.StatusBadRequest, "PR_MERGED", "Cannot reassign a merged PR")
		case constants.ErrNotAssigned:
			h.sendError(
				w,
				http.StatusBadRequest,
				"NOT_ASSIGNED",
				"old_user_id is not assigned to the pull request",
			)
		case constants.ErrUserTeamNotFound:
			h.sendError(
				w,
				http.StatusBadRequest,
				"USER_TEAM_NOT_FOUND",
				"old_user_id does not belong to any team",
			)
		case constants.ErrNoCandidate:
			h.sendError(
				w,
				http.StatusBadRequest,
				"NO_CANDIDATE",
				"No candidate available for reassignment",
			)
		default:
			h.sendError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Internal server error",
			)
		}

		return
	}

	updatedPR, err := h.storage.GetPR(r.Context(), req.PullRequestID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get updated PR")
		return
	}

	// Возвращаем данные по спецификации
	resp := map[string]interface{}{
		"pr":          updatedPR,
		"replaced_by": newReviewerID,
	}

	h.sendJSON(w, http.StatusOK, resp)
}
