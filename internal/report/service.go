package report

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidMonth = errors.New("invalid month format, use YYYY-MM")

type RepositoryContract interface {
	GetMonthlyTotals(ctx context.Context, userID int64, monthPeriod time.Time) (totalIncome int64, totalExpense int64, err error)
	GetExpenseByCategoryBreakdown(ctx context.Context, userID int64, monthPeriod time.Time) ([]CategoryExpenseBreakdown, error)
}

type Service struct {
	repo RepositoryContract
}

func NewService(repo RepositoryContract) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GenerateMonthlySummary(ctx context.Context, userID int64, monthStr string) (*MonthlySummaryResponse, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	parsedMonth, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return nil, ErrInvalidMonth
	}

	// Ambil agregat income dan expense
	income, expense, err := s.repo.GetMonthlyTotals(ctx, userID, parsedMonth)
	if err != nil {
		return nil, err
	}

	breakdown, err := s.repo.GetExpenseByCategoryBreakdown(ctx, userID, parsedMonth)
	if err != nil {
		return nil, err
	}

	return &MonthlySummaryResponse{
		MonthPeriod:   parsedMonth,
		TotalIncome:   income,
		TotalExpense:  expense,
		NetSavings:    income - expense,
		ExpenseByType: breakdown,
	}, nil
}
