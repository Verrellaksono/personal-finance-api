package budget

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidAmount = errors.New("budget amount must be greater than zero")
	ErrInvalidMonth  = errors.New("invalid month format, use YYYY-MM")
)

type RepositoryContract interface {
	GetBudgetProgress(ctx context.Context, userID int64, monthPeriod time.Time) ([]BudgetProgress, error)
	Upsert(ctx context.Context, b *Budget) error
}

type Service struct {
	repo RepositoryContract
}

func NewService(repo RepositoryContract) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetBudget(ctx context.Context, input SetBudgetInput) (*Budget, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	parsedMonth, err := time.Parse("2006-01", input.Month)
	if err != nil {
		return nil, ErrInvalidMonth
	}

	b := &Budget{
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		MonthPeriod: parsedMonth,
	}

	if err := s.repo.Upsert(ctx, b); err != nil {
		return nil, fmt.Errorf("failed to set budget: %w", err)
	}

	return b, nil
}

func (s *Service) CheckBudgetProgress(ctx context.Context, userID int64, monthStr string) ([]BudgetProgress, error) {
	parsedMonth, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return nil, ErrInvalidMonth
	}

	return s.repo.GetBudgetProgress(ctx, userID, parsedMonth)
}
