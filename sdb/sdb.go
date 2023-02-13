// Package sdb contains some utility functions for golang database/sql
package sdb

import (
	"context"
	"database/sql"
	"errors"
)

type dbContextKeyType struct{}

var dbContextKey dbContextKeyType

var ErrNotInitialized = errors.New("database has not been initialized")

func BindContext(ctx context.Context, db *sql.DB) context.Context {
	if db == nil {
		return ctx
	}
	newCtx := context.WithValue(ctx, dbContextKey, db)
	return newCtx
}

func FromContext(ctx context.Context) (*sql.DB, error) {
	val := ctx.Value(dbContextKey)
	if val == nil {
		return nil, ErrNotInitialized
	}
	if v, ok := val.(*sql.DB); ok {
		return v, nil
	}
	return nil, ErrNotInitialized
}

// QueryContext wrap sql.DB's QueryContext method
func QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

// ExecContext wrap sql.DB's ExecContext method
func ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}
