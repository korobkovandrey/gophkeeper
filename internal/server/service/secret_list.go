package service

import (
	"context"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
)

// SecretLister defines the interface for listing secrets for a user.
type SecretLister interface {
	ListSecrets(ctx context.Context, userID int64) ([]query.Secret, error)
}

type ListSecretsFunc func(ctx context.Context, userID int64) ([]*Secret, error)

// NewListSecretsFunc returns a function that handles listing secrets for a user.
func NewListSecretsFunc(lister SecretLister) func(ctx context.Context, userID int64) ([]*Secret, error) {
	return func(ctx context.Context, userID int64) ([]*Secret, error) {
		qSecrets, err := lister.ListSecrets(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}
		secrets := make([]*Secret, len(qSecrets))
		for i := range qSecrets {
			secrets[i] = newSecretFromQuerySecret(qSecrets[i])
		}
		return secrets, nil
	}
}
