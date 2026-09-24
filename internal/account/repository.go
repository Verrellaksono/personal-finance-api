package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, acc *Account) error {
	query := `
			INSERT INTO accounts (user_id, name, type, currency, balance)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at;
	`

	if err := r.db.QueryRowContext(
		ctx, query,
		acc.UserID,
		acc.Name,
		acc.Type,
		acc.Currency,
		acc.Balance,
	).Scan(&acc.ID, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
		return fmt.Errorf("Failed to insert account: %w", err)
	}

	return nil
}

func (r *Repository) GetById(ctx context.Context, id int64) (*Account, error) {
	query := `
			SELECT id, user_id, name, type, currency, balance, created_at, updated_at
			FROM accounts
			WHERE id = $1;
	`

	var acc Account
	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&acc.ID, &acc.UserID, &acc.Name, &acc.Type, &acc.Currency, &acc.Balance, &acc.CreatedAt, &acc.UpdatedAt)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Failed to query account: %w", err)
	}

	return &acc, nil
}
