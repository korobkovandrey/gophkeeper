package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
	"time"
)

type DeleteSecretFunc func(ctx context.Context, userID int64, id string, updatedAt time.Time) (*Secret, error)

// NewDeleteSecretFunc returns a function that handles deleting a secret.
func NewDeleteSecretFunc(deleter secretDeleter) DeleteSecretFunc {
	return func(ctx context.Context, userID int64, id string, updatedAt time.Time) (*Secret, error) {
		qSecret, err := deleter.DeleteSecret(ctx, query.DeleteSecretParams{
			UserID:    userID,
			ID:        id,
			UpdatedAt: updatedAt.UTC(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = ErrConflict
			}
			return nil, fmt.Errorf("failed to delete: %w", err)
		}
		return newSecretFromQuerySecret(qSecret), nil
	}
}
