package handlers

import (
	"encoding/json"
	"net/http"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
	"pr-review-service/internal/storage"
)

type pullRequestHandler struct {
	*BaseHandler
}

func NewPullRequestHandler(storage *storage.Storage) *pullRequestHandler {
	return &pullRequestHandler{
		BaseHandler: NewBaseHandler(storage),
	}
}

func (h *pullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "pull_request_id, pull_request_name and author_id are required", http.StatusBadRequest)
		return
	}

	err := h.storage.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch err {
		case constants.ErrPRExists:
			h.sendError(w, http.StatusBadRequest, "PR_EXISTS", "pull_request_id already exists")
		case constants.ErrUserTeamNotFound:
			h.sendError(w, http.StatusBadRequest, "USER_TEAM_NOT_FOUND", "author_id does not belong to any team")
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



