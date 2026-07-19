package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("my-secret")

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || len(req.Password) < 8 {
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

	user, err := h.FindUserByEmail(
		context.Background(),
		req.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)

	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(24 * time.Hour),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Couldn't generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		Token: tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) FindUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {

	query := `
	SELECT id, name, email, password_hash
	FROM users
	WHERE email = $1
	`

	row := h.pool.QueryRow(
		ctx,
		query,
		email,
	)

	var user User

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}