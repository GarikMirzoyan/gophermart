package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GarikMirzoyan/gophermart/internal/domain/balance"
)

type BalancePG struct {
	db *sql.DB
}

func NewBalancePG(db *sql.DB) *BalancePG {
	return &BalancePG{db: db}
}

func (r *BalancePG) GetByUserID(ctx context.Context, userID int) (*balance.Balance, error) {
	var current, withdrawn float64
	err := r.db.QueryRowContext(ctx,
		`SELECT current_balance, total_withdrawn FROM user_balances WHERE user_id = $1`, userID).
		Scan(&current, &withdrawn)
	if err != nil {
		if err == sql.ErrNoRows {
			// Возвращаем 0 баланса, если записи ещё нет
			return &balance.Balance{}, nil
		}
		return nil, err
	}

	return &balance.Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}, nil
}

func (r *BalancePG) Add(ctx context.Context, userID int, amount float64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_balances (user_id, current_balance, total_withdrawn)
		VALUES ($1, $2, 0)
		ON CONFLICT (user_id) DO UPDATE
		SET current_balance = user_balances.current_balance + $2
	`, userID, amount)

	if err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}
	return nil
}
