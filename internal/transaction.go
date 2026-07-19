package internal

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID         uuid.UUID `json:"id"`
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Amount     int64     `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *UserHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {

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

	rows, err := h.pool.Query(
		r.Context(),
		`
		SELECT id, sender_id, receiver_id, amount, created_at
		FROM transactions
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY created_at DESC
		`,
		userID,
	)
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var t Transaction

		if err := rows.Scan(
			&t.ID,
			&t.SenderID,
			&t.ReceiverID,
			&t.Amount,
			&t.CreatedAt,
		); err != nil {
			http.Error(w, "Database Error", http.StatusInternalServerError)
			return
		}

		transactions = append(transactions, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}