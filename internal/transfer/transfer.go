package transfer

import "time"

// Status untuk Idempotency Key
const (
	IdempotencyStatusPending   = "PENDING"
	IdempotencyStatusCompleted = "COMPLETED"
	IdempotencyStatusFailed    = "FAILED"
)

type Transfer struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"uset_id"`
	FromAccountID int64     `json:"from_account_id"`
	ToAccountID   int64     `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

type TransferResult struct {
	Transfer           Transfer `json:"transfer"`
	FromAccountBalance int64    `json:"from_account_balance"`
}

type CreateTransferInput struct {
	UserID         int64  `json:"-"`
	IdempotencyKey string `json:"-"`
	FromAccountID  int64  `json:"from_account_id"`
	ToAccountID    int64  `json:"to_account_id"`
	Amount         int64  `json:"amount"`
	Notes          string `json:"notes"`
}

type IdempotencyRecord struct {
	Key          string
	UserID       int64
	Status       string
	ResponseCode *int
	ResponseBody []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
