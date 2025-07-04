package withdrawal

import (
	"context"
	"errors"
	"log"

	"github.com/GarikMirzoyan/gophermart/internal/domain/withdrawal"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/order"
)

type Service struct {
	repo withdrawal.Repository
}

func New(repo withdrawal.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	// Валидация номера заказа
	if !order.ValidateLuhn(orderNumber) {
		return withdrawal.ErrInvalidOrderNumber
	}

	err := s.repo.Withdraw(ctx, userID, orderNumber, sum)
	if err != nil {
		if errors.Is(err, withdrawal.ErrInsufficientFunds) {
			return withdrawal.ErrInsufficientFunds
		}
		log.Printf("Withdraw failed: userID=%d order=%s sum=%.2f error=%v", userID, orderNumber, sum, err)
		return withdrawal.ErrWithdrawSaveFailed
	}

	return nil
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userID int) ([]*withdrawal.Withdrawal, error) {
	return s.repo.GetUserWithdrawals(ctx, userID)
}
