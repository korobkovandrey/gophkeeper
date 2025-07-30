package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
	"time"
)

type secretRepository interface {
	FindSecret(ctx context.Context, arg query.FindSecretParams) (query.Secret, error)
	ListSecrets(ctx context.Context, userID int64) ([]query.Secret, error)
	CreateSecret(ctx context.Context, arg query.CreateSecretParams) (query.Secret, error)
	UpdateSecret(ctx context.Context, arg query.UpdateSecretParams) (query.Secret, error)
	DeleteSecret(ctx context.Context, arg query.DeleteSecretParams) (query.Secret, error)
}

type SecretService struct {
	r secretRepository
}

func NewSecretService(r secretRepository) *SecretService {
	return &SecretService{
		r: r,
	}
}

func (s *SecretService) Find(ctx context.Context, userID int64, id string) (*Secret, error) {
	qSecret, err := s.r.FindSecret(ctx, query.FindSecretParams{
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

func (s *SecretService) List(ctx context.Context, userID int64) ([]*Secret, error) {
	qSecrets, err := s.r.ListSecrets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	var secrets []*Secret
	for _, qSecret := range qSecrets {
		secrets = append(secrets, newSecretFromQuerySecret(qSecret))
	}
	return secrets, nil
}

func (s *SecretService) Store(ctx context.Context, userID int64, id, newID string, crypt, meta, data []byte, storeTime time.Time) (*Secret, error) {
	qSecret, err := s.r.FindSecret(ctx, query.FindSecretParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to find: %w", err)
		}
		qSecret, err = s.r.CreateSecret(ctx, query.CreateSecretParams{
			ID:        newID,
			UserID:    userID,
			Crypt:     crypt,
			Meta:      meta,
			Data:      data,
			CreatedAt: storeTime,
		})
		if err == nil {
			return newSecretFromQuerySecret(qSecret), nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to create: %w", err)
		}
	}
	qSecret, err = s.r.UpdateSecret(ctx, query.UpdateSecretParams{
		ID:        id,
		ID_2:      newID,
		UserID:    userID,
		Crypt:     crypt,
		Meta:      meta,
		Data:      data,
		UpdatedAt: storeTime,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrConflict
		}
		return nil, fmt.Errorf("failed to update: %w", err)
	}
	return newSecretFromQuerySecret(qSecret), nil
}

func (s *SecretService) Delete(ctx context.Context, userID int64, id string, updatedAt time.Time) (*Secret, error) {
	qSecret, err := s.r.DeleteSecret(ctx, query.DeleteSecretParams{
		UserID:    userID,
		ID:        id,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrConflict
		}
		return nil, fmt.Errorf("failed to delete: %w", err)
	}
	return newSecretFromQuerySecret(qSecret), nil
}
