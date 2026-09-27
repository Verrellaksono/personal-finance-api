package transaction

import (
	"context"
	"database/sql"
)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) GetAccountBalanceForUpdate(ctx context.Context, tx *sql.Tx, accountID int64) (balance int64, ownerID int64, err error) {
	query := `
			SELECT balance, user_id
			FROM accounts
			WHERE id = $1
			FOR UPDATE;
	`

	if err := tx.QueryRowContext(ctx, query, accountID).Scan(&balance, &ownerID); err != nil {
		return 0, 0, err
	}

	return balance, ownerID, nil
}

func (r *Repository) UpdateAccountBalance(ctx context.Context, tx *sql.Tx, accountID int64, newBalance int64) error {
	query := `
			UPDATE accounts
			SET balance = $1
			WHERE id = $2;
	`
	_, err := tx.ExecContext(ctx, query, newBalance, accountID)
	return err
}

func (r *Repository) CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *Transaction) error {
	query := `
			INSERT INTO transactions (account_id, category_id, amount, type, description, transaction_date, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			RETURNING id, created_at;
	`
	return tx.QueryRowContext(
		ctx,
		query,
		transaction.AccountID,
		transaction.CategoryID,
		transaction.Amount,
		transaction.Type,
		transaction.Description,
		transaction.TransactionDate,
	).Scan(
		&transaction.ID,
		&transaction.CreatedAt,
	)
}
