package http

import (
	"net/http"

	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/handler"
	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/middleware"
	infraauth "github.com/GarikMirzoyan/gophermart/internal/infrastructure/auth"
	LoyaltyHandler "github.com/GarikMirzoyan/gophermart/internal/loyalty/handler"
	"github.com/go-chi/chi/v5"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
	withdrawalHandler *handler.WithdrawalHandler,
	loyaltyHandler *LoyaltyHandler.LoyaltyHandler,
	jwtManager *infraauth.JWTManager,
) http.Handler {
	r := chi.NewRouter()

	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.Get("/api/orders/{number}", loyaltyHandler.GetOrderAccrual)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtManager))

		r.Post("/api/user/orders", orderHandler.AddOrder)
		r.Get("/api/user/orders", orderHandler.GetOrders)

		r.Get("/api/user/balance", balanceHandler.GetBalance)

		r.Post("/api/user/balance/withdraw", withdrawalHandler.Withdraw)
		r.Get("/api/user/withdrawals", withdrawalHandler.GetWithdrawals)
	})

	return r
}
