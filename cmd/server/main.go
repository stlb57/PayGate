package main

import (
	"log"
	"net/http"
	"paygate/internal/handlers"
)

func main() {
	http.HandleFunc("/health", handlers.HealthHandler)
	log.Println("Starting PayGate on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
