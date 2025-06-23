package handler

import (
	"encoding/json"
	"net/http"

	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/middleware"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/balance"
)

type BalanceHandler struct {
	BalanceService *balance.Service
}

func NewBalanceHandler(balanceService *balance.Service) *BalanceHandler {
	return &BalanceHandler{BalanceService: balanceService}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
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
