package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/GarikMirzoyan/gophermart/internal/domain/withdrawal"
)

type WithdrawalPG struct {
	db *sql.DB
}

func NewWithdrawalPG(db *sql.DB) *WithdrawalPG {
	return &WithdrawalPG{db: db}
}

func (r *WithdrawalPG) Withdraw(ctx context.Context, userID int, order string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверка баланса
	var current float64
	err = tx.QueryRowContext(ctx, `
		SELECT current_balance FROM user_balances WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&current)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			current = 0
		} else {
			return err
		}
	}

	if current < sum {
		return withdrawal.ErrInsufficientFunds
	}

	// Списание
	_, err = tx.ExecContext(ctx, `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, $4)
	`, userID, order, sum, time.Now())
	if err != nil {
		return err
	}

	// Обновление баланса
	_, err = tx.ExecContext(ctx, `
		UPDATE user_balances
		SET current_balance = current_balance - $1, total_withdrawn = total_withdrawn + $1
		WHERE user_id = $2
	`, sum, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *WithdrawalPG) GetUserWithdrawals(ctx context.Context, userID int) ([]*withdrawal.Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*withdrawal.Withdrawal
	for rows.Next() {
		var w withdrawal.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *WithdrawalPG) GetTotalWithdrawn(ctx context.Context, userID int) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1
	`, userID).Scan(&total)
	return total, err
}
