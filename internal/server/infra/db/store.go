package db

import (
	"database/sql"
	"gophkeeper/internal/server/infra/db/query"
)

type Store struct {
	*query.Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: query.New(db),
	}
}
