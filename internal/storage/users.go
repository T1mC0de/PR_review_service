package storage

import (
	"context"
	"database/sql"
	"log"

	"github.com/lib/pq"
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

func (s *Storage) MassDeactivateUsers(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (*models.MassDeactivateResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	response := &models.MassDeactivateResponse{
		TeamName:         teamName,
		DeactivatedUsers: []string{},
		ReassignedPRs:    []string{},
		FailedReassigns:  []string{},
	}

	var teamID int

	err = tx.QueryRowContext(ctx, "SELECT id FROM teams WHERE name = $1", teamName).Scan(&teamID)
	if err != nil {
		return nil, constants.ErrTeamNotFound
	}

	deactivateQuery := `
		UPDATE users 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP 
		WHERE team_id = $1 AND user_id = ANY($2) AND is_active = true
		RETURNING user_id`

	rows, err := tx.QueryContext(ctx, deactivateQuery, teamID, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string

		scanErr := rows.Scan(&userID)
		if scanErr != nil {
			return nil, scanErr
		}

		response.DeactivatedUsers = append(response.DeactivatedUsers, userID)
	}

	if len(response.DeactivatedUsers) == 0 {
		return response, tx.Commit()
	}

	prsToReassign, err := s.findPRsWithDeactivatedReviewers(ctx, tx, response.DeactivatedUsers)
	if err != nil {
		return nil, err
	}

	for _, prID := range prsToReassign {
		success, err := s.safelyReassignDeactivatedReviewers(
			ctx,
			tx,
			prID,
			response.DeactivatedUsers,
		)
		if err != nil {
			response.FailedReassigns = append(response.FailedReassigns, prID)
		} else if success {
			response.ReassignedPRs = append(response.ReassignedPRs, prID)
		}
	}

	return response, tx.Commit()
}

func (s *Storage) findPRsWithDeactivatedReviewers(
	ctx context.Context,
	tx *sql.Tx,
	deactivatedUsers []string,
) ([]string, error) {
	query := `
		SELECT DISTINCT pr.pull_request_id
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pr_id
		WHERE pr.status = 'OPEN' 
		AND prr.reviewer_id = ANY($1)`

	rows, err := tx.QueryContext(ctx, query, pq.Array(deactivatedUsers))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var prID string

		scanErr := rows.Scan(&prID)
		if scanErr != nil {
			return nil, scanErr
		}

		prIDs = append(prIDs, prID)
	}

	return prIDs, nil
}

func (s *Storage) safelyReassignDeactivatedReviewers(
	ctx context.Context,
	tx *sql.Tx,
	prID string,
	deactivatedUsers []string,
) (bool, error) {
	var authorID, status string
	if err := tx.QueryRowContext(ctx, "SELECT author_id, status FROM pull_requests WHERE pull_request_id = $1", prID).Scan(&authorID, &status); err != nil {
		return false, err
	}

	if status == "MERGED" {
		return true, nil
	}

	authorWillBeDeactivated := false

	for _, u := range deactivatedUsers {
		if u == authorID {
			authorWillBeDeactivated = true
			break
		}
	}

	if authorWillBeDeactivated {
		_, err := tx.ExecContext(
			ctx,
			`DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = ANY($2)`,
			prID,
			pq.Array(deactivatedUsers),
		)

		return err == nil, err
	}

	var authorTeamID int
	if err := tx.QueryRowContext(ctx, "SELECT team_id FROM users WHERE user_id = $1", authorID).Scan(&authorTeamID); err != nil {
		return false, err
	}

	availableReviewers, err := s.findAvailableReviewers(
		ctx,
		tx,
		authorTeamID,
		authorID,
		deactivatedUsers,
	)
	if err != nil {
		return false, err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = ANY($2)`, prID, pq.Array(deactivatedUsers)); err != nil {
		return false, err
	}

	rows, err := tx.QueryContext(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1", prID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var currentReviewers []string
	for rows.Next() {
		var r string
		if scanErr := rows.Scan(&r); scanErr != nil {
			return false, scanErr
		}

		currentReviewers = append(currentReviewers, r)
	}

	needed := 2 - len(currentReviewers)
	if needed > 0 && len(availableReviewers) > 0 {
		add := availableReviewers
		if len(add) > needed {
			add = add[:needed]
		}

		for _, r := range add {
			if _, err = tx.ExecContext(ctx, `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`, prID, r); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func (s *Storage) findAvailableReviewers(
	ctx context.Context,
	tx *sql.Tx,
	teamID int,
	excludeAuthor string,
	excludeUsers []string,
) ([]string, error) {
	query := `
        SELECT user_id FROM users 
        WHERE team_id = $1 
        AND is_active = true 
        AND user_id != $2 
        AND user_id != ALL($3)
        ORDER BY user_id
        LIMIT 10`

	rows, err := tx.QueryContext(ctx, query, teamID, excludeAuthor, pq.Array(excludeUsers))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewer string
		if err := rows.Scan(&reviewer); err != nil {
			return nil, err
		}

		reviewers = append(reviewers, reviewer)
	}

	return reviewers, nil
}
