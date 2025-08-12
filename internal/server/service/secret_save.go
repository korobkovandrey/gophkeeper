package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
	"time"
)

// SecretFinder defines the interface for finding a secret by ID and user ID.
type SecretFinder interface {
	FindSecret(ctx context.Context, arg query.FindSecretParams) (query.Secret, error)
}

// SecretCreator defines the interface for creating a secret.
type SecretCreator interface {
	CreateSecret(ctx context.Context, arg query.CreateSecretParams) (query.Secret, error)
}

// SecretUpdater defines the interface for updating a secret.
type SecretUpdater interface {
	UpdateSecret(ctx context.Context, arg query.UpdateSecretParams) (query.Secret, error)
}

// SecretDeleter defines the interface for deleting a secret.
type SecretDeleter interface {
	DeleteSecret(ctx context.Context, arg query.DeleteSecretParams) (query.Secret, error)
}

type SecretRepository interface {
	SecretFinder
	SecretCreator
	SecretUpdater
	SecretDeleter
}

type SaveSecretFunc func(ctx context.Context, userID int64, id, newID string,
	crypt, meta, data []byte, storeTime time.Time) (*Secret, error)

// NewSaveSecretFunc returns a function that handles storing a secret.
func NewSaveSecretFunc(r SecretRepository) SaveSecretFunc {
	return func(ctx context.Context, userID int64, id, newID string, crypt, meta, data []byte, storeTime time.Time) (*Secret, error) {
		_, err := r.FindSecret(ctx, query.FindSecretParams{
			ID:     id,
			UserID: userID,
		})
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to find: %w", err)
			}
			qSecret, err := r.CreateSecret(ctx, query.CreateSecretParams{
				ID:        newID,
				UserID:    userID,
				Crypt:     crypt,
				Meta:      meta,
				Data:      data,
				CreatedAt: storeTime.UTC(),
			})
			if err == nil {
				return newSecretFromQuerySecret(qSecret), nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to create: %w", err)
			}
		}
		if id != newID {
			_, err := r.FindSecret(ctx, query.FindSecretParams{
				ID:     newID,
				UserID: userID,
			})
			if err == nil {
				_, err = r.DeleteSecret(ctx, query.DeleteSecretParams{
					UserID:    userID,
					ID:        newID,
					UpdatedAt: storeTime.UTC(),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to delete: %w", err)
				}
			} else if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to find: %w", err)
			}
		}
		qSecret, err := r.UpdateSecret(ctx, query.UpdateSecretParams{
			ID:        newID,
			ID_2:      id,
			UserID:    userID,
			Crypt:     crypt,
			Meta:      meta,
			Data:      data,
			UpdatedAt: storeTime.UTC(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = ErrConflict
			}
			return nil, fmt.Errorf("failed to update: %w, %v, %v, %v", err, id, newID, storeTime.UTC())
		}
		return newSecretFromQuerySecret(qSecret), nil
	}
}
