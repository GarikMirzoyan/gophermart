package order

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/GarikMirzoyan/gophermart/internal/domain/order"
	"github.com/GarikMirzoyan/gophermart/internal/loyalty"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/balance"
)

var ErrInvalidOrderNumber = errors.New("invalid order number")
var ErrOrderAlreadyExists = errors.New("order already exists")
var ErrOrderBelongsToAnotherUser = errors.New("order belongs to another user")

type Service struct {
	repo           order.Repository
	loyaltyService *loyalty.Service
	balanceService *balance.Service
}

func New(repo order.Repository, loyaltyService *loyalty.Service, balanceService *balance.Service) *Service {
	return &Service{repo: repo, loyaltyService: loyaltyService, balanceService: balanceService}
}

// Луна для проверки номера заказа (цифры произвольной длины)
func ValidateLuhn(number string) bool {
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		n := int(number[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}

func (s *Service) AddOrder(ctx context.Context, userID int, number string) error {
	// Проверка номера
	matched, _ := regexp.MatchString(`^\d+$`, number)
	if !matched || !ValidateLuhn(number) {
		return ErrInvalidOrderNumber
	}

	// Проверка владельца заказа
	ownerID, err := s.repo.GetOrderOwner(ctx, number)
	if err != nil {
		return err
	}
	if ownerID != 0 {
		if ownerID == userID {
			return ErrOrderAlreadyExists
		}
		return ErrOrderBelongsToAnotherUser
	}

	order := &order.Order{
		Number:     number,
		Status:     order.StatusNew,
		UploadedAt: time.Now(),
		UserID:     userID,
	}
	if err := s.repo.AddOrder(ctx, order); err != nil {
		return err
	}

	log.Printf("order %s accepted for processing", number)

	return nil
}

func (s *Service) GetOrdersByUser(ctx context.Context, userID int) ([]*order.Order, error) {
	return s.repo.GetOrdersByUser(ctx, userID)
}

func (s *Service) ProcessPendingOrders(ctx context.Context) {
	orders, err := s.repo.GetOrdersForProcessing(ctx)
	if err != nil {
		log.Printf("failed to fetch orders for processing: %v", err)
		return
	}

	for _, o := range orders {
		accrual, err := s.loyaltyService.GetOrderAccrual(ctx, o.Number)
		if err != nil {
			log.Printf("failed to get accrual for order %s: %v", o.Number, err)
			continue
		}
		if accrual == nil {
			continue // Нет данных — пропускаем
		}

		if accrual.Status == loyalty.StatusProcessed && accrual.Accrual != nil {
			err = s.repo.UpdateAccrual(ctx, o.Number, string(accrual.Status), *accrual.Accrual)
			if err != nil {
				log.Printf("failed to update accrual for order %s: %v", o.Number, err)
				continue
			}
			err = s.balanceService.AddBalance(ctx, o.UserID, *accrual.Accrual)
			if err != nil {
				log.Printf("failed to update balance for user %d: %v", o.UserID, err)
			}
		} else {
			err = s.repo.UpdateStatus(ctx, o.Number, string(accrual.Status))
			if err != nil {
				log.Printf("failed to update status for order %s: %v", o.Number, err)
			}
		}
	}
}
