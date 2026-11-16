package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"clothingStore/internal/models"
)

type ProductRepo struct{ db *sqlx.DB }

func NewProductRepo(db *sqlx.DB) *ProductRepo { return &ProductRepo{db: db} }

type ProductFilter struct {
	Type, Gender, Size, Color, Brand, Search, SortByPrice string
	Seasons                                               []string
	Limit, Offset                                         int
}

const productSelect = `SELECT p.*,
    (SELECT id FROM photos WHERE product_id = p.id ORDER BY rowid ASC LIMIT 1) AS main_photo_id
    FROM products p`

func (r *ProductRepo) Create(ctx context.Context, p *models.Product) (*models.Product, bool, error) {
	existing, err := r.findMatching(ctx, p)
	if err == nil && existing != nil {
		existing.Quantity += p.Quantity
		if _, err := r.db.ExecContext(ctx,
			`UPDATE products SET quantity = ? WHERE id = ?`,
			existing.Quantity, existing.ID); err != nil {
			return nil, false, err
		}
		return existing, true, nil
	}

	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO products (id,name,type,gender,size,color,description,brand,season,price,quantity)
         VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Name, p.Type, p.Gender, p.Size, p.Color,
		p.Description, p.Brand, p.Season, p.Price, p.Quantity,
	)
	if err != nil {
		return nil, false, err
	}
	return p, false, nil
}

func (r *ProductRepo) findMatching(ctx context.Context, p *models.Product) (*models.Product, error) {
	var found models.Product
	err := r.db.GetContext(ctx, &found,
		`SELECT * FROM products
         WHERE name = ? AND COALESCE(brand,'') = ?
           AND COALESCE(size,'') = ? AND COALESCE(color,'') = ?
         LIMIT 1`,
		p.Name, p.Brand, p.Size, p.Color,
	)
	if err != nil {
		return nil, err
	}
	return &found, nil
}

func (r *ProductRepo) List(ctx context.Context, f ProductFilter) ([]models.Product, error) {
	var where []string
	var args []any
	add := func(col, val string) {
		if val != "" {
			where = append(where, fmt.Sprintf("p.%s = ?", col))
			args = append(args, val)
		}
	}
	add("type", f.Type)
	add("gender", f.Gender)
	add("size", f.Size)
	add("color", f.Color)
	add("brand", f.Brand)

	if len(f.Seasons) > 0 {
		phs := make([]string, len(f.Seasons))
		for i, s := range f.Seasons {
			phs[i] = "?"
			args = append(args, s)
		}
		where = append(where, "p.season IN ("+strings.Join(phs, ",")+")")
	}
	if f.Search != "" {
		where = append(where, "(p.name LIKE ? OR p.id LIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like)
	}
	q := productSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	switch f.SortByPrice {
	case "asc":
		q += " ORDER BY p.price ASC"
	case "desc":
		q += " ORDER BY p.price DESC"
	}
	q += " LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	var items []models.Product
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*models.Product, error) {
	var p models.Product
	if err := r.db.GetContext(ctx, &p, productSelect+" WHERE p.id = ?", id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Update(ctx context.Context, p *models.Product) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products
         SET name=?, type=?, gender=?, size=?, color=?, description=?,
             brand=?, season=?, price=?, quantity=?
         WHERE id=?`,
		p.Name, p.Type, p.Gender, p.Size, p.Color, p.Description,
		p.Brand, p.Season, p.Price, p.Quantity, p.ID,
	)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = ?", id)
	return err
}
