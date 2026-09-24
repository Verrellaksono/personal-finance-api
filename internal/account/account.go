package account

import "time"

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
