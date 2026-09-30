DROP INDEX IF EXISTS idx_categories_user_id;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS fk_transactions_category_id;

DROP TABLE IF EXISTS budgets;
DROP TABLE IF EXISTS categories;