package handlers

import (
	"encoding/json"
	"net/http"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if request.Name == "" || request.Email == "" {
	http.Error(w, "name and email are required", http.StatusBadRequest)
	return
}

if len(request.Password) < 8 {
	http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
	return
}

	response := CreateUserResponse{
		Name:  request.Name,
		Email: request.Email,
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}