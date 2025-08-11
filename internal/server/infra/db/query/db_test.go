package query

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	assert.NotNil(t, queries)
	assert.Equal(t, db, queries.db)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQueries_WithTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	queries := New(db)
	ctx := context.Background()
	now := time.Now()

	// Begin a transaction
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)

	// Create a new Queries instance with the transaction
	txQueries := queries.WithTx(tx)
	assert.NotNil(t, txQueries)
	assert.Equal(t, tx, txQueries.db)

	// Test a simple query with the transaction
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

	_, err = txQueries.GetUserByFingerprint(ctx, fingerprint)
	assert.NoError(t, err)

	// Commit the transaction
	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}
