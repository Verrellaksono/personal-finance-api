package budget

import "time"

type Budget struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	CategoryID  int64     `json:"category_id"`
	Amount      int64     `json:"amount"`
	MonthPeriod time.Time `json:"month_period"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SetBudgetInput struct {
	UserID     int64  `json:"-"`
	CategoryID int64  `json:"category_id"`
	Amount     int64  `json:"amount"`
	Month      string `json:"month"` // Format yang diminta dari client: "YYYY-MM" (misal: "2026-10")
}

type BudgetProgress struct {
	BudgetID    int64     `json:"budget_id"`
	CategoryID  int64     `json:"category_id"`
	LimitAmount int64     `json:"limit_amount"`
	TotalSpent  int64     `json:"total_spent"`
	MonthPeriod time.Time `json:"month_period"`
}
