package scylla

import (
	"strings"

	"github.com/netlifeguru/db"
)

func (c *Connect) AnalyzeSQL(q string) (prepared string, placeholders int, err error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "", 0, db.ErrQueryIsEmpty
	}

	var (
		inSingle       bool
		inDouble       bool
		inLineComment  bool
		inBlockComment bool
	)

	i := 0
	for i < len(q) {
		ch := q[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}

			i++
			continue
		}

		if inBlockComment {
			if ch == '*' && i+1 < len(q) && q[i+1] == '/' {
				i += 2
				inBlockComment = false
				continue
			}

			i++
			continue
		}

		if !inSingle && !inDouble {
			if ch == '-' && i+1 < len(q) && q[i+1] == '-' {
				if i+2 == len(q) || isCQLSpace(q[i+2]) {
					inLineComment = true
					i += 2
					continue
				}
			}

			if ch == '/' && i+1 < len(q) && q[i+1] == '*' {
				inBlockComment = true
				i += 2
				continue
			}
		}

		if ch == '\'' && !inDouble {
			if inSingle {
				if i+1 < len(q) && q[i+1] == '\'' {
					i += 2
					continue
				}

				inSingle = false
			} else {
				inSingle = true
			}

			i++
			continue
		}

		if ch == '"' && !inSingle {
			if inDouble {
				if i+1 < len(q) && q[i+1] == '"' {
					i += 2
					continue
				}

				inDouble = false
			} else {
				inDouble = true
			}

			i++
			continue
		}

		if ch == '?' && !inSingle && !inDouble {
			placeholders++
		}

		i++
	}

	switch {
	case inSingle:
		return "", 0, db.ErrPrepareUnterminatedSingleQuotedString
	case inDouble:
		return "", 0, db.ErrPrepareUnterminatedDoubleQuotedIdent
	case inBlockComment:
		return "", 0, db.ErrPrepareUnterminatedBlockComment
	}

	return q, placeholders, nil
}

func isCQLSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

func (c *Connect) SelectSQL(q db.DialectSQL) string {
	return q.Scylla
}
