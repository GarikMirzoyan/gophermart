package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GarikMirzoyan/gophermart/internal/config"
	delivery "github.com/GarikMirzoyan/gophermart/internal/delivery/http"
	"github.com/GarikMirzoyan/gophermart/internal/delivery/http/handler"
	"github.com/GarikMirzoyan/gophermart/internal/infrastructure/auth"
	"github.com/GarikMirzoyan/gophermart/internal/infrastructure/storage"
	"github.com/GarikMirzoyan/gophermart/internal/loyalty"
	LoyaltyHandler "github.com/GarikMirzoyan/gophermart/internal/loyalty/handler"
	authusecase "github.com/GarikMirzoyan/gophermart/internal/usecase/auth"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/balance"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/order"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/withdrawal"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type App struct {
	Config            *config.Config
	JWTManager        *auth.JWTManager
	OrderService      *order.Service
	AuthService       *authusecase.Service
	BalanceService    *balance.Service
	WithdrawalService *withdrawal.Service
	LoyaltyService    *loyalty.Service
	DB                *sql.DB
}

func New() (*App, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, proceeding without it")
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	// TODO: сделать секрет ключ из переменной окружения
	jwtManager := auth.NewJWTManager("supersecretkey", time.Hour*24)

	// Для работы с пользователями
	userRepo := storage.NewUserPG(db)
	authService := authusecase.New(userRepo)

	// Для работы с балансом
	balanceRepo := storage.NewBalancePG(db)
	balanceService := balance.New(balanceRepo)

	// Для работы с выводами
	withdrawalRepo := storage.NewWithdrawalPG(db)
	withdrawalService := withdrawal.New(withdrawalRepo)

	// Для работы с баллами
	loyaltyService := loyalty.New()

	// Для работы с заказами
	orderRepo := storage.NewOrderPG(db)
	orderService := order.New(orderRepo, loyaltyService, balanceService)

	return &App{
		Config:            cfg,
		JWTManager:        jwtManager,
		OrderService:      orderService,
		AuthService:       authService,
		BalanceService:    balanceService,
		WithdrawalService: withdrawalService,
		LoyaltyService:    loyaltyService,
		DB:                db,
	}, nil
}

func (a *App) Run() error {
	authHandler := handler.NewAuthHandler(a.AuthService, a.JWTManager)
	orderHandler := handler.NewOrderHandler(a.OrderService)
	balanceHandler := handler.NewBalanceHandler(a.BalanceService)
	withdrawalHandler := handler.NewWithdrawalHandler(a.WithdrawalService)
	loyaltyHandler := LoyaltyHandler.NewLoyaltyHandler(a.LoyaltyService)
	router := delivery.NewRouter(authHandler, orderHandler, balanceHandler, withdrawalHandler, loyaltyHandler, a.JWTManager)

	defer a.DB.Close()
	log.Printf("Starting server on %s...", a.Config.RunAddress)
	return http.ListenAndServe(a.Config.RunAddress, router)
}
