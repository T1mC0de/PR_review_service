package storage

import (
	"context"
	"os"
	"testing"
)

func TestSetIsActive_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" { t.Skip("no TEST_DB_DSN") }
	stor, err := NewStorage(dsn)
	if err != nil { t.Fatalf("NewStorage: %v", err) }
	t.Cleanup(func(){ _ = stor.Close() })

	tests := []struct {
		name     string
		userID   string
		isActive bool
		wantErr  bool
	}{
		{"activate-existing", "user-a1", true, false},
		{"deactivate-nonexistent", "no-user-xyz", false, true},
	}
	for _, tc := range tests {
		 t.Run(tc.name, func(t *testing.T){
			 ctx := context.Background()
			 _, err := stor.SetIsActive(ctx, tc.userID, tc.isActive)
			 if (err != nil) != tc.wantErr {
				 t.Errorf("err=%v wantErr=%v", err, tc.wantErr)
			 }
		 })
	}
}

func TestGetReview_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" { t.Skip("no TEST_DB_DSN") }
	stor, err := NewStorage(dsn)
	if err != nil { t.Fatalf("NewStorage: %v", err) }
	t.Cleanup(func(){ _ = stor.Close() })

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"existing-user", "user-a1", false},
		{"missing-user", "missing-123", true},
	}
	for _, tc := range tests {
		 t.Run(tc.name, func(t *testing.T){
			 _, err := stor.GetReview(context.Background(), tc.userID)
			 if (err != nil) != tc.wantErr {
				 t.Errorf("err=%v wantErr=%v", err, tc.wantErr)
			 }
		 })
	}
}

func TestMassDeactivateUsers_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" { t.Skip("no TEST_DB_DSN") }
	stor, err := NewStorage(dsn)
	if err != nil { t.Fatalf("NewStorage: %v", err) }
	t.Cleanup(func(){ _ = stor.Close() })

	tests := []struct {
		name     string
		team     string
		users    []string
		wantErr  bool
	}{
		{"empty-set", "big-team", []string{}, false},
		{"some-users", "big-team", []string{"user-a1","user-a2"}, false},
		{"team-missing", "no-team-x", []string{"user-a1"}, true},
	}
	for _, tc := range tests {
		 t.Run(tc.name, func(t *testing.T){
			 _, err := stor.MassDeactivateUsers(context.Background(), tc.team, tc.users)
			 if (err != nil) != tc.wantErr {
				 t.Errorf("err=%v wantErr=%v", err, tc.wantErr)
			 }
		 })
	}
}
