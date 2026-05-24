package scylla

import (
	"context"
	"errors"
	"testing"

	"github.com/netlifeguru/db"
)

func TestExecCtxEmptyQuery(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	res, err := c.ExecCtx(context.Background(), db.Query{SQL: "   "})
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}

	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecEmptyQuery(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	res, err := c.Exec(db.Query{SQL: "\n\t "})
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}

	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecCtxNoConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	res, err := c.ExecCtx(context.Background(), db.Query{
		SQL:  "insert into users (id) values (?)",
		Args: []any{int64(1)},
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecNoConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	res, err := c.Exec(db.Query{
		SQL:  "insert into users (id) values (?)",
		Args: []any{int64(1)},
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecCtxNilPoolConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 1, 1, false)

	res, err := c.ExecCtx(context.Background(), db.Query{
		SQL:  "insert into users (id) values (?)",
		Args: []any{int64(1)},
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestQueryRowsEmptyQuery(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	rows, err := c.QueryRows(context.Background(), db.Query{SQL: "   "})
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}

	if rows != nil {
		t.Fatalf("expected nil rows, got %#v", rows)
	}
}

func TestQueryRowsNoConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	rows, err := c.QueryRows(context.Background(), db.Query{
		SQL:  "select * from users where id = ?",
		Args: []any{int64(1)},
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if rows != nil {
		t.Fatalf("expected nil rows, got %#v", rows)
	}
}

func TestQueryRowsNilPoolConnection(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 1, 1, false)

	rows, err := c.QueryRows(context.Background(), db.Query{
		SQL:  "select * from users where id = ?",
		Args: []any{int64(1)},
	})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if rows != nil {
		t.Fatalf("expected nil rows, got %#v", rows)
	}
}

func TestQueryPropagatesQueryCtxErrorEmptyQuery(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	err := c.Query(db.Query{SQL: "   "}, func(row map[string]any) error {
		return nil
	})
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}
}

func TestQueryNilCallback(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	err := c.Query(db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}
