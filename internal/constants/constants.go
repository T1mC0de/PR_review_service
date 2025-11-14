package constants

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTeamExists   = errors.New("TEAM_EXISTS")
	ErrTeamNotFound = errors.New("TEAM_NOT_FOUND")
	ErrPRExists     = errors.New("PR_EXISTS")
	ErrPRMerged     = errors.New("PR_MERGED")
	ErrNotAssigned  = errors.New("NOT_ASSIGNED")
	ErrNoCandidate  = errors.New("NO_CANDIDATE")
	ErrNotFound     = errors.New("NOT_FOUND")
	ErrNoRows       = sql.ErrNoRows
)

const (
	StorageTimeout = 5 * time.Second
)
