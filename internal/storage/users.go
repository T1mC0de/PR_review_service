package storage

import (
	"context"

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

func (s *Storage) getUserTeamId(ctx context.Context, userID string) (*int, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	var teamId int

	err := s.db.QueryRowContext(ctx, `
		SELECT team_id 
		FROM users 
		WHERE user_id = $1`, userID).Scan(&teamId)
	if err != nil {
		switch err {
		case constants.ErrNoRows:
			return nil, constants.ErrTeamNotFound
		default:
			return nil, err
		}
	}

	return &teamId, nil
}



