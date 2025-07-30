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

type userRepository interface {
	CreateUser(context.Context, query.CreateUserParams) (userID int64, err error)
	GetUserByFingerprint(ctx context.Context, fingerprint string) (query.User, error)
	GetUserById(ctx context.Context, id int64) (query.User, error)
}

type UserService struct {
	r userRepository
}

func NewUserService(r userRepository) *UserService {
	return &UserService{
		r: r,
	}
}

func (s *UserService) Register(ctx context.Context, publicKeyBytes []byte) (id int64, err error) {
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	_, err = s.r.GetUserByFingerprint(ctx, fingerprint)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("failed to get user by fingerprint: %w", err)
		}
	} else {
		return 0, fmt.Errorf("%w: fingerprint already exists", ErrConflict)
	}
	id, err = s.r.CreateUser(ctx, query.CreateUserParams{
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

func (s *UserService) Find(ctx context.Context, id int64) (*User, error) {
	qUser, err := s.r.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user #%d: %w", id, err)
	}
	user, err := newUserFromQueryUser(qUser)
	if err != nil {
		return nil, fmt.Errorf("failed to new user: %w", err)
	}
	return user, nil
}

func (s *UserService) FindByFingerprint(ctx context.Context, fingerprint string) (*User, error) {
	qUser, err := s.r.GetUserByFingerprint(ctx, fingerprint)
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
