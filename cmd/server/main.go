package main

import (
	"log"
	"net/http"

	"pr-review-service/internal/handlers"
	"pr-review-service/internal/storage"
)

func main() {
	// Connection string для PostgreSQL
	connectionString := "postgres://pr_user:password@localhost:5432/pr_reviewer?sslmode=disable"

	// Для Docker используйте:
	// connectionString := "postgres://pr_user:password@localhost:5432/pr_reviewer?sslmode=disable"

	storage, err := storage.NewStorage(connectionString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer storage.Close()

	router := handlers.SetupTeamRoutes(storage)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
