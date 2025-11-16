package handlers

import (
	"encoding/json"
	"net/http"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
	"pr-review-service/internal/storage"
)

// UserHandler - структура для обработки запросов, связанных с пользователями
type UserHandler struct {
	*BaseHandler
}

func NewUserHandler(storage *storage.Storage) *UserHandler {
	return &UserHandler{
		BaseHandler: NewBaseHandler(storage),
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		IsActive *bool  `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if req.IsActive == nil {
		http.Error(w, "is_active is required", http.StatusBadRequest)
		return
	}

	user, err := h.storage.SetIsActive(r.Context(), req.UserID, *req.IsActive)
	if err != nil {
		switch err {
		case constants.ErrUserNotFound:
			h.sendError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		default:
			h.sendError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Failed to update user status",
			)

			return
		}
	}

	h.sendJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(
			w,
			"user_id is required",
			http.StatusBadRequest,
		)

		return
	}

	dbPRs, err := h.storage.GetReview(r.Context(), userID)
	if err != nil {
		switch err {
		case constants.ErrUserNotFound:
			h.sendError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
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

	prShorts := make([]models.PRShortResponse, len(dbPRs))
	for i, dbPR := range dbPRs {
		prShorts[i] = models.PRShortResponse{
			PullRequestID: dbPR.PullRequestID,
			Title:         dbPR.Title,
			AuthorID:      dbPR.AuthorID,
			Status:        dbPR.Status,
		}
	}

	resp := models.UserPRsResponse{
		UserID:       userID,
		PullRequests: prShorts,
	}

	h.sendJSON(w, http.StatusOK, resp)
}
