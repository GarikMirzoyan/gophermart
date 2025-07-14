package orderhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/middleware"
	"github.com/GarikMirzoyan/gophermart/internal/domain/order"
	orderUseCase "github.com/GarikMirzoyan/gophermart/internal/usecase/order"
)

type Service interface {
	AddOrder(ctx context.Context, userID int, number string) error
	GetOrdersByUser(ctx context.Context, userID int) ([]*order.Order, error)
	ProcessPendingOrders(ctx context.Context)
}

type Handler struct {
	OrderService Service
}

func New(orderService Service) *Handler {
	return &Handler{
		OrderService: orderService,
	}
}

func (h *Handler) AddOrder(w http.ResponseWriter, r *http.Request) {
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
		case errors.Is(err, orderUseCase.ErrInvalidOrderNumber):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, orderUseCase.ErrOrderAlreadyExists):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, orderUseCase.ErrOrderBelongsToAnotherUser):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.OrderService.GetOrdersByUser(r.Context(), userID)

	if err != nil {
		http.Error(w, fmt.Sprintf("server error: %v", err), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
