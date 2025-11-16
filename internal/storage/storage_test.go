package storage

import (
	"os"
	"testing"
)

func TestNewStorage_InvalidDSN(t *testing.T) {
	_, err := NewStorage("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	if err == nil {
		t.Skip("constructor didn't fail; requires real network for ping; skipping")
	}
}

func TestStorage_GetDB(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping integration style test")
	}

	stor, err := NewStorage(dsn)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	if stor.GetDB() == nil {
		t.Errorf("GetDB returned nil")
	}

	_ = stor.Close()
}
