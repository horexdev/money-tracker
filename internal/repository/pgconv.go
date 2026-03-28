package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// pgTimestamptz converts a Go time.Time to a pgtype.Timestamptz.
func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// goTime converts a pgtype.Timestamptz to a Go time.Time.
func goTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// goTimePtr converts a pgtype.Timestamptz to a *time.Time (nil if invalid).
func goTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tt := t.Time
	return &tt
}

// pgDate converts a *time.Time to a pgtype.Date.
func pgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

// goDatePtr converts a pgtype.Date to a *time.Time (nil if invalid).
func goDatePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}

// pgInt8 converts an int64 to a pgtype.Int8 (valid).
func pgInt8(v int64) pgtype.Int8 {
	return pgtype.Int8{Int64: v, Valid: true}
}

// goInt64 converts a pgtype.Int8 to an int64 (0 if invalid).
func goInt64(v pgtype.Int8) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}
