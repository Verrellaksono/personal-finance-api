package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrDuplicateCategory = errors.New("category with this name and type already exists")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *Category) error {
	query := `
			INSERT INTO categories (user_id, name, type)
			VALUES ($1, $2, $3)
			RETURNING id, created_at, updated_at;
	`

	if err := r.db.QueryRowContext(ctx, query, c.UserID, c.Name, c.Type).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "uq_categories_user_name_type") {
			return ErrDuplicateCategory
		}
		return fmt.Errorf("failed to insert category: %w", err)
	}

	return nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID int64) ([]Category, error) {
	query := `
			SELECT id, user_id, name, type, created_at, updated_at
			FROM categories
			WHERE user_id = $1;
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}
