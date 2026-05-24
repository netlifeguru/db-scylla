package scylla

import "testing"

func TestResultRowsAffected(t *testing.T) {
	t.Parallel()

	res := Result{rowsAffected: 42}

	if got := res.RowsAffected(); got != 42 {
		t.Fatalf("expected 42 rows affected, got %d", got)
	}
}

func TestResultRowsAffectedZeroValue(t *testing.T) {
	t.Parallel()

	var res Result

	if got := res.RowsAffected(); got != 0 {
		t.Fatalf("expected zero value rows affected to be 0, got %d", got)
	}
}

func TestResultLastInsertId(t *testing.T) {
	t.Parallel()

	res := Result{lastInsertId: 123}

	if got := res.LastInsertId(); got != 123 {
		t.Fatalf("expected last insert id 123, got %d", got)
	}
}

func TestResultLastInsertIdZeroValue(t *testing.T) {
	t.Parallel()

	var res Result

	if got := res.LastInsertId(); got != 0 {
		t.Fatalf("expected zero value last insert id to be 0, got %d", got)
	}
}
