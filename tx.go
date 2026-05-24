package scylla

import (
	"context"
	"errors"

	"github.com/netlifeguru/db"
)

var ErrTxUnsupported = errors.New("scylla: transactions are not supported")

func (c *Connect) Transaction(fn func(tx db.Conn) error) (retErr error) {
	return ErrTxUnsupported
}

func (c *Connect) TransactionCtx(ctx context.Context, fn func(tx db.Conn) error) error {
	return ErrTxUnsupported
}
