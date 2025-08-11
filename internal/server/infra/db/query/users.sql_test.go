package query

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestQueries_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	params := CreateUserParams{
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

	result, err := queries.CreateUser(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_DeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	userID := int64(1)

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = queries.DeleteUser(ctx, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_GetUserByFingerprint(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()
	fingerprint := "user-fingerprint"

	expectedUser := User{
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

	result, err := queries.GetUserByFingerprint(ctx, fingerprint)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_GetUserById(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()
	userID := int64(1)

	expectedUser := User{
		ID:          userID,
		Fingerprint: "user-fingerprint",
		PublicKey:   "public-key",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, fingerprint, public_key, created_at, updated_at FROM users WHERE id=$1 LIMIT 1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "fingerprint", "public_key", "created_at", "updated_at"}).
			AddRow(
				expectedUser.ID,
				expectedUser.Fingerprint,
				expectedUser.PublicKey,
				expectedUser.CreatedAt,
				expectedUser.UpdatedAt,
			))

	result, err := queries.GetUserById(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_UpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	params := UpdateUserParams{
		Fingerprint: "new-fingerprint",
		PublicKey:   "new-public-key",
		UpdatedAt:   now,
		ID:          1,
	}

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE users
		SET fingerprint=$1, public_key = $2, updated_at = $3
		WHERE id = $4`)).
		WithArgs(params.Fingerprint, params.PublicKey, params.UpdatedAt, params.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = queries.UpdateUser(ctx, params)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}
