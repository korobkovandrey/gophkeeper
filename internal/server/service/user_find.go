package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
)

//go:generate mockgen -source=user_find.go -destination=mocks/user_find.go -package=mocks

// userFinderByID defines the interface for finding a user by ID.
type userFinderByID interface {
	GetUserById(ctx context.Context, id int64) (query.User, error)
}

type FindUserByIDFunc func(ctx context.Context, id int64) (*User, error)

// NewFindUserByIDFunc returns a function that handles finding a user by ID.
func NewFindUserByIDFunc(finderByID userFinderByID) FindUserByIDFunc {
	return func(ctx context.Context, id int64) (*User, error) {
		qUser, err := finderByID.GetUserById(ctx, id)
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
}
