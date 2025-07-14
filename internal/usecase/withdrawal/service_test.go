package withdrawal_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GarikMirzoyan/gophermart/internal/domain/withdrawal"
	mock_withdrawal "github.com/GarikMirzoyan/gophermart/internal/domain/withdrawal/mocks"
	usecase "github.com/GarikMirzoyan/gophermart/internal/usecase/withdrawal"

	"go.uber.org/mock/gomock"
)

func TestService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_withdrawal.NewMockRepository(ctrl)

	validOrder := "79927398713"
	invalidOrder := "123"

	validate := func(s string) bool {
		return s == validOrder
	}

	service := usecase.New(mockRepo, validate)
	ctx := context.Background()

	t.Run("invalid order number", func(t *testing.T) {
		err := service.Withdraw(ctx, 1, invalidOrder, 10)
		if !errors.Is(err, withdrawal.ErrInvalidOrderNumber) {
			t.Errorf("expected ErrInvalidOrderNumber, got %v", err)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		mockRepo.EXPECT().
			Withdraw(ctx, 1, validOrder, 10.0).
			Return(withdrawal.ErrInsufficientFunds)

		err := service.Withdraw(ctx, 1, validOrder, 10)
		if !errors.Is(err, withdrawal.ErrInsufficientFunds) {
			t.Errorf("expected ErrInsufficientFunds, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			Withdraw(ctx, 1, validOrder, 10.0).
			Return(errors.New("some db error"))

		err := service.Withdraw(ctx, 1, validOrder, 10)
		if !errors.Is(err, withdrawal.ErrWithdrawSaveFailed) {
			t.Errorf("expected ErrWithdrawSaveFailed, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			Withdraw(ctx, 1, validOrder, 10.0).
			Return(nil)

		err := service.Withdraw(ctx, 1, validOrder, 10)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestService_GetUserWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_withdrawal.NewMockRepository(ctrl)
	service := usecase.New(mockRepo, func(string) bool { return true })
	ctx := context.Background()

	expected := []*withdrawal.Withdrawal{
		{UserID: 1, Order: "123", Sum: 10.0},
	}

	mockRepo.EXPECT().
		GetUserWithdrawals(ctx, 1).
		Return(expected, nil)

	actual, err := service.GetUserWithdrawals(ctx, 1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(actual) != len(expected) || actual[0].Order != expected[0].Order {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}
