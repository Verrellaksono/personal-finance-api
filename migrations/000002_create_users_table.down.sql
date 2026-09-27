-- Lepas Foreign Key terlebih dahulu
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS fk_accounts_user_id;

-- Hapus tabel users
DROP TABLE IF EXISTS users;