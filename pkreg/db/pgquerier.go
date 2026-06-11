package db

import (
	"context"
	"database/sql"
	"strings"
)

type PGQuerier struct {
	db *sql.DB
}

func NewPGQuerier(db *sql.DB) *PGQuerier {
	return &PGQuerier{db: db}
}

func rewritePlaceholders(query string) string {
	var b strings.Builder
	n := 1
	inSingleQuote := false
	for i := 0; i < len(query); i++ {
		c := query[i]
		if c == '\'' {
			inSingleQuote = !inSingleQuote
			b.WriteByte(c)
		} else if c == '?' && !inSingleQuote {
			b.WriteByte('$')
			b.WriteString(itoa(n))
			n++
		} else if c == '`' && !inSingleQuote {
			b.WriteByte('"')
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

func (p *PGQuerier) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, rewritePlaceholders(query), args...)
}

func (p *PGQuerier) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, rewritePlaceholders(query), args...)
}

func (p *PGQuerier) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, rewritePlaceholders(query), args...)
}
