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

func TestUserSetIsActive(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)
	team := "team_users_" + time.Now().Format("150405.000")
	members := []models.TeamMember{
		{UserID: "ua1", Username: "UA1", IsActive: true},
		{UserID: "ua2", Username: "UA2", IsActive: true},
	}
	b, _ := json.Marshal(models.CreateTeamRequest{TeamName: team, Members: members})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(b)))

	if w.Code != http.StatusCreated {
		t.Fatalf("setup team failed %d", w.Code)
	}

	act, _ := json.Marshal(map[string]interface{}{"user_id": "ua1", "is_active": false})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(
		w2,
		httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(act)),
	)

	if w2.Code != http.StatusOK {
		t.Fatalf("set active want 200 got %d", w2.Code)
	}
}

func TestUserSetIsActiveErrors(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		body string
		want int
	}

	tests := []tc{
		{"{", http.StatusBadRequest},
		{toJSON(map[string]interface{}{"is_active": true}), http.StatusBadRequest},
		{toJSON(map[string]interface{}{"user_id": "x"}), http.StatusBadRequest},
		{
			toJSON(map[string]interface{}{"user_id": "__missing__", "is_active": true}),
			http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(
			w,
			httptest.NewRequest(
				http.MethodPost,
				"/users/setIsActive",
				bytes.NewBufferString(tt.body),
			),
		)

		if w.Code != tt.want {
			t.Fatalf("setIsActive body=%s want %d got %d", tt.body, tt.want, w.Code)
		}
	}
}

func TestUserGetReviewErrors(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		path string
		want int
	}

	tests := []tc{
		{"/users/getReview", http.StatusBadRequest},
		{"/users/getReview?user_id=__missing__", http.StatusNotFound},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if w.Code != tt.want {
			t.Fatalf("getReview path=%s want %d got %d", tt.path, w.Code, tt.want)
		}
	}
}

func TestUserMassDeactivateErrors(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		body string
		want int
	}

	ids := make([]string, 101)
	for i := range ids {
		ids[i] = "x"
	}

	tests := []tc{
		{"{", http.StatusBadRequest},
		{
			toJSON(models.MassDeactivateRequest{TeamName: "", UserIDs: []string{"u1"}}),
			http.StatusBadRequest,
		},
		{
			toJSON(models.MassDeactivateRequest{TeamName: "t", UserIDs: []string{}}),
			http.StatusBadRequest,
		},
		{toJSON(models.MassDeactivateRequest{TeamName: "t", UserIDs: ids}), http.StatusBadRequest},
		{
			toJSON(models.MassDeactivateRequest{TeamName: "__missing__", UserIDs: []string{"u1"}}),
			http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(
			w,
			httptest.NewRequest(
				http.MethodPost,
				"/users/massDeactivate",
				bytes.NewBufferString(tt.body),
			),
		)

		if w.Code != tt.want {
			t.Fatalf("massDeactivate body=%s want %d got %d", tt.body, tt.want, w.Code)
		}
	}
}
