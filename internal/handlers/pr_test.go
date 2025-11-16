package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPRCreateDuplicate(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	team := "team_pr_" + time.Now().Format("150405.000")
	teamPayload := `{"team_name":"` + team + `","members":[{"user_id":"a","username":"A","is_active":true},{"user_id":"b","username":"B","is_active":true}]}`
	w := httptest.NewRecorder()
	r.ServeHTTP(
		w,
		httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(teamPayload)),
	)

	if w.Code != http.StatusCreated {
		t.Fatalf("team create want 201 got %d", w.Code)
	}

	prID := fmt.Sprintf("pr_%d", time.Now().UnixNano())
	prPayload := `{"pull_request_id":"` + prID + `","pull_request_name":"N","author_id":"a"}`
	w2 := httptest.NewRecorder()
	r.ServeHTTP(
		w2,
		httptest.NewRequest(
			http.MethodPost,
			"/pullRequest/create",
			bytes.NewBufferString(prPayload),
		),
	)

	if w2.Code != http.StatusCreated {
		t.Fatalf("pr create want 201 got %d", w2.Code)
	}

	wDup := httptest.NewRecorder()
	r.ServeHTTP(
		wDup,
		httptest.NewRequest(
			http.MethodPost,
			"/pullRequest/create",
			bytes.NewBufferString(prPayload),
		),
	)

	if wDup.Code != http.StatusBadRequest {
		t.Fatalf("duplicate want 400 got %d", wDup.Code)
	}
}

func TestPRCreateValidation(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		body string
		want int
	}

	tests := []tc{
		{"{", http.StatusBadRequest},
		{`{"pull_request_name":"x","author_id":"a"}`, http.StatusBadRequest},
		{`{"pull_request_id":"x","author_id":"a"}`, http.StatusBadRequest},
		{`{"pull_request_id":"x","pull_request_name":"n"}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(
			w,
			httptest.NewRequest(
				http.MethodPost,
				"/pullRequest/create",
				bytes.NewBufferString(tt.body),
			),
		)

		if w.Code != tt.want {
			t.Fatalf("createPR body=%s want %d got %d", tt.body, tt.want, w.Code)
		}
	}
}

func TestPRReassignMerge(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	team := "team_pr2_" + time.Now().Format("150405.000")
	teamPayload := `{"team_name":"` + team + `","members":[{"user_id":"a","username":"A","is_active":true},{"user_id":"b","username":"B","is_active":true},{"user_id":"c","username":"C","is_active":true}]}`
	w := httptest.NewRecorder()
	r.ServeHTTP(
		w,
		httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(teamPayload)),
	)

	if w.Code != http.StatusCreated {
		t.Fatalf("team create want 201 got %d", w.Code)
	}

	prID := fmt.Sprintf("pr2_%d", time.Now().UnixNano())
	prPayload := `{"pull_request_id":"` + prID + `","pull_request_name":"R","author_id":"a"}`
	w2 := httptest.NewRecorder()
	r.ServeHTTP(
		w2,
		httptest.NewRequest(
			http.MethodPost,
			"/pullRequest/create",
			bytes.NewBufferString(prPayload),
		),
	)

	if w2.Code != http.StatusCreated {
		t.Fatalf("pr create want 201 got %d", w2.Code)
	}

	reassign := `{"pull_request_id":"` + prID + `","old_user_id":"b"}`
	wr := httptest.NewRecorder()
	r.ServeHTTP(
		wr,
		httptest.NewRequest(
			http.MethodPost,
			"/pullRequest/reassign",
			bytes.NewBufferString(reassign),
		),
	)

	if wr.Code != http.StatusOK && wr.Code != http.StatusBadRequest {
		t.Fatalf("reassign want 200|400 got %d", wr.Code)
	}

	merge := `{"pull_request_id":"` + prID + `"}`
	wm := httptest.NewRecorder()
	r.ServeHTTP(
		wm,
		httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(merge)),
	)

	if wm.Code != http.StatusOK {
		t.Fatalf("merge want 200 got %d", wm.Code)
	}

	wm2 := httptest.NewRecorder()
	r.ServeHTTP(
		wm2,
		httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(merge)),
	)

	if wm2.Code != http.StatusOK {
		t.Fatalf("idempotent merge want 200 got %d", wm2.Code)
	}

	wr2 := httptest.NewRecorder()
	r.ServeHTTP(
		wr2,
		httptest.NewRequest(
			http.MethodPost,
			"/pullRequest/reassign",
			bytes.NewBufferString(reassign),
		),
	)

	if wr2.Code != http.StatusBadRequest {
		t.Fatalf("reassign merged want 400 got %d", wr2.Code)
	}
}

func TestPRReassignErrors(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		body string
		want int
	}

	tests := []tc{
		{"{", http.StatusBadRequest},
		{`{"old_user_id":"x"}`, http.StatusBadRequest},
		{`{"pull_request_id":"x"}`, http.StatusBadRequest},
		{`{"pull_request_id":"missing","old_user_id":"x"}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(
			w,
			httptest.NewRequest(
				http.MethodPost,
				"/pullRequest/reassign",
				bytes.NewBufferString(tt.body),
			),
		)

		if w.Code != tt.want {
			t.Fatalf("reassign body=%s want %d got %d", tt.body, tt.want, w.Code)
		}
	}
}
