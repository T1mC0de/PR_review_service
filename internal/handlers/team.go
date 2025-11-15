package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
	"pr-review-service/internal/storage"
)

type TeamHandler struct {
	*BaseHandler
}

func NewTeamHandler(storage *storage.Storage) *TeamHandler {
	return &TeamHandler{
		BaseHandler: NewBaseHandler(storage),
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if req.TeamName == "" || len(req.Members) == 0 {
		http.Error(w, "Team name and members are required", http.StatusBadRequest)
		return
	}

	dbMembers := make([]models.DBUser, len(req.Members))
	for i, member := range req.Members {
		dbMembers[i] = models.DBUser{
			UserID:    member.UserID,
			Username:  member.Username,
			IsActive:  member.IsActive,
			CreatedAt: time.Now(),
		}
	}

	err := h.storage.CreateTeam(r.Context(), req.TeamName, dbMembers)
	if err != nil {
		switch err {
		case constants.ErrTeamExists:
			h.sendError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
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

	resp := models.CreateTeamResponse{
		Team: models.TeamResponse(req),
	}

	h.sendJSON(w, http.StatusCreated, resp)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name is required", http.StatusBadRequest)
		return
	}

	team, members, err := h.storage.GetTeam(r.Context(), teamName)
	if err != nil {
		switch err {
		case constants.ErrTeamNotFound:
			http.Error(w, "Team not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	resp := models.TeamResponse{
		TeamName: team.Name,
		Members:  make([]models.TeamMember, len(members)),
	}

	for i, member := range members {
		resp.Members[i] = models.TeamMember{
			UserID:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	h.sendJSON(w, http.StatusOK, resp)
}
