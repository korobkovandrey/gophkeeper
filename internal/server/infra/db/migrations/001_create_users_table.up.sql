CREATE TABLE IF NOT EXISTS users
(
    id BIGSERIAL PRIMARY KEY,
    fingerprint VARCHAR(255) NOT NULL CONSTRAINT users_fingerprint_unique UNIQUE,
    public_key TEXT NOT NULL,
    created_at TIMESTAMP(0) NOT NULL,
    updated_at TIMESTAMP(0) NOT NULL
);
