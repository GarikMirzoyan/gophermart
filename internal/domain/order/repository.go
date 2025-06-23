package order

import "context"

type Repository interface {
	// Добавить заказ, вернуть ошибку в случае конфликта или некорректного номера
	AddOrder(ctx context.Context, order *Order) error

	// Получить список заказов пользователя, отсортированных по uploaded_at DESC
	GetOrdersByUser(ctx context.Context, userID int) ([]*Order, error)

	// Проверить существует ли номер заказа и кому он принадлежит
	GetOrderOwner(ctx context.Context, number string) (int, error)

	// Обновить статус и начисленные баллы по заказу
	UpdateAccrual(ctx context.Context, orderNumber string, status string, accrual int64) error

	// Обновить только статус заказа
	UpdateStatus(ctx context.Context, orderNumber string, status string) error
}
