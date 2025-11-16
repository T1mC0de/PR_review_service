package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStatsGet(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	team := "team_stats_" + time.Now().Format("150405.000")
	teamPayload := `{"team_name":"` + team + `","members":[{"user_id":"a","username":"A","is_active":true},{"user_id":"b","username":"B","is_active":true}]}`
	w := httptest.NewRecorder()
	r.ServeHTTP(
		w,
		httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(teamPayload)),
	)

	if w.Code != http.StatusCreated {
		t.Fatalf("team create want 201 got %d", w.Code)
	}

	for i := 0; i < 2; i++ {
		pid := fmt.Sprintf("spr_%d_%d", time.Now().UnixNano(), i)
		payload := `{"pull_request_id":"` + pid + `","pull_request_name":"N","author_id":"a"}`
		pw := httptest.NewRecorder()
		r.ServeHTTP(
			pw,
			httptest.NewRequest(
				http.MethodPost,
				"/pullRequest/create",
				bytes.NewBufferString(payload),
			),
		)

		if pw.Code != http.StatusCreated {
			t.Fatalf("create pr want 201 got %d", pw.Code)
		}
	}

	ws := httptest.NewRecorder()
	r.ServeHTTP(ws, httptest.NewRequest(http.MethodGet, "/stats/get", nil))

	if ws.Code != http.StatusOK {
		t.Fatalf("stats want 200 got %d", ws.Code)
	}
}

func TestStatsMethodNotAllowed(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/stats/get", nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("stats post want 405 got %d", w.Code)
	}
}
