package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/pkg/crypt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

// userCreator defines the interface for creating a user.
type userCreator interface {
	CreateUser(ctx context.Context, params query.CreateUserParams) (userID int64, err error)
}

// userFinderByFingerprint defines the interface for finding a user by fingerprint.
type userFinderByFingerprint interface {
	GetUserByFingerprint(ctx context.Context, fingerprint string) (query.User, error)
}

// RegisterUserFunc is a function that handles user registration.
type RegisterUserFunc func(ctx context.Context, publicKeyBytes []byte) (int64, error)

type registerRepository interface {
	userCreator
	userFinderByFingerprint
}

// NewRegisterUserFunc returns a function that handles user registration.
func NewRegisterUserFunc(r registerRepository) RegisterUserFunc {
	return func(ctx context.Context, publicKeyBytes []byte) (id int64, err error) {
		fingerprint := crypt.Fingerprint(publicKeyBytes)
		_, err = r.GetUserByFingerprint(ctx, fingerprint)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("failed to get user by fingerprint: %w", err)
			}
		} else {
			return 0, fmt.Errorf("%w: fingerprint already exists", ErrConflict)
		}
		id, err = r.CreateUser(ctx, query.CreateUserParams{
			Fingerprint: crypt.Fingerprint(publicKeyBytes),
			PublicKey:   crypt.EncodePublicKey(publicKeyBytes),
			CreatedAt:   time.Now(),
		})
		if err != nil {
			var e *pgconn.PgError
			if errors.As(err, &e) && pgerrcode.IsIntegrityConstraintViolation(e.Code) {
				return 0, fmt.Errorf("%w: public key", ErrConflict)
			}
			return 0, fmt.Errorf("failed to create user: %w", err)
		}
		return id, nil
	}
}
