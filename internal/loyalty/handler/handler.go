package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GarikMirzoyan/gophermart/internal/loyalty"
)

type LoyaltyHandler struct {
	LoyaltyService *loyalty.Service
}

func NewLoyaltyHandler(loyaltyService *loyalty.Service) *LoyaltyHandler {
	return &LoyaltyHandler{}
}

func (h *LoyaltyHandler) GetOrderAccrual(w http.ResponseWriter, r *http.Request) {
	number := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/orders/"))
	if number == "" {
		http.Error(w, "order number required", http.StatusBadRequest)
		return
	}

	accrual, err := h.LoyaltyService.GetOrderAccrual(r.Context(), number)
	if err != nil {
		http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if accrual == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Можно проверить принадлежность заказа пользователю, если нужно:
	// если accrual.UserID != userID { http.Error(w, "forbidden", http.StatusForbidden); return }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accrual)
}
