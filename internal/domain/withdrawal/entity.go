package withdrawal

import "time"

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	UserID      int       `json:"-"`
	ProcessedAt time.Time `json:"processed_at"`
}
