package order_test

import (
	"context"
	"testing"

	"github.com/GarikMirzoyan/gophermart/internal/domain/order"
	"github.com/GarikMirzoyan/gophermart/internal/loyalty"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/balance"
	orderUC "github.com/GarikMirzoyan/gophermart/internal/usecase/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	orderrepomocks "github.com/GarikMirzoyan/gophermart/internal/domain/order/mocks"
	loyaltymocks "github.com/GarikMirzoyan/gophermart/internal/loyalty/mocks"
	balancemocks "github.com/GarikMirzoyan/gophermart/internal/usecase/balance/mocks"
)

// ===== MOCKS =====

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetOrderOwner(ctx context.Context, number string) (int, error) {
	args := m.Called(ctx, number)
	return args.Int(0), args.Error(1)
}

func (m *MockRepo) AddOrder(ctx context.Context, o *order.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

func (m *MockRepo) GetOrdersByUser(ctx context.Context, userID int) ([]*order.Order, error) {
	return nil, nil
}

func (m *MockRepo) GetOrdersForProcessing(ctx context.Context) ([]*order.Order, error) {
	return nil, nil
}

func (m *MockRepo) UpdateAccrual(ctx context.Context, number string, status string, accrual float64) error {
	return nil
}

func (m *MockRepo) UpdateStatus(ctx context.Context, number string, status string) error {
	return nil
}

// ===== TEST =====

func TestAddOrder(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	loyaltySvc := &loyalty.Service{} // заглушка, не используется здесь
	balanceSvc := &balance.Service{} // заглушка, не используется здесь
	service := orderUC.New(mockRepo, loyaltySvc, balanceSvc)

	t.Run("invalid number format", func(t *testing.T) {
		err := service.AddOrder(ctx, 1, "abc123")
		assert.ErrorIs(t, err, orderUC.ErrInvalidOrderNumber)
	})

	t.Run("invalid Luhn number", func(t *testing.T) {
		err := service.AddOrder(ctx, 1, "1234567890")
		assert.ErrorIs(t, err, orderUC.ErrInvalidOrderNumber)
	})

	t.Run("order already exists for same user", func(t *testing.T) {
		mockRepo.On("GetOrderOwner", ctx, "79927398713").Return(1, nil).Once()

		err := service.AddOrder(ctx, 1, "79927398713") // correct Luhn
		assert.ErrorIs(t, err, orderUC.ErrOrderAlreadyExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("order belongs to another user", func(t *testing.T) {
		mockRepo.On("GetOrderOwner", ctx, "79927398713").Return(2, nil).Once()

		err := service.AddOrder(ctx, 1, "79927398713")
		assert.ErrorIs(t, err, orderUC.ErrOrderBelongsToAnotherUser)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successfully adds order", func(t *testing.T) {
		orderNumber := "79927398713"
		mockRepo.On("GetOrderOwner", ctx, orderNumber).Return(0, nil).Once()
		mockRepo.On("AddOrder", ctx, mock.MatchedBy(func(o *order.Order) bool {
			return o.Number == orderNumber && o.UserID == 1
		})).Return(nil).Once()

		err := service.AddOrder(ctx, 1, orderNumber)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetOrdersByUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(orderrepomocks.Repository)
	service := orderUC.New(mockRepo, nil, nil)

	expected := []*order.Order{
		{Number: "123", Status: "NEW", UserID: 1},
	}
	mockRepo.On("GetOrdersByUser", ctx, 1).Return(expected, nil)

	orders, err := service.GetOrdersByUser(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, orders)
	mockRepo.AssertExpectations(t)
}

func TestProcessPendingOrders_ProcessedOrder(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(orderrepomocks.Repository)
	mockBalance := new(balancemocks.IService)
	mockLoyaltyClient := new(loyaltymocks.Client)

	loyaltySvc := loyalty.New(mockLoyaltyClient)
	orderSvc := orderUC.New(mockRepo, loyaltySvc, mockBalance)

	orders := []*order.Order{
		{Number: "12345678903", UserID: 1, Status: order.StatusNew},
	}

	accrualVal := 42.5
	accrual := &loyalty.OrderAccrual{
		Order:   "12345678903",
		Status:  loyalty.StatusProcessed,
		Accrual: &accrualVal,
	}

	mockRepo.On("GetOrdersForProcessing", mock.Anything).Return(orders, nil)
	mockLoyaltyClient.On("GetAccrual", mock.Anything, "12345678903").Return(accrual, nil)
	mockRepo.On("UpdateAccrual", mock.Anything, "12345678903", string(loyalty.StatusProcessed), accrualVal).Return(nil)
	mockBalance.On("AddBalance", mock.Anything, 1, accrualVal).Return(nil)

	orderSvc.ProcessPendingOrders(ctx)

	mockRepo.AssertExpectations(t)
	mockBalance.AssertExpectations(t)
	mockLoyaltyClient.AssertExpectations(t)

	require.True(t, true)
}
