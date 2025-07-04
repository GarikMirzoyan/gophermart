package balance

import (
	"context"

	"github.com/GarikMirzoyan/gophermart/internal/domain/balance"
)

type IService interface {
	GetBalance(ctx context.Context, userID int) (*balance.Balance, error)
	AddBalance(ctx context.Context, userID int, amount float64) error
}

type Service struct {
	repo balance.Repository
}

func New(repo balance.Repository) IService {
	return &Service{repo: repo}
}

func (s *Service) GetBalance(ctx context.Context, userID int) (*balance.Balance, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) AddBalance(ctx context.Context, userID int, amount float64) error {
	return s.repo.Add(ctx, userID, amount)
}
