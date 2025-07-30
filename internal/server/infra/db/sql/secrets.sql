-- name: FindSecret :one
SELECT id, user_id, crypt, meta, data, created_at, updated_at FROM secrets WHERE id=$1 AND user_id=$2 LIMIT 1;

-- name: ListSecrets :many
SELECT id, user_id, crypt, meta, data, created_at, updated_at FROM secrets WHERE user_id=$1;

-- name: CreateSecret :one
INSERT INTO secrets (id, user_id, crypt, meta, data, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (id, user_id) DO UPDATE SET crypt = excluded.crypt, meta = excluded.meta, data = excluded.data, updated_at = excluded.updated_at WHERE secrets.updated_at <= excluded.updated_at
RETURNING id, user_id, crypt, meta, data, created_at, updated_at;

-- name: UpdateSecret :one
UPDATE secrets
SET id=$1, crypt=$2, meta=$3, data=$4, updated_at=$5
WHERE id=$6 AND user_id=$7 AND updated_at <= $5
RETURNING id, user_id, crypt, meta, data, created_at, updated_at;

-- name: DeleteSecret :one
DELETE FROM secrets WHERE id=$1 AND user_id=$2 AND updated_at <= $3
RETURNING id, user_id, crypt, meta, data, created_at, updated_at;