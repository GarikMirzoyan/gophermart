package balance_test

import (
	"context"
	"testing"

	"github.com/GarikMirzoyan/gophermart/internal/domain/balance"
	"github.com/GarikMirzoyan/gophermart/internal/domain/balance/mocks"
	usecase "github.com/GarikMirzoyan/gophermart/internal/usecase/balance"

	"go.uber.org/mock/gomock"
)

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := usecase.New(mockRepo)

	ctx := context.Background()
	expected := &balance.Balance{Current: 100, Withdrawn: 50}

	mockRepo.EXPECT().
		GetByUserID(ctx, 1).
		Return(expected, nil)

	result, err := service.GetBalance(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Current != 100 || result.Withdrawn != 50 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestAddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := usecase.New(mockRepo)

	ctx := context.Background()

	mockRepo.EXPECT().
		Add(ctx, 1, 42.0).
		Return(nil)

	err := service.AddBalance(ctx, 1, 42.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
