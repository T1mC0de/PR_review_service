package storage

import (
	"context"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
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

func (s *Storage) ReassignPullRequest(ctx context.Context, prID string, oldUserID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck

	var status string
    err = tx.QueryRowContext(ctx, "SELECT status FROM pull_requests WHERE pull_request_id = $1", prID).Scan(&status)
    if err != nil {
		if err == constants.ErrNoRows {
			return "", constants.ErrPRNotFound
		}
        return "", err
    }
    if status == "MERGED" {
        return "", constants.ErrPRMerged
    }

	var assigned bool
    err = tx.QueryRowContext(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2
        )`, prID, oldUserID).Scan(&assigned)
    if err != nil {
        return "", err
    }
    if !assigned {
        return "", constants.ErrNotAssigned
    }

	var teamID int
    err = tx.QueryRowContext(ctx, "SELECT team_id FROM users WHERE user_id = $1", oldUserID).Scan(&teamID)
    if err != nil {
        return "", constants.ErrUserTeamNotFound
    }

	var authorID string
    err = tx.QueryRowContext(ctx, "SELECT author_id FROM pull_requests WHERE pull_request_id = $1", prID).Scan(&authorID)
    if err != nil {
        return "", constants.ErrPRNotFound
    }

	var newReviewerID string
    err = tx.QueryRowContext(ctx, `
        SELECT user_id FROM users
        WHERE team_id = $1
          AND is_active = true
          AND user_id != $2
          AND user_id != $3
          AND user_id NOT IN (
              SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $4
          )
        LIMIT 1
    `, teamID, authorID, oldUserID, prID).Scan(&newReviewerID)
    if err != nil {
        return "", constants.ErrNoCandidate
    }

	_, err = tx.ExecContext(ctx, `
        UPDATE pr_reviewers SET reviewer_id = $1
        WHERE pr_id = $2 AND reviewer_id = $3
    `, newReviewerID, prID, oldUserID)
    if err != nil {
        return "", err
    }

	return newReviewerID, tx.Commit()
}


func (s *Storage) GetPR(ctx context.Context, prID string) (*models.DBPullRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	var pr models.DBPullRequest

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	err = tx.QueryRowContext(ctx, `
		SELECT 
			pull_request_id, 
			title, 
			author_id, 
			status, 
			created_at, 
			merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1`,
		prID,
	).Scan(
		&pr.PullRequestID,
		&pr.Title, 
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if err != nil {
		if err == constants.ErrNoRows {
			return nil, constants.ErrPRNotFound
		}
		return nil, err
	}

	return &pr, nil
}


