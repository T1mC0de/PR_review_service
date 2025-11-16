package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"pr-review-service/internal/models"
)

func TestCreateTeam_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" { t.Skip("no TEST_DB_DSN") }
	stor, err := NewStorage(dsn)
	if err != nil { t.Fatalf("NewStorage: %v", err) }
	t.Cleanup(func(){ _ = stor.Close() })

	timeNow := time.Now()
	members := []models.DBUser{{UserID:"ct-user-1", Username:"CT1", IsActive:true, CreatedAt: timeNow}}
	cases := []struct{ name, team string; mem []models.DBUser; wantErr bool }{
		{"new-team", "ct-team-1", members, false},
		{"duplicate-team", "ct-team-1", members, true},
	}
	for _, c := range cases {
		 t.Run(c.name, func(t *testing.T){
			 err := stor.CreateTeam(context.Background(), c.team, c.mem)
			 if (err != nil) != c.wantErr { t.Errorf("err=%v wantErr=%v", err, c.wantErr) }
		 })
	}
}

func TestGetTeam_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" { t.Skip("no TEST_DB_DSN") }
	stor, err := NewStorage(dsn)
	if err != nil { t.Fatalf("NewStorage: %v", err) }
	t.Cleanup(func(){ _ = stor.Close() })

	cases := []struct{ name, team string; wantErr bool }{
		{"existing", "ct-team-1", false},
		{"missing", "no-such-team", true},
	}
	for _, c := range cases {
		 t.Run(c.name, func(t *testing.T){
			 _, users, err := stor.GetTeam(context.Background(), c.team)
			 if (err != nil) != c.wantErr { t.Errorf("err=%v wantErr=%v", err, c.wantErr) }
			 if err == nil && len(users) == 0 { t.Log("team exists but no users returned") }
		 })
	}
}
