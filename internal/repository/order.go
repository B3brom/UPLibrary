package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type OrderRepo struct{ db *sqlx.DB }

func NewOrderRepo(db *sqlx.DB) *OrderRepo { return &OrderRepo{db: db} }

type OrderFilter struct {
	From, To time.Time
	Status   string
	Limit    int
	Offset   int
}

const orderListSelect = `SELECT o.id, o.user_id, o.datetime, o.status, o.refunded, o.cancel_reason,
    u.name AS user_name,
    COALESCE((SELECT SUM(quantity) FROM order_items WHERE order_id = o.id), 0) AS items_count,
    COALESCE((SELECT SUM(price * quantity) FROM order_items WHERE order_id = o.id), 0) AS total
    FROM orders o
    JOIN users u ON u.id = o.user_id`

func (r *OrderRepo) List(ctx context.Context, f OrderFilter) ([]models.OrderListItem, error) {
	var where []string
	var args []any
	if !f.From.IsZero() {
		where = append(where, "o.datetime >= ?")
		args = append(args, f.From)
	}
	if !f.To.IsZero() {
		where = append(where, "o.datetime <= ?")
		args = append(args, f.To)
	}
	if f.Status != "" {
		where = append(where, "o.status = ?")
		args = append(args, f.Status)
	}
	q := orderListSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY o.datetime DESC LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	var items []models.OrderListItem
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]models.OrderListItem, error) {
	var items []models.OrderListItem
	err := r.db.SelectContext(ctx, &items,
		orderListSelect+" WHERE o.user_id = ? ORDER BY o.datetime DESC LIMIT ? OFFSET ?",
		userID, limit, offset)
	return items, err
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var o models.Order
	if err := r.db.GetContext(ctx, &o, "SELECT * FROM orders WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetDetail(ctx context.Context, id string) (*models.OrderDetail, error) {
	var d models.OrderDetail
	if err := r.db.GetContext(ctx, &d,
		`SELECT o.id, o.user_id, o.datetime, o.status, o.refunded, o.cancel_reason,
                u.name AS user_name
         FROM orders o JOIN users u ON u.id = o.user_id
         WHERE o.id = ?`, id); err != nil {
		return nil, err
	}
	var items []models.OrderItemDetail
	if err := r.db.SelectContext(ctx, &items,
		`SELECT oi.id, oi.product_id, oi.price, oi.quantity,
                p.name AS product_name,
                (SELECT id FROM photos WHERE product_id = p.id ORDER BY rowid ASC LIMIT 1) AS photo_id
         FROM order_items oi
         JOIN products p ON p.id = oi.product_id
         WHERE oi.order_id = ?
         ORDER BY oi.rowid`, id); err != nil {
		return nil, err
	}
	d.Items = items
	for _, it := range items {
		d.Total += it.Price * float64(it.Quantity)
	}
	return &d, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	var o models.Order
	if err := r.db.GetContext(ctx, &o, "SELECT * FROM orders WHERE id = ?", id); err != nil {
		return errors.New("Заказ не найден")
	}
	if o.Refunded || o.Status == models.StatusCancelled {
		return errors.New("Отменённый заказ нельзя изменить")
	}
	if o.Status == models.StatusIssued {
		return errors.New("Полученный заказ нельзя изменить")
	}
	if status == models.StatusCancelled {
		return errors.New("Для отмены используйте кнопку «Отменить заказ»")
	}
	if status == models.StatusIssued && o.Status != models.StatusReady {
		return errors.New("Статус «Получен» можно установить только из «Готов к выдаче»")
	}
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *OrderRepo) Cancel(ctx context.Context, orderID, callerUserID, reason string, isClient bool) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var o models.Order
	if err := tx.GetContext(ctx, &o, "SELECT * FROM orders WHERE id = ?", orderID); err != nil {
		return errors.New("Заказ не найден")
	}
	if isClient && o.UserID != callerUserID {
		return errors.New("Это не ваш заказ")
	}
	if o.Refunded || o.Status == models.StatusCancelled {
		return errors.New("Заказ уже отменён")
	}
	if o.Status == models.StatusIssued {
		return errors.New("Полученный заказ нельзя отменить")
	}
	if isClient && o.Status != models.StatusAccepted {
		return errors.New("Отменить заказ можно только на этапе «Принят». Обратитесь к менеджеру.")
	}
	if reason == "" {
		return errors.New("Укажите причину отмены")
	}

	rows, err := tx.QueryContext(ctx,
		"SELECT product_id, price, quantity FROM order_items WHERE order_id = ?", orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type line struct {
		ProductID string
		Price     float64
		Quantity  int
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.ProductID, &l.Price, &l.Quantity); err != nil {
			return err
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	total := 0.0
	for _, l := range lines {
		total += l.Price * float64(l.Quantity)
		if _, err := tx.ExecContext(ctx,
			"UPDATE products SET quantity = quantity + ? WHERE id = ?",
			l.Quantity, l.ProductID); err != nil {
			return err
		}
	}
	// Возврат строго владельцу заказа (o.UserID), а не callerUserID.
	if total > 0 {
		if _, err := tx.ExecContext(ctx,
			"UPDATE users SET balance = balance + ? WHERE id = ?",
			total, o.UserID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE orders SET status = ?, refunded = 1, cancel_reason = ? WHERE id = ?",
		models.StatusCancelled, reason, orderID); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *OrderRepo) CreateWithItems(ctx context.Context, o *models.Order, items []models.OrderItem) (string, error) {
	if len(items) == 0 {
		return "", errors.New("пустой заказ")
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if o.ID == "" {
		o.ID = uuid.NewString()
	}

	type line struct {
		ProductID string
		Price     float64
		Quantity  int
		Name      string
	}
	lines := make([]line, 0, len(items))
	total := 0.0
	for _, it := range items {
		var p models.Product
		if err := tx.GetContext(ctx, &p, "SELECT * FROM products WHERE id = ?", it.ProductID); err != nil {
			return "", fmt.Errorf("товар не найден: %s", it.ProductID)
		}
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		if p.Quantity < qty {
			return "", fmt.Errorf("«%s»: в наличии %d, запрошено %d", p.Name, p.Quantity, qty)
		}
		total += p.Price * float64(qty)
		lines = append(lines, line{ProductID: p.ID, Price: p.Price, Quantity: qty, Name: p.Name})
	}

	res, err := tx.ExecContext(ctx,
		"UPDATE users SET balance = balance - ? WHERE id = ? AND balance >= ?",
		total, o.UserID, total)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", fmt.Errorf("недостаточно средств: нужно %.2f ₽", total)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, status) VALUES (?,?,?)`,
		o.ID, o.UserID, o.Status); err != nil {
		return "", err
	}

	for _, ln := range lines {
		if _, err := tx.ExecContext(ctx,
			"UPDATE products SET quantity = quantity - ? WHERE id = ?",
			ln.Quantity, ln.ProductID); err != nil {
			return "", err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (id, order_id, product_id, price, quantity)
             VALUES (?,?,?,?,?)`,
			uuid.NewString(), o.ID, ln.ProductID, ln.Price, ln.Quantity); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return o.ID, nil
}
