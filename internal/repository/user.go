package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type UserRepo struct{ db *sqlx.DB }

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{db: db} }

const userColumns = `id, name, email, COALESCE(phone, '') AS phone,
	password_hash, role_id, registration_date, balance`

func (r *UserRepo) Create(ctx context.Context, u *models.User) (string, error) {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	var phone any
	if u.Phone != "" {
		phone = u.Phone
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, phone, password_hash, role_id)
         VALUES (?,?,?,?,?,?)`,
		u.ID, u.Name, u.Email, phone, u.PasswordHash, u.RoleID,
	)
	if err != nil {
		return "", err
	}
	return u.ID, nil
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	var items []models.User
	err := r.db.SelectContext(ctx, &items,
		"SELECT "+userColumns+" FROM users ORDER BY name LIMIT ? OFFSET ?", limit, offset)
	return items, err
}

func (r *UserRepo) ListByRole(ctx context.Context, roleID string, limit, offset int) ([]models.User, error) {
	var items []models.User
	err := r.db.SelectContext(ctx, &items,
		"SELECT "+userColumns+" FROM users WHERE role_id = ? ORDER BY name LIMIT ? OFFSET ?",
		roleID, limit, offset)
	return items, err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	if err := r.db.GetContext(ctx, &u, "SELECT "+userColumns+" FROM users WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := r.db.GetContext(ctx, &u, "SELECT "+userColumns+" FROM users WHERE email = ?", email); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name=?, email=?, phone=?, role_id=? WHERE id=?`,
		u.Name, u.Email, u.Phone, u.RoleID, u.ID,
	)
	return err
}

func (r *UserRepo) UpdateRole(ctx context.Context, id, roleID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET role_id = ? WHERE id = ?`, roleID, id)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

// AddBalance — пополнение баланса. Возвращает новый баланс.
func (r *UserRepo) AddBalance(ctx context.Context, userID string, amount float64) (float64, error) {
	if amount <= 0 {
		return 0, errors.New("сумма должна быть положительной")
	}
	res, err := r.db.ExecContext(ctx,
		"UPDATE users SET balance = balance + ? WHERE id = ?", amount, userID)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return 0, errors.New("пользователь не найден")
	}
	u, err := r.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return u.Balance, nil
}
