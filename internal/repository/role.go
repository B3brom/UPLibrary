package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type RoleRepo struct{ db *sqlx.DB }

func NewRoleRepo(db *sqlx.DB) *RoleRepo { return &RoleRepo{db: db} }

func (r *RoleRepo) Create(ctx context.Context, role *models.Role) (string, error) {
	if role.ID == "" {
		role.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO roles (id, name) VALUES (?,?)",
		role.ID, role.Name)
	if err != nil {
		return "", err
	}
	return role.ID, nil
}

func (r *RoleRepo) List(ctx context.Context) ([]models.Role, error) {
	var items []models.Role
	err := r.db.SelectContext(ctx, &items, "SELECT * FROM roles ORDER BY name")
	return items, err
}

func (r *RoleRepo) GetByID(ctx context.Context, id string) (*models.Role, error) {
	var role models.Role
	if err := r.db.GetContext(ctx, &role, "SELECT * FROM roles WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) Update(ctx context.Context, role *models.Role) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE roles SET name = ? WHERE id = ?", role.Name, role.ID)
	return err
}

func (r *RoleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM roles WHERE id = ?", id)
	return err
}
