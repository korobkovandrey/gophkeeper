package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
)

// UserFinderByFingerprint defines the interface for finding a user by fingerprint.
type UserFinderByFingerprint interface {
	GetUserByFingerprint(ctx context.Context, fingerprint string) (query.User, error)
}

type FindUserByFingerprintFunc func(ctx context.Context, fingerprint string) (*User, error)

// NewFindUserByFingerprintFunc returns a function that handles finding a user by fingerprint.
func NewFindUserByFingerprintFunc(finderByFingerprint UserFinderByFingerprint) FindUserByFingerprintFunc {
	return func(ctx context.Context, fingerprint string) (*User, error) {
		qUser, err := finderByFingerprint.GetUserByFingerprint(ctx, fingerprint)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = ErrNotFound
			}
			return nil, fmt.Errorf("failed to get user #%s: %w", fingerprint, err)
		}
		user, err := newUserFromQueryUser(qUser)
		if err != nil {
			return nil, fmt.Errorf("failed to new user: %w", err)
		}
		return user, nil
	}
}
