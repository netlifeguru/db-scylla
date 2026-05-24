package scylla

import (
	"context"

	"github.com/netlifeguru/db"
	"github.com/netlifeguru/mapper"
)

func (c *Connect) QueryCtx(ctx context.Context, query db.Query, each func(row map[string]any) error) error {
	if each == nil {
		return db.ErrNilEachCallback
	}

	rows, err := c.QueryRows(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	return mapper.ScanMapRows(rows, each)
}
