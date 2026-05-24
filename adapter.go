package scylla

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/mapper"
)

type rowsAdapter struct {
	iter    *gocql.Iter
	cols    []string
	current map[string]any
	err     error
	ready   bool
	closed  bool
}

func adaptRows(iter *gocql.Iter) mapper.Rows {
	colsMeta := iter.Columns()
	cols := make([]string, 0, len(colsMeta))
	for _, col := range colsMeta {
		cols = append(cols, col.Name)
	}

	return &rowsAdapter{
		iter: iter,
		cols: cols,
	}
}

func (r *rowsAdapter) Next() bool {
	if r.closed || r.iter == nil {
		return false
	}

	m := make(map[string]any, len(r.cols))
	if !r.iter.MapScan(m) {
		r.err = r.iter.Close()
		r.closed = true
		return false
	}

	r.current = m
	r.ready = true
	return true
}

func (r *rowsAdapter) Scan(dest ...any) error {
	if !r.ready {
		return errors.New("scylla: Scan called without Next")
	}
	if len(dest) != len(r.cols) {
		return fmt.Errorf("scylla: expected %d destinations, got %d", len(r.cols), len(dest))
	}

	for i, col := range r.cols {
		val := r.current[col]

		if ptr, ok := dest[i].(*any); ok {
			*ptr = val
			continue
		}

		rv := reflect.ValueOf(dest[i])
		if !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil() {
			return fmt.Errorf("scylla: destination %d for column %q must be non-nil pointer", i, col)
		}

		if err := mapper.AssignValue(rv.Elem(), val); err != nil {
			return fmt.Errorf("scylla: assign column %q: %w", col, err)
		}
	}

	r.ready = false
	return nil
}

func (r *rowsAdapter) Err() error {
	return r.err
}

func (r *rowsAdapter) Close() error {
	if r.closed {
		return r.err
	}

	r.closed = true

	if r.iter != nil && r.err == nil {
		r.err = r.iter.Close()
	}

	return r.err
}

func (r *rowsAdapter) Columns() ([]string, error) {
	out := make([]string, len(r.cols))
	copy(out, r.cols)

	return out, nil
}
