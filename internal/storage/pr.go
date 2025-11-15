package storage

import (
	"context"

	"pr-review-service/internal/constants"
)

func (s *Storage) CreatePullRequest(ctx context.Context, pr_id string, prName string, prAuthorId string) error {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	
	var exists bool

	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", pr_id).
		Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return constants.ErrPRExists
	}

	teamId, err := s.getUserTeamId(ctx, prAuthorId)

	if err != nil {
		switch err {
		case constants.ErrUserNotFound:
			return constants.ErrUserTeamNotFound
		default:
			return err
		}
	}

	activeMembers, err := s.getActiveMembersExcept(ctx, *teamId, prAuthorId)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO pull_requests (pull_request_id, title, author_id) 
		VALUES ($1, $2, $3)`,
		pr_id, prName, prAuthorId)
	if err != nil {
		return err
	}

	for _, memberId := range activeMembers {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO pr_reviewers (pr_id, reviewer_id) 
			VALUES ($1, $2)`,
			pr_id, memberId)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}