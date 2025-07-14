package order_test

import (
	"context"
	"testing"

	"github.com/GarikMirzoyan/gophermart/internal/domain/order"
	"github.com/GarikMirzoyan/gophermart/internal/loyalty"
	orderUC "github.com/GarikMirzoyan/gophermart/internal/usecase/order"

	mock_balance "github.com/GarikMirzoyan/gophermart/internal/delivery/http/handler/balancehandler/mocks"
	mock_order_rep "github.com/GarikMirzoyan/gophermart/internal/domain/order/mocks"
	mock_loyalty "github.com/GarikMirzoyan/gophermart/internal/loyalty/mocks"

	"go.uber.org/mock/gomock"
)

func TestAddOrder(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_order_rep.NewMockRepository(ctrl)
	service := orderUC.New(mockRepo, nil, nil, func(s string) bool {
		return s == "79927398713"
	})

	t.Run("invalid number format", func(t *testing.T) {
		err := service.AddOrder(ctx, 1, "abc123")
		if err != orderUC.ErrInvalidOrderNumber {
			t.Fatalf("expected ErrInvalidOrderNumber, got %v", err)
		}
	})

	t.Run("invalid Luhn number", func(t *testing.T) {
		err := service.AddOrder(ctx, 1, "1234567890")
		if err != orderUC.ErrInvalidOrderNumber {
			t.Fatalf("expected ErrInvalidOrderNumber, got %v", err)
		}
	})

	t.Run("order already exists for same user", func(t *testing.T) {
		mockRepo.EXPECT().
			GetOrderOwner(ctx, "79927398713").
			Return(1, nil)

		err := service.AddOrder(ctx, 1, "79927398713")
		if err != orderUC.ErrOrderAlreadyExists {
			t.Fatalf("expected ErrOrderAlreadyExists, got %v", err)
		}
	})

	t.Run("order belongs to another user", func(t *testing.T) {
		mockRepo.EXPECT().
			GetOrderOwner(ctx, "79927398713").
			Return(2, nil)

		err := service.AddOrder(ctx, 1, "79927398713")
		if err != orderUC.ErrOrderBelongsToAnotherUser {
			t.Fatalf("expected ErrOrderBelongsToAnotherUser, got %v", err)
		}
	})

	t.Run("successfully adds order", func(t *testing.T) {
		orderNumber := "79927398713"

		mockRepo.EXPECT().
			GetOrderOwner(ctx, orderNumber).
			Return(0, nil)

		mockRepo.EXPECT().
			AddOrder(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, o *order.Order) error {
				if o.Number != orderNumber || o.UserID != 1 {
					t.Fatalf("unexpected order %+v", o)
				}
				return nil
			})

		err := service.AddOrder(ctx, 1, orderNumber)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetOrdersByUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_order_rep.NewMockRepository(ctrl)
	service := orderUC.New(mockRepo, nil, nil, func(s string) bool {
		return s == "12345678903"
	})

	expected := []*order.Order{{Number: "123", Status: "NEW", UserID: 1}}

	mockRepo.EXPECT().
		GetOrdersByUser(ctx, 1).
		Return(expected, nil)

	orders, err := service.GetOrdersByUser(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 1 || orders[0].Number != "123" {
		t.Fatalf("unexpected orders: %+v", orders)
	}
}

func TestProcessPendingOrders_ProcessedOrder(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_order_rep.NewMockRepository(ctrl)
	mockBalance := mock_balance.NewMockService(ctrl)
	mockLoyaltyClient := mock_loyalty.NewMockClient(ctrl)

	loyaltySvc := loyalty.New(mockLoyaltyClient)
	service := orderUC.New(mockRepo, loyaltySvc, mockBalance, func(s string) bool {
		return s == "12345678903"
	})

	orderNumber := "12345678903"
	accrualVal := 42.5
	orders := []*order.Order{
		{Number: orderNumber, UserID: 1, Status: order.StatusNew},
	}
	accrual := &loyalty.OrderAccrual{
		Order:   orderNumber,
		Status:  loyalty.StatusProcessed,
		Accrual: &accrualVal,
	}

	mockRepo.EXPECT().
		GetOrdersForProcessing(ctx).
		Return(orders, nil)

	mockLoyaltyClient.EXPECT().
		GetAccrual(ctx, orderNumber).
		Return(accrual, nil)

	mockRepo.EXPECT().
		UpdateAccrual(ctx, orderNumber, string(loyalty.StatusProcessed), accrualVal).
		Return(nil)

	mockBalance.EXPECT().
		AddBalance(ctx, 1, accrualVal).
		Return(nil)

	service.ProcessPendingOrders(ctx)
}
