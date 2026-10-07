package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type mockTransferRepo struct {
	tryAcquireFunc  func(ctx context.Context, key string, userID int64) (*IdempotencyRecord, bool, error)
	saveSuccessFunc func(ctx context.Context, key string, statusCode int, responseBody []byte) error
	markFailedFunc  func(ctx context.Context, key string) error
	executeTxFunc   func(ctx context.Context, t *Transfer) (int64, error)
}

func (m *mockTransferRepo) TryAcquireIdempotencyKey(ctx context.Context, key string, userID int64) (*IdempotencyRecord, bool, error) {
	if m.tryAcquireFunc != nil {
		return m.tryAcquireFunc(ctx, key, userID)
	}
	return &IdempotencyRecord{Key: key, UserID: userID, Status: IdempotencyStatusPending}, true, nil
}

func (m *mockTransferRepo) SaveIdempotencySuccess(ctx context.Context, key string, statusCode int, responseBody []byte) error {
	if m.saveSuccessFunc != nil {
		return m.saveSuccessFunc(ctx, key, statusCode, responseBody)
	}
	return nil
}

func (m *mockTransferRepo) MarkIdempotencyFailed(ctx context.Context, key string) error {
	if m.markFailedFunc != nil {
		return m.markFailedFunc(ctx, key)
	}
	return nil
}

func (m *mockTransferRepo) ExecuteTransfer(ctx context.Context, t *Transfer) (int64, error) {
	if m.executeTxFunc != nil {
		return m.executeTxFunc(ctx, t)
	}
	t.ID = 100
	t.CreatedAt = time.Now()
	return 50000, nil // saldo sisa default
}

func TestProcessTransfer(t *testing.T) {
	sampleCachedResult := TransferResult{
		Transfer: Transfer{
			ID:            90,
			UserID:        1,
			FromAccountID: 1,
			ToAccountID:   2,
			Amount:        25000,
		},
		FromAccountBalance: 75000,
	}
	cachedJSON, _ := json.Marshal(sampleCachedResult)

	tests := []struct {
		name       string
		input      CreateTransferInput
		setupMock  func(m *mockTransferRepo)
		wantErr    error
		wantResult *TransferResult
	}{
		{
			name: "Gagal - Idempotency Key Kosong",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "",
				FromAccountID:  1,
				ToAccountID:    2,
				Amount:         50000,
			},
			setupMock: nil,
			wantErr:   ErrMissingKey,
		},
		{
			name: "Gagal - Nominal Kurang Dari atau Sama Dengan Nol",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "uuid-123",
				FromAccountID:  1,
				ToAccountID:    2,
				Amount:         0,
			},
			setupMock: nil,
			wantErr:   ErrInvalidAmount,
		},
		{
			name: "Gagal - Transfer ke Akun yang Sama",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "uuid-123",
				FromAccountID:  1,
				ToAccountID:    1,
				Amount:         50000,
			},
			setupMock: nil,
			wantErr:   ErrSameAccount,
		},
		{
			name: "Sukses - Idempotency Cache Hit (Request Ulang)",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "uuid-cached",
				FromAccountID:  1,
				ToAccountID:    2,
				Amount:         25000,
			},
			setupMock: func(m *mockTransferRepo) {
				m.tryAcquireFunc = func(ctx context.Context, key string, userID int64) (*IdempotencyRecord, bool, error) {
					respCode := 201
					return &IdempotencyRecord{
						Key:          key,
						UserID:       userID,
						Status:       IdempotencyStatusCompleted,
						ResponseCode: &respCode,
						ResponseBody: cachedJSON,
					}, false, nil
				}
				// ExecuteTransfer tidak boleh dipanggil saat cache hit
				m.executeTxFunc = func(ctx context.Context, tr *Transfer) (int64, error) {
					t.Fatal("ExecuteTransfer should not be called on cache hit")
					return 0, nil
				}
			},
			wantErr:    nil,
			wantResult: &sampleCachedResult,
		},
		{
			name: "Gagal - Idempotency Conflict (Request Sedang Berjalan)",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "uuid-running",
				FromAccountID:  1,
				ToAccountID:    2,
				Amount:         50000,
			},
			setupMock: func(m *mockTransferRepo) {
				m.tryAcquireFunc = func(ctx context.Context, key string, userID int64) (*IdempotencyRecord, bool, error) {
					// isNew = false dan status masih PENDING
					return &IdempotencyRecord{
						Key:    key,
						UserID: userID,
						Status: IdempotencyStatusPending,
					}, false, nil
				}
			},
			wantErr: ErrIdempotencyConflict,
		},
		{
			name: "Gagal - Saldo Pengirim Tidak Cukup",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "uuid-insufficient",
				FromAccountID:  1,
				ToAccountID:    2,
				Amount:         500000,
			},
			setupMock: func(m *mockTransferRepo) {
				m.executeTxFunc = func(ctx context.Context, tr *Transfer) (int64, error) {
					return 0, ErrInsufficientBalance
				}
			},
			wantErr: ErrInsufficientBalance,
		},
		{
			name: "Sukses - Transfer Berhasil Dieksekusi",
			input: CreateTransferInput{
				UserID:         1,
				IdempotencyKey: "uuid-success",
				FromAccountID:  1,
				ToAccountID:    2,
				Amount:         50000,
				Notes:          "Bayar hutang",
			},
			setupMock: func(m *mockTransferRepo) {
				m.executeTxFunc = func(ctx context.Context, tr *Transfer) (int64, error) {
					tr.ID = 101
					return 150000, nil // Sisa saldo pengirim menjadi 150.000
				}
			},
			wantErr: nil,
			wantResult: &TransferResult{
				Transfer: Transfer{
					ID:            101,
					UserID:        1,
					FromAccountID: 1,
					ToAccountID:   2,
					Amount:        50000,
					Notes:         "Bayar hutang",
				},
				FromAccountBalance: 150000,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockTransferRepo{}
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			svc := NewService(mockRepo)
			res, err := svc.Processtransfer(context.Background(), tc.input)

			// Validasi Error
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}

			// Validasi Sukses
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.FromAccountBalance != tc.wantResult.FromAccountBalance {
				t.Errorf("expected balance %d, got %d", tc.wantResult.FromAccountBalance, res.FromAccountBalance)
			}
			if res.Transfer.ID != tc.wantResult.Transfer.ID {
				t.Errorf("expected transfer ID %d, got %d", tc.wantResult.Transfer.ID, res.Transfer.ID)
			}
		})
	}

}
