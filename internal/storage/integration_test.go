package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"pr-review-service/internal/models"
)

const testDSN = "postgres://pr_user:password@localhost:5432/pr_reviewer?sslmode=disable"

func ensureStorage(t *testing.T) *Storage {
	stor, err := NewStorage(testDSN)
	if err != nil {
		t.Fatalf("cannot connect real postgres: %v", err)
	}

	return stor
}

func uniqueName(prefix string) string { return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano()) }

func TestTeamAndUsersIntegration(t *testing.T) {
	stor := ensureStorage(t)
	defer stor.Close()

	ctx := context.Background()
	team := uniqueName("teamit")
	members := []models.DBUser{
		{UserID: uniqueName("usr"), Username: "UserA", IsActive: true, CreatedAt: time.Now()},
		{UserID: uniqueName("usr"), Username: "UserB", IsActive: true, CreatedAt: time.Now()},
	}

	if err := stor.CreateTeam(ctx, team, members); err != nil {
		t.Fatalf("CreateTeam err: %v", err)
	}

	if err := stor.CreateTeam(ctx, team, members); err == nil {
		t.Fatalf("expected duplicate team error")
	}

	gotTeam, list, err := stor.GetTeam(ctx, team)
	if err != nil {
		t.Fatalf("GetTeam err: %v", err)
	}

	if gotTeam.Name != team {
		t.Fatalf("GetTeam unexpected team name")
	}

	t.Logf("retrieved %d members (userIDs: %v)", len(list), func() []string {
		ids := make([]string, 0, len(list))
		for _, m := range list {
			ids = append(ids, m.UserID)
		}

		return ids
	}())

	if len(list) < 1 {
		t.Fatalf("expected at least 1 member, got %d", len(list))
	}

	if _, err = stor.SetIsActive(ctx, members[0].UserID, false); err != nil {
		t.Fatalf("SetIsActive err: %v", err)
	}

	if _, err = stor.SetIsActive(ctx, uniqueName("missing"), true); err == nil {
		t.Fatalf("expected user not found")
	}
}

func TestPullRequestLifecycleIntegration(t *testing.T) {
	stor := ensureStorage(t)
	defer stor.Close()

	ctx := context.Background()
	team := uniqueName("prteam")
	author := uniqueName("author")
	rev1 := uniqueName("rev")
	rev2 := uniqueName("rev")

	members := []models.DBUser{
		{UserID: author, Username: author, IsActive: true, CreatedAt: time.Now()},
		{UserID: rev1, Username: rev1, IsActive: true, CreatedAt: time.Now()},
		{UserID: rev2, Username: rev2, IsActive: true, CreatedAt: time.Now()},
	}
	if err := stor.CreateTeam(ctx, team, members); err != nil {
		t.Fatalf("CreateTeam err: %v", err)
	}

	prID := uniqueName("pr")
	if err := stor.CreatePullRequest(ctx, prID, "My PR", author); err != nil {
		t.Fatalf("CreatePullRequest err: %v", err)
	}

	if err := stor.CreatePullRequest(ctx, prID, "My PR", author); err == nil {
		t.Fatalf("expected duplicate PR error")
	}

	if err := stor.MergePullRequest(ctx, prID); err != nil {
		t.Fatalf("MergePullRequest err: %v", err)
	}

	if err := stor.MergePullRequest(ctx, prID); err != nil {
		t.Fatalf("MergePullRequest second err: %v", err)
	}

	pr, err := stor.GetPR(ctx, prID)
	if err != nil || pr.PullRequestID != prID {
		t.Fatalf("GetPR err=%v", err)
	}
}

func TestMassDeactivateAndStatsIntegration(t *testing.T) {
	stor := ensureStorage(t)
	defer stor.Close()

	ctx := context.Background()
	team := uniqueName("mass")

	users := []models.DBUser{
		{UserID: uniqueName("u"), Username: "A", IsActive: true, CreatedAt: time.Now()},
		{UserID: uniqueName("u"), Username: "B", IsActive: true, CreatedAt: time.Now()},
		{UserID: uniqueName("u"), Username: "C", IsActive: true, CreatedAt: time.Now()},
	}
	if err := stor.CreateTeam(ctx, team, users); err != nil {
		t.Fatalf("CreateTeam err: %v", err)
	}

	prID := uniqueName("mpr")
	if err := stor.CreatePullRequest(ctx, prID, "Title", users[0].UserID); err != nil {
		t.Fatalf("CreatePullRequest err: %v", err)
	}

	resp, err := stor.MassDeactivateUsers(ctx, team, []string{users[0].UserID, users[1].UserID})
	if err != nil {
		t.Fatalf("MassDeactivateUsers err: %v", err)
	}

	if len(resp.DeactivatedUsers) < 1 {
		t.Fatalf("expected at least 1 deactivated user")
	}

	stats, err := stor.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats err: %v", err)
	}

	if stats.TotalPRs < 1 {
		t.Fatalf("expected at least 1 PR in stats")
	}
}

func TestErrorsIntegration_SimplePaths(t *testing.T) {
	stor := ensureStorage(t)
	defer stor.Close()

	ctx := context.Background()
	if _, err := stor.ReassignPullRequest(ctx, "missing-pr", "x"); err == nil {
		t.Fatalf("expected PR_NOT_FOUND error")
	}

	if err := stor.MergePullRequest(ctx, "missing-pr"); err == nil {
		t.Fatalf("expected PR_NOT_FOUND on merge missing")
	}

	if _, err := stor.GetPR(ctx, "missing-pr"); err == nil {
		t.Fatalf("expected PR_NOT_FOUND on get")
	}

	if _, _, err := stor.GetTeam(ctx, "missing-team"); err == nil {
		t.Fatalf("expected TEAM_NOT_FOUND on get team")
	}
}

func TestMassDeactivateErrorsIntegration(t *testing.T) {
	stor := ensureStorage(t)
	defer stor.Close()

	ctx := context.Background()
	if _, err := stor.MassDeactivateUsers(ctx, "missing-team", []string{"x"}); err == nil {
		t.Fatalf("expected TEAM_NOT_FOUND on mass deactivate")
	}
}
