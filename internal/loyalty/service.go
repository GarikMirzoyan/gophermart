package loyalty

import (
	"context"
	"time"

	"math/rand"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

var statuses = []AccrualStatus{
	StatusRegistered,
	StatusInvalid,
	StatusProcessing,
	StatusProcessed,
}

func (s *Service) GetOrderAccrual(ctx context.Context, number string) (*OrderAccrual, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	statuses := []AccrualStatus{
		StatusRegistered,
		StatusInvalid,
		StatusProcessing,
		StatusProcessed,
	}
	status := statuses[rand.Intn(len(statuses))]

	var accrual *int64
	if status == StatusProcessed {
		val := int64(r.Intn(1000))
		accrual = &val
	}

	return &OrderAccrual{
		Order:   number,
		Status:  status,
		Accrual: accrual,
	}, nil
}
