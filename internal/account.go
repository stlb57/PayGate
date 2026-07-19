package internal

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type Account struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Balance   int64     `json:"balance"`
	CreatedAt string    `json:"created_at"`
}

func (h *UserHandler) GetBalance(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {

	var balance int64

	err := h.pool.QueryRow(
		ctx,
		`
		SELECT balance
		FROM accounts
		WHERE user_id = $1
		`,
		userID,
	).Scan(&balance)

	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (h *UserHandler) Deposit(
	ctx context.Context,
	userID uuid.UUID,
	amount int64,
) error {

	_, err := h.pool.Exec(
		ctx,
		`
		UPDATE accounts
		SET balance = balance + $1
		WHERE user_id = $2
		`,
		amount,
		userID,
	)

	return err
}

func (h *UserHandler) Withdraw(
	ctx context.Context,
	userID uuid.UUID,
	amount int64,
) error {

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var balance int64

	err = tx.QueryRow(
		ctx,
		`
		SELECT balance
		FROM accounts
		WHERE user_id = $1
		FOR UPDATE
		`,
		userID,
	).Scan(&balance)

	if err != nil {
		return err
	}

	if balance < amount {
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
		userID,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (h *UserHandler) Transfer(
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

	err = tx.QueryRow(
		ctx,
		`
		SELECT balance
		FROM accounts
		WHERE user_id = $1
		FOR UPDATE
		`,
		senderID,
	).Scan(&senderBalance)

	if err != nil {
		return err
	}

	if senderBalance < amount {
		return ErrInsufficientBalance
	}

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