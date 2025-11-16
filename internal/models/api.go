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

type StatsResponse struct {
	TotalPRs         int64                        `json:"total_prs"`
	OpenPRs          int64                        `json:"open_prs"`
	MergedPRs        int64                        `json:"merged_prs"`
	TotalUsers       int64                        `json:"total_users"`
	ActiveUsers      int64                        `json:"active_users"`
	TotalAssignments int64                        `json:"total_assignments"`
	TopReviewers     []UserAssignmentStats        `json:"top_reviewers"`
	PRStats          []PullRequestAssignmentStats `json:"pr_stats"`
}

type UserAssignmentStats struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Assignments int64  `json:"assignments"`
}

type PullRequestAssignmentStats struct {
	PRID        string `json:"pr_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Assignments int64  `json:"assignments"`
}

type MassDeactivateResponse struct {
	TeamName         string   `json:"team_name"`
	DeactivatedUsers []string `json:"deactivated_users"`
	ReassignedPRs    []string `json:"reassigned_prs"`
	FailedReassigns  []string `json:"failed_reassigns"`
}

type MassDeactivateRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}
