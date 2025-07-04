package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/middleware"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/order"
)

type OrderHandler struct {
	OrderService *order.Service
}

func NewOrderHandler(orderService *order.Service) *OrderHandler {
	return &OrderHandler{
		OrderService: orderService,
	}
}

func (h *OrderHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))

	err = h.OrderService.AddOrder(r.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidOrderNumber):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, order.ErrOrderAlreadyExists):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, order.ErrOrderBelongsToAnotherUser):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("starting GetOrdersByUser for user %d", userID)
	orders, err := h.OrderService.GetOrdersByUser(r.Context(), userID)
	log.Printf("finished GetOrdersByUser for user %d", userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
