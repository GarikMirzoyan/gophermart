package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GarikMirzoyan/gophermart/internal/domain/order"
)

var ErrNoRows = errors.New("no rows in result set")

type OrderPG struct {
	DB *sql.DB
}

func NewOrderPG(db *sql.DB) *OrderPG {
	return &OrderPG{DB: db}
}

// Добавить заказ
func (r *OrderPG) AddOrder(ctx context.Context, o *order.Order) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO orders (number, status, accrual, uploaded_at, user_id)
		VALUES ($1, $2, $3, $4, $5)
	`, o.Number, string(o.Status), o.Accrual, o.UploadedAt, o.UserID)
	return err
}

// Получить список заказов пользователя, сортировка по времени DESC
func (r *OrderPG) GetOrdersByUser(ctx context.Context, userID int) ([]*order.Order, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*order.Order
	for rows.Next() {
		var o order.Order
		var accrual sql.NullInt32
		var status string

		err := rows.Scan(&o.Number, &status, &accrual, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		o.Status = order.Status(status)
		if accrual.Valid {
			v := int(accrual.Int32)
			o.Accrual = &v
		}
		o.UserID = userID
		orders = append(orders, &o)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// Получить владельца заказа по номеру
func (r *OrderPG) GetOrderOwner(ctx context.Context, number string) (int, error) {
	var userID int
	err := r.DB.QueryRowContext(ctx, `
		SELECT user_id FROM orders WHERE number = $1
	`, number).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil // Заказа нет
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *OrderPG) UpdateAccrual(ctx context.Context, orderNumber string, status string, accrual float64) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3
	`, status, accrual, orderNumber)
	return err
}

func (r *OrderPG) UpdateStatus(ctx context.Context, orderNumber string, status string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE orders
		SET status = $1
		WHERE number = $2
	`, status, orderNumber)
	return err
}

func (r *OrderPG) GetOrdersForProcessing(ctx context.Context) ([]*order.Order, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT number, user_id
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*order.Order
	for rows.Next() {
		var o order.Order
		if err := rows.Scan(&o.Number, &o.UserID); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
