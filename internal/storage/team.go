package storage

import (
	"context"

	"pr-review-service/internal/constants"
	"pr-review-service/internal/models"
)

func (s *Storage) CreateTeam(ctx context.Context, teamName string, members []models.DBUser) error {
	ctx, cancel := context.WithTimeout(ctx, constants.StorageTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var exists bool

	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", teamName).
		Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return constants.ErrTeamExists
	}

	var teamID int

	err = tx.QueryRowContext(ctx, "INSERT INTO teams (name) VALUES ($1) RETURNING id", teamName).
		Scan(&teamID)
	if err != nil {
		return err
	}

	for _, member := range members {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO users (user_id, username, is_active, team_id, created_at) 
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (user_id) 
            DO UPDATE SET 
                username = EXCLUDED.username, 
                is_active = EXCLUDED.is_active, 
                team_id = EXCLUDED.team_id,
                updated_at = CURRENT_TIMESTAMP`,
			member.ID, member.Username, member.IsActive, teamID, member.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetTeam(
	ctx context.Context,
	teamName string,
) (*models.DBTeam, []models.DBUser, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	var team models.DBTeam

	err = tx.QueryRowContext(ctx, "SELECT id, name, created_at FROM teams WHERE name = $1", teamName).
		Scan(&team.ID, &team.Name, &team.CreatedAt)
	if err != nil {
		if err == constants.ErrNoRows {
			return nil, nil, constants.ErrTeamNotFound
		}

		return nil, nil, err
	}

	rows, err := tx.QueryContext(ctx, `
        SELECT id, user_id, username, is_active, team_id, created_at 
        FROM users 
        WHERE team_id = $1
        ORDER BY created_at`, team.ID)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	var members []models.DBUser

	for rows.Next() {
		var member models.DBUser

		err := rows.Scan(
			&member.ID,
			&member.UserID,
			&member.Username,
			&member.IsActive,
			&member.TeamID,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return &team, members, tx.Commit()
}
