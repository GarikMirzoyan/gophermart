package balance

import "context"

type Repository interface {
	GetByUserID(ctx context.Context, userID int) (*Balance, error)
	Add(ctx context.Context, userID int, amount int64) error
}
