package scylla

import (
	"context"
	"errors"
	"testing"

	"github.com/netlifeguru/db"
)

func TestQueryCtxNilCallback(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}

func TestQueryCtxEmptyQuery(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	err := c.QueryCtx(context.Background(), db.Query{SQL: "   "}, func(row map[string]any) error {
		return nil
	})
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}
}

func TestQueryCtxNoConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	err := c.QueryCtx(context.Background(), db.Query{
		SQL:  "select * from users where id = ?",
		Args: []any{int64(1)},
	}, func(row map[string]any) error {
		return nil
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestQueryCtxNilPoolConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 1, 1, false)

	err := c.QueryCtx(context.Background(), db.Query{
		SQL:  "select * from users where id = ?",
		Args: []any{int64(1)},
	}, func(row map[string]any) error {
		return nil
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestQueryCtxNilCallbackTakesPrecedenceOverEmptyQuery(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	err := c.QueryCtx(context.Background(), db.Query{SQL: "   "}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}

func TestQueryCtxNilCallbackTakesPrecedenceOverNoConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}
