package storage

import (
	"context"

	"pr-review-service/internal/constants"
)

func (s *Storage) CreatePullRequest(
	ctx context.Context,
	prID string,
	prName string,
	prAuthorID string,
) error {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var exists bool

	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).
		Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return constants.ErrPRExists
	}

	teamID, err := s.getUserTeamID(ctx, prAuthorID)
	if err != nil {
		switch err {
		case constants.ErrUserNotFound:
			return constants.ErrUserTeamNotFound
		default:
			return err
		}
	}

	activeMembers, err := s.getActiveMembersExcept(ctx, *teamID, prAuthorID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
	       INSERT INTO pull_requests (pull_request_id, title, author_id) 
	       VALUES ($1, $2, $3)`,
		prID, prName, prAuthorID)
	if err != nil {
		return err
	}

	for _, memberID := range activeMembers {
		_, err = tx.ExecContext(ctx, `
		      INSERT INTO pr_reviewers (pr_id, reviewer_id) 
		      VALUES ($1, $2)`,
			prID, memberID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) MergePullRequest(ctx context.Context, prID string) error {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var status string

	err = tx.QueryRowContext(ctx, "SELECT status FROM pull_requests WHERE pull_request_id = $1", prID).
		Scan(&status)
	if err != nil {
		if err == constants.ErrNoRows {
			return constants.ErrPRNotFound
		}

		return err
	}

	switch status {
	case "MERGED":
		return tx.Commit()
	default:
		_, err = tx.ExecContext(ctx, `
		       UPDATE pull_requests 
		       SET status = $1, merged_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		       WHERE pull_request_id = $2`,
			"MERGED", prID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
