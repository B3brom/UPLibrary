package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type OrderItemRepo struct{ db *sqlx.DB }

func NewOrderItemRepo(db *sqlx.DB) *OrderItemRepo { return &OrderItemRepo{db: db} }

func (r *OrderItemRepo) Create(ctx context.Context, it *models.OrderItem) (string, error) {
	if it.Quantity <= 0 {
		it.Quantity = 1
	}
	if it.ID == "" {
		it.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO order_items (id, order_id, product_id, price, quantity)
         VALUES (?,?,?,?,?)`,
		it.ID, it.OrderID, it.ProductID, it.Price, it.Quantity,
	)
	if err != nil {
		return "", err
	}
	return it.ID, nil
}

func (r *OrderItemRepo) ListByOrder(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := r.db.SelectContext(ctx, &items,
		"SELECT * FROM order_items WHERE order_id = ? ORDER BY rowid", orderID)
	return items, err
}
