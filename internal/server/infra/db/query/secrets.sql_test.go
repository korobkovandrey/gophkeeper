package query

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestQueries_CreateSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	params := CreateSecretParams{
		ID:        "secret-1",
		UserID:    1,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("meta-data"),
		Data:      []byte("raw-data"),
		CreatedAt: now,
	}

	expectedSecret := Secret{
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

	result, err := queries.CreateSecret(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_DeleteSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	params := DeleteSecretParams{
		ID:        "secret-1",
		UserID:    1,
		UpdatedAt: now,
	}

	expectedSecret := Secret{
		ID:        params.ID,
		UserID:    params.UserID,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("meta-data"),
		Data:      []byte("raw-data"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`DELETE FROM secrets WHERE id=$1 AND user_id=$2 AND updated_at <= $3
		RETURNING id, user_id, crypt, meta, data, created_at, updated_at`)).
		WithArgs(params.ID, params.UserID, params.UpdatedAt).
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

	result, err := queries.DeleteSecret(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_FindSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	params := FindSecretParams{
		ID:     "secret-1",
		UserID: 1,
	}

	expectedSecret := Secret{
		ID:        params.ID,
		UserID:    params.UserID,
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

	result, err := queries.FindSecret(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_ListSecrets(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()
	userID := int64(1)

	expectedSecrets := []Secret{
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

	result, err := queries.ListSecrets(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecrets, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_UpdateSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	params := UpdateSecretParams{
		ID:        "new-secret-id",
		Crypt:     []byte("new-encrypted-data"),
		Meta:      []byte("new-meta-data"),
		Data:      []byte("new-raw-data"),
		UpdatedAt: now,
		ID_2:      "secret-1",
		UserID:    1,
	}

	expectedSecret := Secret{
		ID:        params.ID,
		UserID:    params.UserID,
		Crypt:     params.Crypt,
		Meta:      params.Meta,
		Data:      params.Data,
		CreatedAt: now.Add(-time.Hour), // Assume created_at is older
		UpdatedAt: params.UpdatedAt,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`UPDATE secrets
		SET id=$1, crypt=$2, meta=$3, data=$4, updated_at=$5
		WHERE id=$6 AND user_id=$7 AND updated_at <= $5
		RETURNING id, user_id, crypt, meta, data, created_at, updated_at`)).
		WithArgs(
			params.ID,
			params.Crypt,
			params.Meta,
			params.Data,
			params.UpdatedAt,
			params.ID_2,
			params.UserID,
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

	result, err := queries.UpdateSecret(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}
