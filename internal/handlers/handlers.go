package handlers

import (
	"encoding/json"
	"net/http"

	"pr-review-service/internal/storage"
)

// BaseHandler - базовая структура для всех обработчиков
type BaseHandler struct {
	storage *storage.Storage
}

// NewBaseHandler - конструктор базового обработчика
func NewBaseHandler(storage *storage.Storage) *BaseHandler {
	return &BaseHandler{storage: storage}
}

func (h *BaseHandler) sendError(w http.ResponseWriter, statusCode int, code, message string) {
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
	if err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

func (h *BaseHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}
