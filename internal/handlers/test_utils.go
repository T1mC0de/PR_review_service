package handlers

import (
	"encoding/json"
	"testing"

	"pr-review-service/internal/storage"
)

const testDSN = "postgres://pr_user:password@localhost:5432/pr_reviewer?sslmode=disable"

func newTestStorage(t *testing.T) *storage.Storage {
	stor, err := storage.NewStorage(testDSN)
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	return stor
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
