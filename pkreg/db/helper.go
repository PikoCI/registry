package db

import (
	"database/sql"
	"strings"
	"time"
)

const timeFormat = time.RFC3339Nano

func isEntityFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func toNullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func toNullInt64(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

// timeToString converts a time.Time to a string for DB storage.
// Stored as RFC3339Nano which works across SQLite (TEXT), MySQL, and PostgreSQL.
func timeToString(t time.Time) sql.NullString {
	if t.IsZero() {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: t.Format(timeFormat), Valid: true}
}

// stringToTime parses a time string from the DB back to time.Time.
func stringToTime(s sql.NullString) time.Time {
	if !s.Valid || s.String == "" {
		return time.Time{}
	}
	t, err := time.Parse(timeFormat, s.String)
	if err != nil {
		// Try other common formats
		t, err = time.Parse(time.RFC3339, s.String)
		if err != nil {
			t, _ = time.Parse("2006-01-02 15:04:05", s.String)
		}
	}
	return t
}

// dateToString converts a time to a date string for the download_log table.
func dateToString(t time.Time) string {
	return t.Format("2006-01-02")
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "duplicate key value")
}
