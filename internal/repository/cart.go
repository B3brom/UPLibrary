package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type CartRepo struct{ db *sqlx.DB }

func NewCartRepo(db *sqlx.DB) *CartRepo { return &CartRepo{db: db} }

func (r *CartRepo) Create(ctx context.Context, c *models.Cart) (string, error) {
	var existing models.Cart
	err := r.db.GetContext(ctx, &existing,
		"SELECT * FROM carts WHERE user_id = ? AND product_id = ?",
		c.UserID, c.ProductID)
	if err == nil {
		if _, err := r.db.ExecContext(ctx,
			"UPDATE carts SET count = count + ? WHERE id = ?", c.Count, existing.ID); err != nil {
			return "", err
		}
		return existing.ID, nil
	}

	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO carts (id, user_id, product_id, count) VALUES (?,?,?,?)`,
		c.ID, c.UserID, c.ProductID, c.Count,
	)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

func (r *CartRepo) GetByUser(ctx context.Context, userID string) ([]models.Cart, error) {
	var items []models.Cart
	err := r.db.SelectContext(ctx, &items,
		"SELECT * FROM carts WHERE user_id = ? ORDER BY rowid", userID)
	return items, err
}

func (r *CartRepo) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	var c models.Cart
	if err := r.db.GetContext(ctx, &c, "SELECT * FROM carts WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CartRepo) UpdateForUser(ctx context.Context, c *models.Cart, userID string) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		"UPDATE carts SET count = ? WHERE id = ? AND user_id = ?", c.Count, c.ID, userID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

func (r *CartRepo) DeleteForUser(ctx context.Context, id, userID string) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM carts WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

func (r *CartRepo) ClearByUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM carts WHERE user_id = ?", userID)
	return err
}
