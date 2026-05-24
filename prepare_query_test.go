package scylla

import (
	"errors"
	"testing"

	"github.com/netlifeguru/db"
)

func TestAnalyzeSQLEmpty(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL("")
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}

	if prepared != "" {
		t.Fatalf("expected empty prepared query, got %q", prepared)
	}

	if placeholders != 0 {
		t.Fatalf("expected 0 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLWhitespaceOnly(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL("   \n\t  ")
	if !errors.Is(err, db.ErrQueryIsEmpty) {
		t.Fatalf("expected ErrQueryIsEmpty, got %v", err)
	}
}

func TestAnalyzeSQLCountsPlaceholders(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL(
		"select * from users where id = ? and email = ? and active = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	want := "select * from users where id = ? and email = ? and active = ?"
	if prepared != want {
		t.Fatalf("expected %q, got %q", want, prepared)
	}

	if placeholders != 3 {
		t.Fatalf("expected 3 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLTrimsQuery(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL("  select * from users where id = ?  ")
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if prepared != "select * from users where id = ?" {
		t.Fatalf("expected trimmed query, got %q", prepared)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLNoPlaceholders(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL("select * from users")
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if prepared != "select * from users" {
		t.Fatalf("unexpected prepared query: %q", prepared)
	}

	if placeholders != 0 {
		t.Fatalf("expected 0 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInSingleQuotedString(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users where text = 'hello ? world' and id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInDoubleQuotedIdentifier(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select "weird ? column" from users where id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInLineComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users -- ignored ?\nwhere id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLDashDashWithoutSpaceIsNotComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users where value--? = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 2 {
		t.Fatalf("expected 2 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInBlockComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users /* ignored ? */ where id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLMultipleContexts(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	query := `
select '?', "weird ? ident"
from users
where id = ?
  and email = ?
-- ignored ?
/* ignored ? */
and status = ?
`

	_, placeholders, err := c.AnalyzeSQL(query)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 3 {
		t.Fatalf("expected 3 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLDoubledSingleQuote(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select * from users where text = 'john''s ? text' and id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLDoubledDoubleQuote(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select "hello "" ? world" from users where id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLLineCommentAtEOF(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users where id = ? -- ignored ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLUnterminatedSingleQuotedString(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL("select * from users where name = 'john")
	if !errors.Is(err, db.ErrPrepareUnterminatedSingleQuotedString) {
		t.Fatalf("expected ErrPrepareUnterminatedSingleQuotedString, got %v", err)
	}

	if prepared != "" {
		t.Fatalf("expected empty prepared query on error, got %q", prepared)
	}

	if placeholders != 0 {
		t.Fatalf("expected 0 placeholders on error, got %d", placeholders)
	}
}

func TestAnalyzeSQLUnterminatedDoubleQuotedIdentifier(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL(`select "name from users`)
	if !errors.Is(err, db.ErrPrepareUnterminatedDoubleQuotedIdent) {
		t.Fatalf("expected ErrPrepareUnterminatedDoubleQuotedIdent, got %v", err)
	}
}

func TestAnalyzeSQLUnterminatedBlockComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL("select * from users /* comment ?")
	if !errors.Is(err, db.ErrPrepareUnterminatedBlockComment) {
		t.Fatalf("expected ErrPrepareUnterminatedBlockComment, got %v", err)
	}
}

func TestIsCQLSpace(t *testing.T) {
	t.Parallel()

	for _, ch := range []byte{' ', '\t', '\n', '\r', '\f'} {
		if !isCQLSpace(ch) {
			t.Fatalf("expected %q to be CQL space", ch)
		}
	}

	if isCQLSpace('x') {
		t.Fatalf("did not expect x to be CQL space")
	}
}

func TestSelectSQL(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	got := c.SelectSQL(db.DialectSQL{
		Postgres: "select $1",
		Mysql:    "select ?",
		Scylla:   "select scylla",
	})

	if got != "select scylla" {
		t.Fatalf("expected scylla SQL, got %q", got)
	}
}
