package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidAmount       = errors.New("transfer amount must be greater than zero")
	ErrSameAccount         = errors.New("cannot transfer funds to the same account")
	ErrMissingKey          = errors.New("idempotency-key header is required")
	ErrIdempotencyConflict = errors.New("a request with this idempotency key is already in progress, please retry shortly")
)

type RepositoryContract interface {
	TryAcquireIdempotencyKey(ctx context.Context, key string, userID int64) (*IdempotencyRecord, bool, error)
	SaveIdempotencySuccess(ctx context.Context, key string, statusCode int, responseBody []byte) error
	MarkIdempotencyFailed(ctx context.Context, key string) error
	ExecuteTransfer(ctx context.Context, t *Transfer) (int64, error)
}

type Service struct {
	repo RepositoryContract
}

func NewService(repo RepositoryContract) *Service {
	return &Service{repo: repo}
}

func (s *Service) Processtransfer(ctx context.Context, input CreateTransferInput) (*TransferResult, error) {
	if input.IdempotencyKey == "" {
		return nil, ErrMissingKey
	}
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if input.FromAccountID == input.ToAccountID {
		return nil, ErrSameAccount
	}

	// Cek & Akuisisi Idempotency Key
	record, isNew, err := s.repo.TryAcquireIdempotencyKey(ctx, input.IdempotencyKey, input.UserID)
	if err != nil {
		return nil, err
	}

	if record.Status == IdempotencyStatusCompleted {
		var cachedResult TransferResult
		if err := json.Unmarshal(record.ResponseBody, &cachedResult); err != nil {
			return nil, fmt.Errorf("failed to parse cached response: %w", err)
		}
		return &cachedResult, nil
	}

	if !isNew && record.Status == IdempotencyStatusPending {
		return nil, ErrIdempotencyConflict
	}

	t := &Transfer{
		UserID:        input.UserID,
		FromAccountID: input.FromAccountID,
		ToAccountID:   input.ToAccountID,
		Amount:        input.Amount,
		Notes:         input.Notes,
	}

	newBalance, err := s.repo.ExecuteTransfer(ctx, t)
	if err != nil {
		_ = s.repo.MarkIdempotencyFailed(ctx, input.IdempotencyKey)
		return nil, err
	}

	result := &TransferResult{
		Transfer:           *t,
		FromAccountBalance: newBalance,
	}

	respBytes, err := json.Marshal(result)
	if err == nil {
		_ = s.repo.SaveIdempotencySuccess(ctx, input.IdempotencyKey, 201, respBytes)
	}

	return result, err
}
