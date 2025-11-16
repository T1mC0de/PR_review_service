package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pr-review-service/internal/models"
)

func TestTeamCreateGet(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	team := "team_" + time.Now().Format("150405.000")
	req := models.CreateTeamRequest{
		TeamName: team,
		Members: []models.TeamMember{
			{UserID: "ta", Username: "A", IsActive: true},
			{UserID: "tb", Username: "B", IsActive: true},
		},
	}
	b, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(b)))

	if w.Code != http.StatusCreated {
		t.Fatalf("create want 201 got %d", w.Code)
	}

	wDup := httptest.NewRecorder()
	r.ServeHTTP(wDup, httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(b)))

	if wDup.Code != http.StatusBadRequest {
		t.Fatalf("duplicate want 400 got %d", wDup.Code)
	}

	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, httptest.NewRequest(http.MethodGet, "/team/get?team_name="+team, nil))

	if wGet.Code != http.StatusOK {
		t.Fatalf("get want 200 got %d", wGet.Code)
	}
}

func TestTeamCreateErrors(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		body string
		want int
	}

	tests := []tc{
		{"{", http.StatusBadRequest},
		{
			toJSON(models.CreateTeamRequest{TeamName: "", Members: []models.TeamMember{}}),
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(
			w,
			httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(tt.body)),
		)

		if w.Code != tt.want {
			t.Fatalf("team/add body=%s want %d got %d", tt.body, tt.want, w.Code)
		}
	}
}

func TestTeamGetNotFound(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/team/get?team_name=__missing__", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d", w.Code)
	}
}
