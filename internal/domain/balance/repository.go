package balance

import "context"

type Repository interface {
	GetByUserID(ctx context.Context, userID int) (*Balance, error)
	Add(ctx context.Context, userID int, amount float64) error
}
