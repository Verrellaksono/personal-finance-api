package transfer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrAccountNotFound     = errors.New("one or both accounts not found")
	ErrInsufficientBalance = errors.New("insufficient balance in source account")
	ErrIdempotencyPending  = errors.New("request with this idempotency key is currently processing")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) TryAcquireIdempotencyKey(ctx context.Context, key string, userID int64) (*IdempotencyRecord, bool, error) {
	query := `
			INSERT INTO idempotency_keys (key, user_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT(key) DO NOTHING
			RETURNING key, user_id, status, response_code, response_body, created_at, updated_at; 
	`

	var record IdempotencyRecord
	err := r.db.QueryRowContext(ctx, query, key, userID, IdempotencyStatusPending).Scan(
		&record.Key,
		&record.UserID,
		&record.Status,
		&record.ResponseCode,
		&record.ResponseBody,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == nil {
		return &record, true, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		selectQuery := `
				SELECT key, user_id, status, response_code, response_body, created_at, updated_at
				FROM idempotency_keys
				WHERE KEY = $1;
		`

		err := r.db.QueryRowContext(ctx, selectQuery, key).Scan(
			&record.Key,
			&record.UserID,
			&record.Status,
			&record.ResponseCode,
			&record.ResponseBody,
			&record.CreatedAt,
			&record.UpdatedAt,
		)

		if err != nil {
			return nil, false, fmt.Errorf("failed to fetch existing idempotency record: %w", err)
		}

		return &record, false, nil
	}

	return nil, false, fmt.Errorf("failed to acquire idempotency key: %w", err)
}

func (r *Repository) SaveIdempotencySuccess(ctx context.Context, key string, statusCode int, responseBody []byte) error {
	query := `
			UPDATE idempotency_keys
			SET status = $1, response_code = $2, response_body = $3, updated_at = NOW()
			WHERE key = $4;
	`

	_, err := r.db.ExecContext(ctx, query, IdempotencyStatusCompleted, statusCode, responseBody, key)
	return err
}

func (r *Repository) MarkIdempotencyFailed(ctx context.Context, key string) error {
	query := `
			UPDATE idempotency_keys
			SET status = $1, updated_at = NOW()
			WHERE key = $2;
	`

	_, err := r.db.ExecContext(ctx, query, IdempotencyStatusFailed, key)
	return err
}

func (r *Repository) ExecuteTransfer(ctx context.Context, t *Transfer) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Pencegahan Deadlock
	firstID, secondID := t.FromAccountID, t.ToAccountID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	// Lock akun 1
	balance1, ownerID1, err := r.getAccountForUpdate(ctx, tx, firstID)
	if err != nil {
		return 0, err
	}

	// Lock akun 2
	balance2, ownerID2, err := r.getAccountForUpdate(ctx, tx, secondID)
	if err != nil {
		return 0, err
	}

	var fromBalance, fromOwnerID int64
	var toBalance, toOwnerID int64
	if t.FromAccountID == firstID {
		fromBalance, fromOwnerID = balance1, ownerID1
		toBalance, toOwnerID = balance2, ownerID2
	} else {
		fromBalance, fromOwnerID = balance2, ownerID2
		toBalance, toOwnerID = balance1, ownerID1
	}

	// Validasi Hak Kepemilikan Akun(Anti-IDOR)
	if fromOwnerID != t.UserID || toOwnerID != t.UserID {
		return 0, errors.New("unauthorized: both accounts must belong to the requesting user")
	}

	// Check Saldo Akun Asal
	if fromBalance < t.Amount {
		return 0, ErrInsufficientBalance
	}

	// Perhitungan Transfer Saldo
	newFromBalance := fromBalance - t.Amount
	newToBalance := toBalance + t.Amount

	// Update Saldo Akun Asal
	updateQuery := `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE id = $2;`
	if _, err := tx.ExecContext(ctx, updateQuery, newFromBalance, t.FromAccountID); err != nil {
		return 0, fmt.Errorf("failed to debit from_account: %w", err)
	}

	// Update Saldo Akun Tujuan
	if _, err := tx.ExecContext(ctx, updateQuery, newToBalance, t.ToAccountID); err != nil {
		return 0, fmt.Errorf("failed to credit to_account: %w", err)
	}

	// Simpan Riwayat Mutasi ke Tabel transfers
	insertTransferQuery := `
		INSERT INTO transfers (user_id, from_account_id, to_account_id, amount, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at;
	`
	if err := tx.QueryRowContext(ctx, insertTransferQuery, t.UserID, t.FromAccountID, t.ToAccountID, t.Amount, t.Notes).
		Scan(&t.ID, &t.CreatedAt); err != nil {
		return 0, fmt.Errorf("failed to insert transfer record: %w", err)
	}

	// Commit Transaksi
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transfer tx: %w", err)
	}

	return newFromBalance, nil
}

func (r *Repository) getAccountForUpdate(ctx context.Context, tx *sql.Tx, accountID int64) (balance int64, ownerID int64, err error) {
	query := `
			SELECT balance, user_id 
			FROM accounts
			WHERE id = $1
			FOR UPDATE;
	`

	err = tx.QueryRowContext(ctx, query, accountID).Scan(&balance, &ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, ErrAccountNotFound
		}
		return 0, 0, err
	}

	return balance, ownerID, nil
}
