package category

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrEmptyName   = errors.New("category name is required")
	ErrInvalidType = errors.New("category type must be INCOME or EXPENSE")
)

type RepositoryContract interface {
	Create(ctx context.Context, c *Category) error
	ListByUserID(ctx context.Context, userID int64) ([]Category, error)
}

type Service struct {
	repo RepositoryContract
}

func NewService(repo RepositoryContract) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCategory(ctx context.Context, input CreateCategoryInput) (*Category, error) {
	trimmedName := strings.TrimSpace(input.Name)
	if trimmedName == "" {
		return nil, ErrEmptyName
	}

	categoryType := strings.ToUpper(strings.TrimSpace(input.Type))
	if categoryType != "INCOME" && categoryType != "EXPENSE" {
		return nil, ErrInvalidType
	}

	c := &Category{
		UserID: input.UserID,
		Name:   trimmedName,
		Type:   categoryType,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Service) ListCategories(ctx context.Context, userID int64) ([]Category, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user is")
	}

	return s.repo.ListByUserID(ctx, userID)
}
