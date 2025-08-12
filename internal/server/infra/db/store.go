package db

import (
	"context"
	"database/sql"
	"fmt"
	"gophkeeper/internal/server/infra/db/query"
)

type Store struct {
	*query.Queries
	db *sql.DB
}

func newStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: query.New(db),
	}
}

func (s *Store) Close() error {
	return s.db.Close()
}

func MakeStoreConnectAndMigrate(ctx context.Context, dsn string) (*Store, error) {
	if err := runMigrate(dsn); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	dbConnect, err := connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return newStore(dbConnect), nil
}
