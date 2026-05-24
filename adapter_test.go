package scylla

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/netlifeguru/mapper"
)

func TestRowsAdapterImplementsMapperRows(t *testing.T) {
	t.Parallel()

	var _ mapper.Rows = (*rowsAdapter)(nil)
}

func TestRowsAdapterColumnsReturnsCopy(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{cols: []string{"id", "name"}}

	first, err := rows.Columns()
	if err != nil {
		t.Fatalf("Columns returned error: %v", err)
	}

	first[0] = "changed"

	second, err := rows.Columns()
	if err != nil {
		t.Fatalf("Columns returned error: %v", err)
	}

	want := []string{"id", "name"}
	if !reflect.DeepEqual(second, want) {
		t.Fatalf("expected columns copy %#v, got %#v", want, second)
	}
}

func TestRowsAdapterNextWithNilIterReturnsFalse(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{}

	if rows.Next() {
		t.Fatalf("expected Next to return false with nil iter")
	}
}

func TestRowsAdapterNextWhenClosedReturnsFalse(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{closed: true}

	if rows.Next() {
		t.Fatalf("expected Next to return false when closed")
	}
}

func TestRowsAdapterScanRequiresNext(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols:    []string{"id"},
		current: map[string]any{"id": int64(1)},
		ready:   false,
	}

	var id int64
	err := rows.Scan(&id)
	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), "Scan called without Next") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRowsAdapterScanDestinationCountMismatch(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols: []string{"id", "name"},
		current: map[string]any{
			"id":   int64(1),
			"name": "Martin",
		},
		ready: true,
	}

	var id int64
	err := rows.Scan(&id)
	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), "expected 2 destinations, got 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRowsAdapterScanIntoAny(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols: []string{"id", "name"},
		current: map[string]any{
			"id":   int64(1),
			"name": "Martin",
		},
		ready: true,
	}

	var id any
	var name any

	err := rows.Scan(&id, &name)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if id != int64(1) {
		t.Fatalf("expected id int64(1), got %#v", id)
	}

	if name != "Martin" {
		t.Fatalf("expected name Martin, got %#v", name)
	}

	if rows.ready {
		t.Fatalf("expected ready=false after successful scan")
	}
}

func TestRowsAdapterScanTypedDestinations(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols: []string{"id", "name", "active"},
		current: map[string]any{
			"id":     int64(1),
			"name":   []byte("Martin"),
			"active": true,
		},
		ready: true,
	}

	var id int64
	var name string
	var active bool

	err := rows.Scan(&id, &name, &active)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if id != 1 {
		t.Fatalf("expected id 1, got %d", id)
	}

	if name != "Martin" {
		t.Fatalf("expected name Martin, got %q", name)
	}

	if !active {
		t.Fatalf("expected active true")
	}

	if rows.ready {
		t.Fatalf("expected ready=false after successful scan")
	}
}

func TestRowsAdapterScanRejectsNonPointerDestination(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols:    []string{"id"},
		current: map[string]any{"id": int64(1)},
		ready:   true,
	}

	var id int64
	err := rows.Scan(id)
	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), `destination 0 for column "id" must be non-nil pointer`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRowsAdapterScanRejectsNilPointerDestination(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols:    []string{"id"},
		current: map[string]any{"id": int64(1)},
		ready:   true,
	}

	var id *int64
	err := rows.Scan(id)
	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), `destination 0 for column "id" must be non-nil pointer`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRowsAdapterScanAssignError(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{
		cols:    []string{"active"},
		current: map[string]any{"active": "not-bool"},
		ready:   true,
	}

	var active bool
	err := rows.Scan(&active)
	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), `assign column "active"`) {
		t.Fatalf("expected column name in error, got %v", err)
	}
}

func TestRowsAdapterErr(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("iter failed")
	rows := &rowsAdapter{err: wantErr}

	if !errors.Is(rows.Err(), wantErr) {
		t.Fatalf("expected err %v, got %v", wantErr, rows.Err())
	}
}

func TestRowsAdapterCloseWithNilIter(t *testing.T) {
	t.Parallel()

	rows := &rowsAdapter{}

	if err := rows.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if !rows.closed {
		t.Fatalf("expected rows to be marked closed")
	}
}

func TestRowsAdapterCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("already failed")
	rows := &rowsAdapter{
		err:    wantErr,
		closed: true,
	}

	if err := rows.Close(); !errors.Is(err, wantErr) {
		t.Fatalf("expected existing error, got %v", err)
	}
}
