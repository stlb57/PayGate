package main

import (
	"log"
	"net/http"

	"paygate/internal"
)

const dbURL = "postgres://postgres:password@localhost:5432/paygate"

func main() {
	pool, err := internal.Connect(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	handler := internal.NewUserHandler(pool)

	http.HandleFunc("/health", internal.Health)
	http.HandleFunc("/users", handler.CreateUser)
	http.Handle("/me",AuthMiddleware(http.HandlerFunc(handler.Me)),)
	log.Println("Server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}