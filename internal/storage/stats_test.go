package storage

import (
	"context"
	"os"
	"testing"
)

func TestGetStats_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("no TEST_DB_DSN")
	}

	stor, err := NewStorage(dsn)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	t.Cleanup(func() { _ = stor.Close() })

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"stats-ok", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := stor.GetStats(context.Background())
			if (err != nil) != tc.wantErr {
				t.Errorf("err=%v wantErr=%v", err, tc.wantErr)
			}

			if err == nil && s.TotalPRs < 0 {
				t.Errorf("invalid total prs")
			}
		})
	}
}
