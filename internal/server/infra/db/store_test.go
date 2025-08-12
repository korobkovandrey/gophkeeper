package db

import (
	"context"
	"regexp"
	"testing"
	"time"

	"gophkeeper/internal/server/infra/db/query"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_CreateSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	store := newStore(db)
	defer func() {
		assert.NoError(t, store.Close())
	}()
	ctx := context.Background()
	now := time.Now()

	params := query.CreateSecretParams{
		ID:        "secret-1",
		UserID:    1,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("meta-data"),
		Data:      []byte("raw-data"),
		CreatedAt: now,
	}

	expectedSecret := query.Secret{
		ID:        params.ID,
		UserID:    params.UserID,
		Crypt:     params.Crypt,
		Meta:      params.Meta,
		Data:      params.Data,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.CreatedAt,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO secrets (id, user_id, crypt, meta, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (id, user_id) DO UPDATE SET crypt = excluded.crypt, meta = excluded.meta, data = excluded.data, updated_at = excluded.updated_at
		WHERE secrets.updated_at <= excluded.updated_at
		RETURNING id, user_id, crypt, meta, data, created_at, updated_at`)).
		WithArgs(
			params.ID,
			params.UserID,
			params.Crypt,
			params.Meta,
			params.Data,
			params.CreatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "crypt", "meta", "data", "created_at", "updated_at"}).
			AddRow(
				expectedSecret.ID,
				expectedSecret.UserID,
				expectedSecret.Crypt,
				expectedSecret.Meta,
				expectedSecret.Data,
				expectedSecret.CreatedAt,
				expectedSecret.UpdatedAt,
			))

	result, err := store.CreateSecret(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestStore_FindSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	store := newStore(db)
	defer func() {
		assert.NoError(t, store.Close())
	}()
	ctx := context.Background()
	now := time.Now()

	params := query.FindSecretParams{
		ID:     "secret-1",
		UserID: 1,
	}

	expectedSecret := query.Secret{
		ID:        "secret-1",
		UserID:    1,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("meta-data"),
		Data:      []byte("raw-data"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, user_id, crypt, meta, data, created_at, updated_at FROM secrets WHERE id=$1 AND user_id=$2 LIMIT 1`)).
		WithArgs(params.ID, params.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "crypt", "meta", "data", "created_at", "updated_at"}).
			AddRow(
				expectedSecret.ID,
				expectedSecret.UserID,
				expectedSecret.Crypt,
				expectedSecret.Meta,
				expectedSecret.Data,
				expectedSecret.CreatedAt,
				expectedSecret.UpdatedAt,
			))

	result, err := store.FindSecret(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestStore_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	store := newStore(db)
	defer func() {
		assert.NoError(t, store.Close())
	}()
	ctx := context.Background()
	now := time.Now()

	params := query.CreateUserParams{
		Fingerprint: "user-fingerprint",
		PublicKey:   "public-key",
		CreatedAt:   now,
	}

	expectedID := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO users (fingerprint, public_key, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id`)).
		WithArgs(params.Fingerprint, params.PublicKey, params.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	result, err := store.CreateUser(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestStore_GetUserByFingerprint(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	store := newStore(db)
	defer func() {
		assert.NoError(t, store.Close())
	}()
	ctx := context.Background()
	now := time.Now()

	fingerprint := "user-fingerprint"

	expectedUser := query.User{
		ID:          1,
		Fingerprint: fingerprint,
		PublicKey:   "public-key",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, fingerprint, public_key, created_at, updated_at FROM users WHERE fingerprint=$1 LIMIT 1`)).
		WithArgs(fingerprint).
		WillReturnRows(sqlmock.NewRows([]string{"id", "fingerprint", "public_key", "created_at", "updated_at"}).
			AddRow(
				expectedUser.ID,
				expectedUser.Fingerprint,
				expectedUser.PublicKey,
				expectedUser.CreatedAt,
				expectedUser.UpdatedAt,
			))

	result, err := store.GetUserByFingerprint(ctx, fingerprint)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestStore_ListSecrets(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	store := newStore(db)
	defer func() {
		assert.NoError(t, store.Close())
	}()
	ctx := context.Background()
	now := time.Now()
	userID := int64(1)

	expectedSecrets := []query.Secret{
		{
			ID:        "secret-1",
			UserID:    userID,
			Crypt:     []byte("encrypted-data-1"),
			Meta:      []byte("meta-data-1"),
			Data:      []byte("raw-data-1"),
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "secret-2",
			UserID:    userID,
			Crypt:     []byte("encrypted-data-2"),
			Meta:      []byte("meta-data-2"),
			Data:      []byte("raw-data-2"),
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, user_id, crypt, meta, data, created_at, updated_at FROM secrets WHERE user_id=$1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "crypt", "meta", "data", "created_at", "updated_at"}).
			AddRow(
				expectedSecrets[0].ID,
				expectedSecrets[0].UserID,
				expectedSecrets[0].Crypt,
				expectedSecrets[0].Meta,
				expectedSecrets[0].Data,
				expectedSecrets[0].CreatedAt,
				expectedSecrets[0].UpdatedAt,
			).
			AddRow(
				expectedSecrets[1].ID,
				expectedSecrets[1].UserID,
				expectedSecrets[1].Crypt,
				expectedSecrets[1].Meta,
				expectedSecrets[1].Data,
				expectedSecrets[1].CreatedAt,
				expectedSecrets[1].UpdatedAt,
			))

	result, err := store.ListSecrets(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecrets, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}
