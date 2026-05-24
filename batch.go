package scylla

import (
	"context"
	"strings"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/db"
)

type BatchConn interface {
	NewBatch(ctx context.Context, batchType gocql.BatchType) *Batch
	NewLoggedBatch(ctx context.Context) *Batch
	NewUnloggedBatch(ctx context.Context) *Batch
	NewCounterBatch(ctx context.Context) *Batch
}

type Batch struct {
	batch       *gocql.Batch
	session     *gocql.Session
	ctx         context.Context
	consistency gocql.Consistency
	conn        db.Conn
}

func (c *Connect) NewBatch(ctx context.Context, batchType gocql.BatchType) *Batch {
	p, err := c.currentPool()
	if err != nil {
		return &Batch{
			ctx:  ctx,
			conn: c,
		}
	}

	batch := p.Connection.NewBatch(batchType)
	batch.Cons = p.Consistency

	return &Batch{
		batch:       batch,
		session:     p.Connection,
		ctx:         ctx,
		consistency: p.Consistency,
		conn:        c,
	}
}

func (b *Batch) Add(q db.Query) error {
	if b == nil || b.batch == nil || b.session == nil {
		return db.ErrNoConnection
	}

	if strings.TrimSpace(q.SQL) == "" {
		return db.ErrQueryIsEmpty
	}

	b.batch.Query(q.SQL, q.Args...)
	return nil
}

func (b *Batch) AddSQL(query string, args ...any) error {
	if b == nil || b.conn == nil {
		return db.ErrNoConnection
	}

	q, err := db.Raw(query, args...)
	if err != nil {
		return err
	}

	return b.Add(q)
}

func (b *Batch) Execute() error {
	if b == nil || b.batch == nil || b.session == nil {
		return db.ErrNoConnection
	}

	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	b.batch.Cons = b.consistency

	return b.session.ExecuteBatch(b.batch.WithContext(ctx))
}

func (c *Connect) NewLoggedBatch(ctx context.Context) *Batch {
	return c.NewBatch(ctx, gocql.LoggedBatch)
}

func (c *Connect) NewUnloggedBatch(ctx context.Context) *Batch {
	return c.NewBatch(ctx, gocql.UnloggedBatch)
}

func (c *Connect) NewCounterBatch(ctx context.Context) *Batch {
	return c.NewBatch(ctx, gocql.CounterBatch)
}
