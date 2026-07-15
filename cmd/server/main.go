package main

import (
	"log"
	"net/http"
	"paygate/internal/handlers"
	"paygate/internal/database"
)

func main() {
	databaseURL := "postgres://postgres:YOUR_PASSWORD@localhost:5432/paygate"
	pool, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/users", handlers.CreateUserHandler)

	log.Println("Starting PayGate on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
