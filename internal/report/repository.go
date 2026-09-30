package report

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetMonthlyTotals(ctx context.Context, userID int64, monthPeriod time.Time) (totalIncome int64, totalExpense int64, err error) {
	query := `
			SELECT 
				COALESCE(SUM(CASE WHEN t.type = 'INCOME' THEN t.amount ELSE 0 END), 0) as total_income,
				COALESCE(SUM(CASE WHEN t.type = 'EXPENSE' THEN t.amount ELSE 0 END), 0) as total_expense
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE a.user_id = $1
				AND t.transaction_date >= $2
				AND t.transaction_date < $2 + INTERVAL '1 month';
	`

	if err := r.db.QueryRowContext(ctx, query, userID, monthPeriod).Scan(&totalIncome, &totalExpense); err != nil {
		return 0, 0, fmt.Errorf("failed to aggregate monthly totals: %w", err)
	}

	return totalIncome, totalExpense, err
}

func (r *Repository) GetExpenseByCategoryBreakdown(ctx context.Context, userID int64, monthPeriod time.Time) ([]CategoryExpenseBreakdown, error) {
	query := `
			SELECT
				c.id AS category_id,
				c.name AS category_name,
				SUM(t.amount) AS total_spent
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			JOIN categories c ON t.category_id = c.id
			WHERE a.user_id = $1
				AND t.type = 'EXPENSE'
				AND t.transaction_date >= $2
				AND t.transaction_date < $2 + INTERVAl '1 month'
			GROUP BY c.id, c.name
			ORDER BY total_spent DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, userID, monthPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense breakdown: %w", err)
	}
	defer rows.Close()

	results := make([]CategoryExpenseBreakdown, 0)
	for rows.Next() {
		var item CategoryExpenseBreakdown
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.TotalSpent); err != nil {
			return nil, fmt.Errorf("failed to scan expense breakdown: %w", err)
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate expense breakdown rows: %w", err)
	}

	return results, nil
}
