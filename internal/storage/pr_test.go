package storage

import (
	"context"
	"os"
	"testing"
)

func TestPullRequestLifecycle_Table(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("no TEST_DB_DSN")
	}

	stor, err := NewStorage(dsn)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	t.Cleanup(func() { _ = stor.Close() })

	cases := []struct {
		name      string
		action    string
		prID      string
		prTitle   string
		authorID  string
		oldUserID string
		wantErr   bool
	}{
		{"create-ok", "create", "pr-test-1", "Title1", "ct-user-1", "", false},
		{"create-duplicate", "create", "pr-test-1", "Title1", "ct-user-1", "", true},
		{"merge-ok", "merge", "pr-test-1", "", "", "", false},
		{"reassign-no-user", "reassign", "pr-test-1", "", "", "unknown-x", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()

			switch c.action {
			case "create":
				err := stor.CreatePullRequest(ctx, c.prID, c.prTitle, c.authorID)
				if (err != nil) != c.wantErr {
					t.Errorf("create err=%v wantErr=%v", err, c.wantErr)
				}
			case "merge":
				err := stor.MergePullRequest(ctx, c.prID)
				if (err != nil) != c.wantErr {
					t.Errorf("merge err=%v wantErr=%v", err, c.wantErr)
				}
			case "reassign":
				_, err := stor.ReassignPullRequest(ctx, c.prID, c.oldUserID)
				if (err != nil) != c.wantErr {
					t.Errorf("reassign err=%v wantErr=%v", err, c.wantErr)
				}
			}
		})
	}

	if pr, err := stor.GetPR(context.Background(), "pr-test-1"); err != nil {
		t.Errorf("GetPR unexpected err: %v", err)
	} else if pr.PullRequestID != "pr-test-1" {
		t.Errorf("GetPR wrong id: %s", pr.PullRequestID)
	}
}
