package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
)

type FindSecretFunc func(ctx context.Context, userID int64, id string) (*Secret, error)

// NewFindSecretFunc returns a function that handles finding a secret by ID and user ID.
func NewFindSecretFunc(finder secretFinder) FindSecretFunc {
	return func(ctx context.Context, userID int64, id string) (*Secret, error) {
		qSecret, err := finder.FindSecret(ctx, query.FindSecretParams{
			ID:     id,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = ErrNotFound
			}
			return nil, fmt.Errorf("failed to find secret %s: %w", id, err)
		}
		return newSecretFromQuerySecret(qSecret), nil
	}
}
