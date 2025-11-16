package storage

import (
	"context"
	"log"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
)

func (s *Storage) SetIsActive(
	ctx context.Context,
	userID string,
	isActive bool,
) (*models.DBUser, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	var user models.DBUser

	err = tx.QueryRowContext(ctx, `
		UPDATE users 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
		RETURNING user_id, username, is_active, team_id`,
		isActive, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.IsActive,
		&user.TeamID,
	)
	if err != nil {
		switch err {
		case constants.ErrNoRows:
			return nil, constants.ErrUserNotFound
		default:
			return nil, err
		}
	}

	return &user, tx.Commit()
}

func (s *Storage) getUserTeamID(ctx context.Context, userID string) (*int, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	var teamID int

	err := s.db.QueryRowContext(ctx, `
		SELECT team_id 
		FROM users 
		WHERE user_id = $1`, userID).Scan(&teamID)
	if err != nil {
		switch err {
		case constants.ErrNoRows:
			return nil, constants.ErrTeamNotFound
		default:
			return nil, err
		}
	}

	return &teamID, nil
}

func (s *Storage) GetReview(ctx context.Context, userID string) ([]models.DBPullRequest, error) {
	log.Printf("=== STORAGE GetReview START: userID=%s", userID)

	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	var userExists bool

	err := s.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)",
		userID,
	).Scan(&userExists)
	if err != nil {
		return nil, err
	}

	if !userExists {
		return nil, constants.ErrUserNotFound
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			pr.pull_request_id,
			pr.title, 
			pr.author_id,
			pr.status
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pr_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.DBPullRequest

	for rows.Next() {
		var pr models.DBPullRequest

		err = rows.Scan(
			&pr.PullRequestID,
			&pr.Title,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, err
		}

		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
