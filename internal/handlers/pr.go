package handlers

import (

	"pr-review-service/internal/storage"
)

type pullRequestHandler struct {
	*BaseHandler
}

func NewPullRequestHandler(storage *storage.Storage) *pullRequestHandler {
	return &pullRequestHandler{
		BaseHandler: NewBaseHandler(storage),
	}
}

func (h *pullRequestHandler) CreatePR() {
}



