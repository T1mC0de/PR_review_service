package storage

import (
	"database/sql"

	_ "github.com/lib/pq" // nolint:gosec
)

type Storage struct {
	db *sql.DB
}

func NewStorage(connectionString string) (*Storage, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) GetDB() *sql.DB {
	return s.db
}
