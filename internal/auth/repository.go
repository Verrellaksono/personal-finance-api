package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrDuplicateEmail = errors.New("email already registered")
	ErrUserNotFound   = errors.New("user not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
			INSERT INTO users (email, password_hash, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			RETURNING id, created_at, updated_at;
	`

	err := r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("Failed to insert user: %w", err)
	}

	return nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
			SELECT id, email, password_hash, created_at, updated_at
			FROM users
			WHERE email = $1;
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("Failed to query user: %w", err)
	}

	return &user, nil
}
