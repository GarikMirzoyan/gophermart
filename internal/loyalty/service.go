package loyalty

import (
	"context"
)

type Service struct {
	client Client
}

func New(client Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetOrderAccrual(ctx context.Context, number string) (*OrderAccrual, error) {
	return s.client.GetAccrual(ctx, number)
}
