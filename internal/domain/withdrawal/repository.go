package withdrawal

import "context"

type Repository interface {
	Withdraw(ctx context.Context, userID int, order string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int) ([]*Withdrawal, error)
	GetTotalWithdrawn(ctx context.Context, userID int) (float64, error)
}
