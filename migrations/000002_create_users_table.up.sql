CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- Seed users
-- Password hash di bawah adalah bcrypt hash dari "password123"
INSERT INTO users(id, email, password_hash)
VALUES(1, 'initial.user@example.com', '$2a$10$7z78Fj5qL2J0Q1/xQZ5xUO5gR7z3L4c4Q6eG2F0F0c4e1Q1Q1Q1Q1')
ON CONFLICT (id) DO NOTHING;

-- Sinkronkan sequence ID postgres agar registrasi user berikutnya mulai dari ID 2
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

-- Pasang Foreign Key constraint ke tabel accounts
ALTER TABLE accounts
ADD CONSTRAINT fk_account_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;