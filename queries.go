package scylla

import (
	"context"
	"strings"

	"github.com/netlifeguru/db"
	"github.com/netlifeguru/mapper"
)

func (c *Connect) Exec(q db.Query) (db.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()
	return c.ExecCtx(ctx, q)
}

func (c *Connect) ExecCtx(ctx context.Context, q db.Query) (db.Result, error) {
	if strings.TrimSpace(q.SQL) == "" {
		return nil, db.ErrQueryIsEmpty
	}

	p, err := c.currentPool()
	if err != nil {
		return nil, err
	}

	err = p.Connection.Query(q.SQL, q.Args...).
		WithContext(ctx).
		Consistency(p.Consistency).
		Exec()
	if err != nil {
		return nil, err
	}

	return Result{
		lastInsertId: 0,
		rowsAffected: 0,
	}, nil
}

func (c *Connect) Query(query db.Query, each func(row map[string]any) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()

	return c.QueryCtx(ctx, query, each)
}

func (c *Connect) QueryRows(ctx context.Context, q db.Query) (mapper.Rows, error) {
	if strings.TrimSpace(q.SQL) == "" {
		return nil, db.ErrQueryIsEmpty
	}

	p, err := c.currentPool()
	if err != nil {
		return nil, err
	}

	iter := p.Connection.Query(q.SQL, q.Args...).
		WithContext(ctx).
		Consistency(p.Consistency).
		Iter()

	return adaptRows(iter), nil
}
