package scylla

import (
	"context"
	"errors"
	"testing"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/db"
)

func TestNewBatchNoConnection(t *testing.T) {
	t.Parallel()

	c := New()
	ctx := context.Background()

	b := c.NewBatch(ctx, gocql.LoggedBatch)
	if b == nil {
		t.Fatalf("expected batch")
	}

	if b.ctx != ctx {
		t.Fatalf("expected context to be preserved")
	}

	if b.batch != nil {
		t.Fatalf("expected nil underlying batch without connection")
	}

	if b.session != nil {
		t.Fatalf("expected nil session without connection")
	}

	if b.conn != c {
		t.Fatalf("expected original connection to be stored")
	}
}

func TestBatchAddNilReceiver(t *testing.T) {
	t.Parallel()

	var b *Batch

	err := b.Add(db.Query{SQL: "insert into users (id) values (?)", Args: []any{1}})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestBatchAddNilBatch(t *testing.T) {
	t.Parallel()

	b := &Batch{}

	err := b.Add(db.Query{SQL: "insert into users (id) values (?)", Args: []any{1}})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestBatchAddEmptyQuery(t *testing.T) {
	t.Parallel()

	tests := []string{"", "   ", "\n\t"}

	for _, query := range tests {
		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			b := &Batch{
				batch:   gocql.NewBatch(gocql.LoggedBatch),
				session: &gocql.Session{},
			}

			err := b.Add(db.Query{SQL: query})
			if !errors.Is(err, db.ErrQueryIsEmpty) {
				t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
			}

			if got := len(b.batch.Entries); got != 0 {
				t.Fatalf("expected no batch entries, got %d", got)
			}
		})
	}
}

func TestBatchAddAddsQuery(t *testing.T) {
	t.Parallel()

	b := &Batch{
		batch:   gocql.NewBatch(gocql.UnloggedBatch),
		session: &gocql.Session{},
	}

	err := b.Add(db.Query{
		SQL:  "insert into users (id, name) values (?, ?)",
		Args: []any{int64(1), "Martin"},
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if got := len(b.batch.Entries); got != 1 {
		t.Fatalf("expected 1 batch entry, got %d", got)
	}

	entry := b.batch.Entries[0]
	if entry.Stmt != "insert into users (id, name) values (?, ?)" {
		t.Fatalf("unexpected statement: %q", entry.Stmt)
	}

	if len(entry.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(entry.Args))
	}

	if entry.Args[0] != int64(1) || entry.Args[1] != "Martin" {
		t.Fatalf("unexpected args: %#v", entry.Args)
	}
}

func TestBatchAddSQLNilReceiver(t *testing.T) {
	t.Parallel()

	var b *Batch

	err := b.AddSQL("insert into users (id) values (?)", 1)
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestBatchAddSQLNilConn(t *testing.T) {
	t.Parallel()

	b := &Batch{}

	err := b.AddSQL("insert into users (id) values (?)", 1)
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestBatchAddSQLEmptyQuery(t *testing.T) {
	t.Parallel()

	b := &Batch{
		batch:   gocql.NewBatch(gocql.LoggedBatch),
		session: &gocql.Session{},
		conn:    New(),
	}

	err := b.AddSQL("   ")
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}
}

func TestBatchAddSQLAddsQuery(t *testing.T) {
	t.Parallel()

	b := &Batch{
		batch:   gocql.NewBatch(gocql.LoggedBatch),
		session: &gocql.Session{},
		conn:    New(),
	}

	err := b.AddSQL("delete from users where id = ?", int64(10))
	if err != nil {
		t.Fatalf("AddSQL returned error: %v", err)
	}

	if got := len(b.batch.Entries); got != 1 {
		t.Fatalf("expected 1 batch entry, got %d", got)
	}

	entry := b.batch.Entries[0]
	if entry.Stmt != "delete from users where id = ?" {
		t.Fatalf("unexpected statement: %q", entry.Stmt)
	}

	if len(entry.Args) != 1 || entry.Args[0] != int64(10) {
		t.Fatalf("unexpected args: %#v", entry.Args)
	}
}

func TestBatchExecuteNilReceiver(t *testing.T) {
	t.Parallel()

	var b *Batch

	err := b.Execute()
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestBatchExecuteNilBatch(t *testing.T) {
	t.Parallel()

	b := &Batch{}

	err := b.Execute()
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestNewBatchHelpersNoConnection(t *testing.T) {
	t.Parallel()

	c := New()
	ctx := context.Background()

	tests := []struct {
		name string
		b    *Batch
	}{
		{"logged", c.NewLoggedBatch(ctx)},
		{"unlogged", c.NewUnloggedBatch(ctx)},
		{"counter", c.NewCounterBatch(ctx)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.b == nil {
				t.Fatalf("expected batch")
			}

			if tt.b.ctx != ctx {
				t.Fatalf("expected context to be preserved")
			}

			if tt.b.batch != nil || tt.b.session != nil {
				t.Fatalf("expected no underlying batch/session without connection")
			}
		})
	}
}
