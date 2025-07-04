package loyalty

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	GetAccrual(ctx context.Context, orderNumber string) (*OrderAccrual, error)
}

type httpClient struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) Client {
	return &httpClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *httpClient) GetAccrual(ctx context.Context, orderNumber string) (*OrderAccrual, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/orders/"+orderNumber, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accrual service error: %d", resp.StatusCode)
	}

	var info OrderAccrual
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
