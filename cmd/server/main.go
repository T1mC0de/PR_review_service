package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"pr-review-service/internal/handlers"
	"pr-review-service/internal/storage"
)

func main() {
	// Небольшая пауза, чтобы Postgres точно поднялся (дополнительно к healthcheck)
	time.Sleep(2 * time.Second)

	// Предпочитаем переменную окружения, чтобы не хардкодить строку
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Println(
			"DATABASE_URL не установлена, используется значение по умолчанию для docker-compose",
		)

		connectionString = "postgres://pr_user:password@postgres:5432/pr_reviewer?sslmode=disable"
	}

	log.Println("Используем строку подключения:", connectionString)

	// Ретраи подключения к БД — помогает при коротких скачках готовности
	var (
		st  *storage.Storage
		err error
	)
	for i := 1; i <= 8; i++ { // до ~16 секунд суммарно
		st, err = storage.NewStorage(connectionString)
		if err == nil {
			break
		}

		log.Printf("Попытка подключения #%d не удалась: %v", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных после ретраев:", err)
	}

	defer st.Close()

	router := handlers.SetupTeamRoutes(st)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
