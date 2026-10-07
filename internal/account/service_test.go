package account

import (
	"context"
	"errors"
	"testing"
)

type mockAccountRepo struct {
	createFunc  func(ctx context.Context, acc *Account) error
	getByIdFunc func(ctx context.Context, id int64) (*Account, error)
}

func (m *mockAccountRepo) Create(ctx context.Context, acc *Account) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, acc)
	}

	return nil
}

func (m *mockAccountRepo) GetById(ctx context.Context, id int64) (*Account, error) {
	if m.getByIdFunc != nil {
		return m.getByIdFunc(ctx, id)
	}

	return nil, nil
}

func TestCreateAccount(t *testing.T) {
	// Daftar Skenario Tes
	tests := []struct {
		name       string
		input      CreateAccountInput
		mockCreate func(ctx context.Context, acc *Account) error
		wantErr    error
	}{
		{
			name: "Gagal - Nama akun kosong",
			input: CreateAccountInput{
				UserID:         1,
				Name:           "",
				Type:           "BANK",
				InitialBalance: 100000,
			},
			mockCreate: nil,
			wantErr:    ErrEmptyName,
		},
		{
			name: "Gagal - Tipe akun tidak valid",
			input: CreateAccountInput{
				UserID:         1,
				Name:           "Tabungan",
				Type:           "INVALID_TYPE",
				InitialBalance: 100000,
			},
			mockCreate: nil,
			wantErr:    ErrInvalidType,
		},
		{
			name: "Gagal - Saldo awal minus",
			input: CreateAccountInput{
				UserID:         1,
				Name:           "Tabungan",
				Type:           "BANK",
				InitialBalance: -50000,
			},
			mockCreate: nil,
			wantErr:    ErrInvalidBalance,
		},
		{
			name: "Sukses - Data Valid",
			input: CreateAccountInput{
				UserID:         1,
				Name:           "BCA Tabungan",
				Type:           "BANK",
				InitialBalance: 500000,
			},
			mockCreate: func(ctx context.Context, acc *Account) error {
				acc.ID = 10
				return nil
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Siapkan mock repo
			mockRepo := &mockAccountRepo{
				createFunc: tc.mockCreate,
			}

			// Masukkan mock repo ke service
			service := NewService(mockRepo)

			// Periksa apakah error-nya sesuai dengan ekspektasi kita
			result, err := service.CreateAccount(context.Background(), tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected error %v, got %v", tc.wantErr, err)
			}

			// Jika skenario sukses, pastikan datanya benar
			if tc.wantErr == nil && result.Name != tc.input.Name {
				t.Errorf("expected result name %s, got %s", tc.input.Name, result.Name)
			}
		})
	}
}

func TestGetAccountByID(t *testing.T) {
	tests := []struct {
		name             string
		accountID        int64
		requestingUserID int64
		mockGetByID      func(ctx context.Context, id int64) (*Account, error)
		wantErr          error
	}{
		{
			name:             "Gagal - ID Akun kurang dari atau sama dengan 0",
			accountID:        0,
			requestingUserID: 1,
			mockGetByID:      nil,
			wantErr:          errors.New("invalid account id"),
		},
		{
			name:             "Gagal - Akun tidak ditemukan di database",
			accountID:        99,
			requestingUserID: 1,
			mockGetByID: func(ctx context.Context, id int64) (*Account, error) {
				return nil, ErrNotFound
			},
			wantErr: ErrNotFound,
		},
		{
			name:             "Gagal - Proteksi IDOR (Pengguna mencoba akses akun orang lain)",
			accountID:        10,
			requestingUserID: 2,
			mockGetByID: func(ctx context.Context, id int64) (*Account, error) {
				return &Account{
					ID:     10,
					UserID: 1,
					Name:   "Tabungan Rahasia user 1",
				}, nil
			},
			wantErr: ErrForbidden,
		},
		{
			name:             "Sukses - Akun milik user yang sama",
			accountID:        10,
			requestingUserID: 1,
			mockGetByID: func(ctx context.Context, id int64) (*Account, error) {
				return &Account{
					ID:      10,
					UserID:  1,
					Name:    "Dompet Utama",
					Balance: 75000,
				}, nil
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockAccountRepo{
				getByIdFunc: tc.mockGetByID,
			}

			service := NewService(mockRepo)

			result, err := service.GetAccountByID(context.Background(), tc.accountID, tc.requestingUserID)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, git nil", tc.wantErr)
				}
				if errors.Is(tc.wantErr, ErrForbidden) || errors.Is(tc.wantErr, ErrNotFound) {
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf("expected error %v, git %v", tc.wantErr, err)
					}
				} else if err.Error() != tc.wantErr.Error() {
					t.Errorf("expected error message %q, got %q", tc.wantErr.Error(), err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tc.accountID || result.UserID != tc.requestingUserID {
				t.Errorf("expected account ID %d and UserID %d, got ID %d and UserID %d",
					tc.accountID, tc.requestingUserID, result.ID, result.UserID)
			}
		})
	}
}
