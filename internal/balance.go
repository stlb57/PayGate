package internal

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

func (h *UserHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value(UserClaimsKey).(*Claims)

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusInternalServerError)
		return
	}

	balance, err := h.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "Unable to fetch balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(BalanceResponse{
		Balance: balance,
	})
}