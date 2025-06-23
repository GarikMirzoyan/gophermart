package withdrawal

import "errors"

var (
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrWithdrawSaveFailed = errors.New("failed to process withdrawal")
)
