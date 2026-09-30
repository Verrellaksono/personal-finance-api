package budget

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

func (r *Repository) Upsert(ctx context.Context, b *Budget) error {
	query := `
			INSERT INTO budgets (user_id, category_id, amount, month_period, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (user_id, category_id, month_period)
			DO UPDATE SET amount = EXCLUDED.amount, updated_at = NOW()
			RETURNING id, created_at, updated_at;
	`

	if err := r.db.QueryRowContext(ctx, query, b.UserID, b.CategoryID, b.Amount, b.MonthPeriod).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return fmt.Errorf("failed to upsert budget: %w", err)
	}

	return nil
}

func (r *Repository) GetBudgetProgress(ctx context.Context, userID int64, monthPeriod time.Time) ([]BudgetProgress, error) {
	query := `
			SELECT
				b.id,
				b.category_id,
				b.amount AS limit_amount,
				COALESCE((
					SELECT SUM(amount) FROM transactions
					WHERE category_id = b.category_id
						AND TYPE = 'EXPENSE'
						AND transaction_date >= b.month_period
						AND transaction_date < b.month_period + INTERVAL '1 month'
				), 0) AS total_spent,
				b.month_period
			FROM budgets b
			WHERE b.user_id = $1 AND b.month_period = $2;
	`

	rows, err := r.db.QueryContext(ctx, query, userID, monthPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to query budget progress: %w", err)
	}
	defer rows.Close()

	var results []BudgetProgress
	for rows.Next() {
		var bp BudgetProgress
		if err := rows.Scan(&bp.BudgetID, &bp.CategoryID, &bp.LimitAmount, &bp.TotalSpent, &bp.MonthPeriod); err != nil {
			return nil, fmt.Errorf("failed to scan budget progress: %w", err)
		}

		results = append(results, bp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}
