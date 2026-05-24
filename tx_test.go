package scylla

import (
	"context"
	"errors"
	"testing"

	"github.com/netlifeguru/db"
)

func TestTransactionUnsupported(t *testing.T) {
	t.Parallel()

	c := New()

	called := false
	err := c.Transaction(func(tx db.Conn) error {
		called = true
		return nil
	})

	if !errors.Is(err, ErrTxUnsupported) {
		t.Fatalf("expected ErrTxUnsupported, got %v", err)
	}

	if called {
		t.Fatalf("transaction callback should not be called")
	}
}

func TestTransactionCtxUnsupported(t *testing.T) {
	t.Parallel()

	c := New()

	called := false
	err := c.TransactionCtx(context.Background(), func(tx db.Conn) error {
		called = true
		return nil
	})

	if !errors.Is(err, ErrTxUnsupported) {
		t.Fatalf("expected ErrTxUnsupported, got %v", err)
	}

	if called {
		t.Fatalf("transaction callback should not be called")
	}
}

func TestTransactionCtxUnsupportedWithCancelledContext(t *testing.T) {
	t.Parallel()

	c := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := c.TransactionCtx(ctx, func(tx db.Conn) error {
		called = true
		return nil
	})

	if !errors.Is(err, ErrTxUnsupported) {
		t.Fatalf("expected ErrTxUnsupported even with cancelled context, got %v", err)
	}

	if called {
		t.Fatalf("transaction callback should not be called")
	}
}

func TestTransactionNilCallbackUnsupported(t *testing.T) {
	t.Parallel()

	c := New()

	err := c.Transaction(nil)
	if !errors.Is(err, ErrTxUnsupported) {
		t.Fatalf("expected ErrTxUnsupported, got %v", err)
	}
}

func TestTransactionCtxNilCallbackUnsupported(t *testing.T) {
	t.Parallel()

	c := New()

	err := c.TransactionCtx(context.Background(), nil)
	if !errors.Is(err, ErrTxUnsupported) {
		t.Fatalf("expected ErrTxUnsupported, got %v", err)
	}
}

func TestErrTxUnsupportedMessage(t *testing.T) {
	t.Parallel()

	if ErrTxUnsupported == nil {
		t.Fatalf("expected ErrTxUnsupported to be defined")
	}

	if got := ErrTxUnsupported.Error(); got != "scylla: transactions are not supported" {
		t.Fatalf("unexpected ErrTxUnsupported message: %q", got)
	}
}
