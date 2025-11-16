package models

import (
	"time"
)

// Team API модели
type CreateTeamRequest struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type CreateTeamResponse struct {
	Team TeamResponse `json:"team"`
}

// User API модели
type SetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserResponseWrapper struct {
	User UserResponse `json:"user"`
}

// PullRequest API модели
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PRResponse struct {
	PullRequestID     string    `json:"pull_request_id"`
	PullRequestName   string    `json:"pull_request_name"`
	AuthorID          string    `json:"author_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	CreatedAt         time.Time `json:"created_at"`
	MergedAt          time.Time `json:"merged_at"`
}

type PRResponseWrapper struct {
	PR PRResponse `json:"pr"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type ReassignResponse struct {
	PR         PRResponse `json:"pr"`
	ReplacedBy string     `json:"replaced_by"`
}

type UserPRsResponse struct {
	UserID       string            `json:"user_id"`
	PullRequests []PRShortResponse `json:"pull_requests"`
}

type PRShortResponse struct {
	PullRequestID string `json:"pull_request_id"`
	Title         string `json:"pull_request_name"`
	AuthorID      string `json:"author_id"`
	Status        string `json:"status"`
}
