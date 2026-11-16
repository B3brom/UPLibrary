package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type PhotoRepo struct{ db *sqlx.DB }

func NewPhotoRepo(db *sqlx.DB) *PhotoRepo { return &PhotoRepo{db: db} }

func (r *PhotoRepo) Create(ctx context.Context, p *models.Photo) (string, error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO photos (id, filename, product_id) VALUES (?,?,?)`,
		p.ID, p.Filename, p.ProductID,
	)
	if err != nil {
		return "", err
	}
	return p.ID, nil
}

func (r *PhotoRepo) GetByID(ctx context.Context, id string) (*models.Photo, error) {
	var p models.Photo
	if err := r.db.GetContext(ctx, &p, "SELECT * FROM photos WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PhotoRepo) ListByProduct(ctx context.Context, productID string) ([]models.Photo, error) {
	var items []models.Photo
	err := r.db.SelectContext(ctx, &items,
		"SELECT * FROM photos WHERE product_id = ? ORDER BY rowid", productID)
	return items, err
}

func (r *PhotoRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM photos WHERE id = ?", id)
	return err
}
