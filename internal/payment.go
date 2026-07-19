package internal

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PaymentRequest struct {
	ReceiverID string `json:"receiver_id"`
	Amount     int64  `json:"amount"`
}

type PaymentResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) MakePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
		return
	}

	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		http.Error(w, "Invalid Receiver ID", http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(UserClaimsKey).(*Claims)

	senderID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid Sender ID", http.StatusInternalServerError)
		return
	}

	err = h.transferMoney(r.Context(), senderID, receiverID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, ErrInsufficientBalance):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "Account not found", http.StatusNotFound)
		default:
			http.Error(w, "Payment Failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(PaymentResponse{
		Message: "Payment Successful",
	})
}

var ErrInsufficientBalance = errors.New("insufficient balance")

func (h *UserHandler) transferMoney(
	ctx context.Context,
	senderID,
	receiverID uuid.UUID,
	amount int64,
) error {

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var senderBalance int64

	var receiverBalance int64

err = tx.QueryRow(
	ctx,
	`
	SELECT balance
	FROM accounts
	WHERE user_id = $1
	FOR UPDATE
	`,
	receiverID,
).Scan(&receiverBalance)

if err != nil {
	return err
}

	if senderBalance < amount {
		return ErrInsufficientBalance
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE accounts
		SET balance = balance - $1
		WHERE user_id = $2
		`,
		amount,
		senderID,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		SELECT balance
		FROM accounts
		WHERE user_id = $1
		FOR UPDATE
		`,
		receiverID,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE accounts
		SET balance = balance + $1
		WHERE user_id = $2
		`,
		amount,
		receiverID,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO transactions
		(id, sender_id, receiver_id, amount)
		VALUES ($1,$2,$3,$4)
		`,
		uuid.New(),
		senderID,
		receiverID,
		amount,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}