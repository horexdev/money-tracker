package repository

import (
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

// pgOptionalInt8 converts a *int64 to a pgtype.Int8 (invalid/null when nil).
func pgOptionalInt8(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

// pgOptionalTimestamptz converts a *time.Time to a pgtype.Timestamptz (invalid/null when nil).
func pgOptionalTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// goInt64Ptr converts a pgtype.Int8 to a *int64 (nil if invalid).
func goInt64Ptr(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	x := v.Int64
	return &x
}

// goInt64 converts a pgtype.Int8 to an int64 (0 if invalid).
func goInt64(v pgtype.Int8) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}

// pgNumeric converts a float64 to a pgtype.Numeric.
func pgNumeric(v float64) pgtype.Numeric {
	s := strconv.FormatFloat(v, 'f', 8, 64)
	var num pgtype.Numeric
	_ = num.Scan(s)
	return num
}

// isDuplicateError returns true when err is a PostgreSQL unique-constraint violation (23505).
func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// goFloat64 converts a pgtype.Numeric to a float64 (1.0 if invalid).
func goFloat64(v pgtype.Numeric) float64 {
	if !v.Valid {
		return 1.0
	}
	f, _ := v.Float64Value()
	if !f.Valid {
		return 1.0
	}
	return f.Float64
}
