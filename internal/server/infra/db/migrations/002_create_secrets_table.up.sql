CREATE TABLE IF NOT EXISTS secrets
(
    id VARCHAR(255),
    user_id BIGINT NOT NULL REFERENCES users ON DELETE CASCADE,
    crypt BYTEA NULL,
    meta BYTEA NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMP(0) NOT NULL,
    updated_at TIMESTAMP(0) NOT NULL,
    PRIMARY KEY (id, user_id)
);
