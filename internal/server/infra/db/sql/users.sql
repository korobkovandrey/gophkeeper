-- name: GetUserByFingerprint :one
SELECT id, fingerprint, public_key, created_at, updated_at FROM users WHERE fingerprint=$1 LIMIT 1;

-- name: GetUserById :one
SELECT id, fingerprint, public_key, created_at, updated_at FROM users WHERE id=$1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (fingerprint, public_key, created_at, updated_at)
VALUES ($1, $2, $3, $3)
RETURNING id;

-- name: UpdateUser :exec
UPDATE users
SET fingerprint=$1, public_key = $2, updated_at = $3
WHERE id = $4;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;