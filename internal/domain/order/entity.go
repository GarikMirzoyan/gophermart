package order

import "time"

type Status string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

type Order struct {
	Number     string
	Status     Status
	Accrual    *float64
	UploadedAt time.Time
	UserID     int
}
