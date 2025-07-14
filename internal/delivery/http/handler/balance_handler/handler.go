package balance_handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/middleware"
	"github.com/GarikMirzoyan/gophermart/internal/domain/balance"
)

type Service interface {
	GetBalance(ctx context.Context, userID int) (*balance.Balance, error)
	AddBalance(ctx context.Context, userID int, amount float64) error
}

type Handler struct {
	BalanceService Service
}

func New(balanceService Service) *Handler {
	return &Handler{BalanceService: balanceService}
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bal, err := h.BalanceService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   bal.Current,
		Withdrawn: bal.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
