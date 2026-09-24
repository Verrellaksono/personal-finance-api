package account

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrEmptyName      = errors.New("Account name is required.")
	ErrInvalidType    = errors.New("Invalid account type, must be Bank, E-wallet or Cash.")
	ErrInvalidBalance = errors.New("Initial balance cannot be negative.")
	ErrNotFound       = errors.New("Account not found")
)

type RepositoryContract interface {
	Create(ctx context.Context, acc *Account) error
	GetById(ctx context.Context, id int64) (*Account, error)
}

type CreateAccountInput struct {
	UserID         int64  `json:"user_id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Currency       string `json:"currency"`
	InitialBalance int64  `json:"initial_balance"`
}

type Service struct {
	repo RepositoryContract
}

func NewService(repo RepositoryContract) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateAccount(ctx context.Context, input *CreateAccountInput) (*Account, error) {
	trimmedName := strings.TrimSpace(input.Name)
	if trimmedName == "" {
		return nil, ErrEmptyName
	}

	accountType := strings.ToUpper(strings.TrimSpace(input.Type))
	if accountType != "BANK" && accountType != "E-WALLET" && accountType != "CASH" {
		return nil, ErrInvalidType
	}

	if input.InitialBalance <= 0 {
		return nil, ErrInvalidBalance
	}

	currency := strings.TrimSpace(input.Currency)
	if currency == "" {
		currency = "IDR"
	}

	userID := input.UserID
	if userID == 0 {
		userID = 1
	}

	acc := &Account{
		UserID:   userID,
		Name:     trimmedName,
		Type:     accountType,
		Currency: currency,
		Balance:  input.InitialBalance,
	}

	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *Service) GetAccountByID(ctx context.Context, id int64) (*Account, error) {
	if id <= 0 {
		return nil, errors.New("Invalid account id")
	}

	return s.repo.GetById(ctx, id)
}
