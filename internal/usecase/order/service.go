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

	go func() {
		accrual, err := s.loyaltyService.GetOrderAccrual(context.Background(), number)
		if err != nil {
			log.Printf("loyalty service error for order %s: %v", number, err)
			return
		}
		if accrual == nil {
			return
		}

		if accrual.Status == loyalty.StatusProcessed && accrual.Accrual != nil {
			if err := s.repo.UpdateAccrual(context.Background(), number, string(accrual.Status), *accrual.Accrual); err != nil {
				log.Printf("failed to update accrual for order %s: %v", number, err)
			}
			if err := s.balanceService.AddBalance(context.Background(), userID, *accrual.Accrual); err != nil {
				log.Printf("failed to update balance for user %d: %v", userID, err)
			}
		} else {
			if err := s.repo.UpdateStatus(context.Background(), number, string(accrual.Status)); err != nil {
				log.Printf("failed to update status for order %s: %v", number, err)
			}
		}
	}()

	return nil
}

func (s *Service) GetOrdersByUser(ctx context.Context, userID int) ([]*order.Order, error) {
	return s.repo.GetOrdersByUser(ctx, userID)
}
