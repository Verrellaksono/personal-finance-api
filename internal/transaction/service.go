package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInsufficientBalance = errors.New("Insufficient Balance")
	ErrAccountNotFounrd    = errors.New("Account Not Found")
	ErrInvalidAmount       = errors.New("Amount must be greater than 0")
	ErrInvalidType         = errors.New("Invalid Transaction Type")
)

type RepositoryContract interface {
	GetAccountBalanceForUpdate(ctx context.Context, tx *sql.Tx, accountID int64) (int64, error)
	UpdateAccountBalance(ctx context.Context, tx *sql.Tx, accountID int64, newBalance int64) error
	CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *Transaction) error
}

type Transaction struct {
	ID              int64     `json:"id"`
	AccountID       int64     `json:"account_id"`
	CategoryID      *int64    `json:"category_id,omitempty"`
	Amount          int64     `json:"amount"`
	Type            string    `json:"type"` //INCOME atau EXPENSE
	CurrentBalance  int64     `json:"current_balance"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateTransactionInput struct {
	AccountID       int64     `json:"account_id"`
	CategoryID      *int64    `json:"category_id"`
	Amount          int64     `json:"amount"`
	Type            string    `json:"type"` //INCOME atau EXPENSE
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
}

type Service struct {
	db   *sql.DB
	repo RepositoryContract
}

func NewService(db *sql.DB, repo RepositoryContract) *Service {
	return &Service{
		db:   db,
		repo: repo,
	}
}

func (s *Service) CreateTransaction(ctx context.Context, input CreateTransactionInput) (*Transaction, error) {
	// Validasi Input Dasar
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if input.Type != "INCOME" && input.Type != "EXPENSE" {
		return nil, ErrInvalidType
	}

	// Mulai Database Transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("Failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	currentBalance, err := s.repo.GetAccountBalanceForUpdate(ctx, tx, input.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFounrd
		}
		return nil, fmt.Errorf("Failed to lock account: %w", err)
	}

	var newBalance int64
	if input.Type == "EXPENSE" {
		if currentBalance < input.Amount {
			return nil, ErrInsufficientBalance
		}
		newBalance = currentBalance - input.Amount
	} else {
		newBalance = currentBalance + input.Amount
	}

	if err := s.repo.UpdateAccountBalance(ctx, tx, input.AccountID, newBalance); err != nil {
		return nil, fmt.Errorf("Failed to update balance: %w", err)
	}

	t := &Transaction{
		AccountID:       input.AccountID,
		CategoryID:      input.CategoryID,
		Amount:          input.Amount,
		Type:            input.Type,
		Description:     input.Description,
		TransactionDate: input.TransactionDate,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateTransaction(ctx, tx, t); err != nil {
		return nil, fmt.Errorf("Failed to create transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Failed to commit transaction: %w", err)
	}

	return t, nil
}
