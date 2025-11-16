package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlersMethodValidation(t *testing.T) {
	stor := newTestStorage(t)
	defer stor.Close()

	r := SetupTeamRoutes(stor)

	type tc struct {
		method, path string
		want         int
	}

	tests := []tc{
		{http.MethodGet, "/team/add", http.StatusMethodNotAllowed},
		{http.MethodGet, "/users/setIsActive", http.StatusMethodNotAllowed},
		{http.MethodGet, "/users/massDeactivate", http.StatusMethodNotAllowed},
		{http.MethodGet, "/pullRequest/create", http.StatusMethodNotAllowed},
		{http.MethodGet, "/pullRequest/reassign", http.StatusMethodNotAllowed},
		{http.MethodGet, "/pullRequest/merge", http.StatusMethodNotAllowed},
		{http.MethodPost, "/team/get?team_name=x", http.StatusMethodNotAllowed},
		{http.MethodPost, "/users/getReview?user_id=x", http.StatusMethodNotAllowed},
		{http.MethodPost, "/stats/get", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))

		if w.Code != tt.want {
			t.Fatalf("%s %s want %d got %d", tt.method, tt.path, tt.want, w.Code)
		}
	}
}
