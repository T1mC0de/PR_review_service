package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"pr-review-service/internal/storage"
)

func SetupTeamRoutes(storage *storage.Storage) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	teamHandler := NewTeamHandler(storage)
	userHandler := NewUserHandler(storage)
	pullRequestHandler := NewPullRequestHandler(storage)

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.CreateTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", pullRequestHandler.CreatePR)
	})

	return r
}
