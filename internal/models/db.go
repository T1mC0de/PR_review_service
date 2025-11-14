package models

import (
	"time"
)

type DBUser struct {
	ID        int       `db:"id"`
	UserID    string    `db:"user_id"`
	Username  string    `db:"username"`
	IsActive  bool      `db:"is_active"`
	TeamID    int       `db:"team_id"`
	CreatedAt time.Time `db:"created_at"`
}

type DBTeam struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type DBPullRequest struct {
	ID            int        `db:"id"`
	PullRequestID string     `db:"pull_request_id"`
	Title         string     `db:"title"`
	AuthorID      string     `db:"author_id"`
	Status        string     `db:"status"`
	Reviewers     []string   `db:"reviewers"`
	CreatedAt     time.Time  `db:"created_at"`
	MergedAt      *time.Time `db:"merged_at"`
}

type DBPRReviewer struct {
	ID         int       `db:"id"`
	PrID       string    `db:"pr_id"`
	ReviewerID string    `db:"reviewer_id"`
	AssignedAt time.Time `db:"assigned_at"`
}
