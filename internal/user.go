package internal

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	pool *pgxpool.Pool
}

func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		pool: pool,
	}
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

	id := uuid.New()

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		http.Error(w, "Password Hash Failed", http.StatusInternalServerError)
		return
	}

	err = h.CreateUserInDB(
		context.Background(),
		id.String(),
		req.Name,
		req.Email,
		string(hashedPassword),
	)
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	resp := CreateUserResponse{
		ID:    id.String(),
		Name:  req.Name,
		Email: req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) CreateUserInDB(
	ctx context.Context,
	id string,
	name string,
	email string,
	passwordHash string,
) error {

	query := `
	INSERT INTO users (id, name, email, password_hash)
	VALUES ($1, $2, $3, $4)
	`

	_, err := h.pool.Exec(
		ctx,
		query,
		id,
		name,
		email,
		passwordHash,
	)

	return err
}