package constants

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTeamExists       = errors.New("TEAM_EXISTS")
	ErrTeamNotFound     = errors.New("TEAM_NOT_FOUND")
	ErrUserTeamNotFound = errors.New("USER_TEAM_NOT_FOUND")
	ErrUserNotFound     = errors.New("USER_NOT_FOUND")
	ErrPRExists         = errors.New("PR_EXISTS")
	ErrPRNotFound       = errors.New("PR_NOT_FOUND")
	ErrPRMerged         = errors.New("PR_MERGED")
	ErrNotAssigned      = errors.New("NOT_ASSIGNED")
	ErrNoCandidate      = errors.New("NO_CANDIDATE")
	ErrNotFound         = errors.New("NOT_FOUND")
	ErrNoRows           = sql.ErrNoRows
)

const (
	StorageTimeout = 5 * time.Second
)
