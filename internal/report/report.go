package report

import "time"

type CategoryExpenseBreakdown struct {
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	TotalSpent   int64  `json:"total_spent"`
}

type MonthlySummaryResponse struct {
	MonthPeriod   time.Time                  `json:"month_period"`
	TotalIncome   int64                      `json:"total_income"`
	TotalExpense  int64                      `json:"total_expense"`
	NetSavings    int64                      `json:"net_savings"`
	ExpenseByType []CategoryExpenseBreakdown `json:"expense_by_type"`
}
