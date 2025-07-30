package service

import (
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/pkg/crypt"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type User struct {
	ID             int64
	Fingerprint    string
	PublicKey      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PublicKeyBytes []byte
}

func newUserFromQueryUser(user query.User) (*User, error) {
	publicKeyBytes, err := crypt.DecodePublicKey(user.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	return &User{
		ID:             user.ID,
		Fingerprint:    user.Fingerprint,
		PublicKey:      user.PublicKey,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		PublicKeyBytes: publicKeyBytes,
	}, nil
}

type Secret struct {
	ID        string
	UserID    int64
	Crypt     []byte
	Meta      []byte
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

func newSecretFromQuerySecret(secret query.Secret) *Secret {
	return &Secret{
		ID:        secret.ID,
		UserID:    secret.UserID,
		Crypt:     secret.Crypt,
		Meta:      secret.Meta,
		Data:      secret.Data,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}
}
