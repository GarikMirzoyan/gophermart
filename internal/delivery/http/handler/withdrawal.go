package handler

import (
	"encoding/json"
	"net/http"

	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/middleware"
	"github.com/GarikMirzoyan/gophermart/internal/domain/withdrawal"
	withdrawalService "github.com/GarikMirzoyan/gophermart/internal/usecase/withdrawal"
)

type WithdrawalHandler struct {
	WithdrawService *withdrawalService.Service
}

func NewWithdrawalHandler(service *withdrawalService.Service) *WithdrawalHandler {
	return &WithdrawalHandler{WithdrawService: service}
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *WithdrawalHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := h.WithdrawService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
	case withdrawal.ErrInvalidOrderNumber:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case withdrawal.ErrInsufficientFunds:
		http.Error(w, err.Error(), http.StatusPaymentRequired)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WithdrawalHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.WithdrawService.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(withdrawals)
}
