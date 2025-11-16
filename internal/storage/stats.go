package storage

import (
	"context"
	"database/sql"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
)

func (s *Storage) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	stats := &models.StatsResponse{}

	err = tx.QueryRowContext(ctx, `
		WITH pr_counts AS (
			SELECT 
				COUNT(*) AS total_prs,
				COUNT(*) FILTER (WHERE status = 'OPEN') AS open_prs,
				COUNT(*) FILTER (WHERE status = 'MERGED') AS merged_prs
			FROM pull_requests
		), user_counts AS (
			SELECT 
				COUNT(*) AS total_users,
				COUNT(*) FILTER (WHERE is_active) AS active_users
			FROM users
		), assignment_counts AS (
			SELECT COUNT(*) AS total_assignments FROM pr_reviewers
		)
		SELECT 
			pr_counts.total_prs,
			pr_counts.open_prs,
			pr_counts.merged_prs,
			user_counts.total_users,
			user_counts.active_users,
			assignment_counts.total_assignments
		FROM pr_counts, user_counts, assignment_counts`).Scan(
		&stats.TotalPRs,
		&stats.OpenPRs,
		&stats.MergedPRs,
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.TotalAssignments,
	)
	if err != nil {
		return nil, err
	}

	reviewerRows, err := tx.QueryContext(ctx, `
		SELECT 
			u.user_id,
			u.username,
			COUNT(pr.reviewer_id) AS assignments
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.reviewer_id
		WHERE u.is_active = TRUE
		GROUP BY u.user_id, u.username
		ORDER BY assignments DESC
		LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer reviewerRows.Close()

	stats.TopReviewers = make([]models.UserAssignmentStats, 0, 10)
	for reviewerRows.Next() {
		var r models.UserAssignmentStats

		scanErr := reviewerRows.Scan(&r.UserID, &r.Username, &r.Assignments)
		if scanErr != nil {
			return nil, scanErr
		}

		stats.TopReviewers = append(stats.TopReviewers, r)
	}

	rowErr := reviewerRows.Err()
	if rowErr != nil {
		return nil, rowErr
	}

	prRows, err := tx.QueryContext(ctx, `
		SELECT 
			pr.pull_request_id,
			pr.title,
			pr.status,
			COUNT(prr.reviewer_id) AS assignments
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.pull_request_id = prr.pr_id
		GROUP BY pr.pull_request_id, pr.title, pr.status
		ORDER BY assignments DESC`)
	if err != nil {
		return nil, err
	}
	defer prRows.Close()

	stats.PRStats = []models.PullRequestAssignmentStats{}
	for prRows.Next() {
		var ps models.PullRequestAssignmentStats

		scanErr := prRows.Scan(&ps.PRID, &ps.Title, &ps.Status, &ps.Assignments)
		if scanErr != nil {
			return nil, scanErr
		}

		stats.PRStats = append(stats.PRStats, ps)
	}

	rowErr = prRows.Err()
	if rowErr != nil {
		return nil, rowErr
	}

	return stats, tx.Commit()
}
