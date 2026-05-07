package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgTimestamptz_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	got := pgTimestamptz(now)
	require.True(t, got.Valid)
	assert.Equal(t, now, got.Time)
	assert.Equal(t, now, goTime(got))
}

func TestGoTime_InvalidReturnsZero(t *testing.T) {
	zero := goTime(pgtype.Timestamptz{})
	assert.True(t, zero.IsZero())
}

func TestGoTimePtr_NilWhenInvalid(t *testing.T) {
	assert.Nil(t, goTimePtr(pgtype.Timestamptz{}))

	now := time.Now().UTC()
	ptr := goTimePtr(pgtype.Timestamptz{Time: now, Valid: true})
	require.NotNil(t, ptr)
	assert.Equal(t, now, *ptr)
}

func TestPgDate_NilReturnsInvalid(t *testing.T) {
	d := pgDate(nil)
	assert.False(t, d.Valid)
}

func TestPgDate_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	d := pgDate(&now)
	require.True(t, d.Valid)
	assert.Equal(t, now, goDateValue(d))
}

func TestGoDateValue_InvalidReturnsZero(t *testing.T) {
	assert.True(t, goDateValue(pgtype.Date{}).IsZero())
}

func TestGoDatePtr_NilWhenInvalid(t *testing.T) {
	assert.Nil(t, goDatePtr(pgtype.Date{}))
	now := time.Now().UTC()
	ptr := goDatePtr(pgtype.Date{Time: now, Valid: true})
	require.NotNil(t, ptr)
	assert.Equal(t, now, *ptr)
}

func TestPgInt8_AlwaysValid(t *testing.T) {
	v := pgInt8(42)
	assert.True(t, v.Valid)
	assert.Equal(t, int64(42), v.Int64)
}

func TestPgOptionalInt8_NilReturnsInvalid(t *testing.T) {
	v := pgOptionalInt8(nil)
	assert.False(t, v.Valid)

	x := int64(99)
	v = pgOptionalInt8(&x)
	assert.True(t, v.Valid)
	assert.Equal(t, int64(99), v.Int64)
}

func TestPgOptionalTimestamptz_NilReturnsInvalid(t *testing.T) {
	v := pgOptionalTimestamptz(nil)
	assert.False(t, v.Valid)

	now := time.Now().UTC()
	v = pgOptionalTimestamptz(&now)
	assert.True(t, v.Valid)
	assert.Equal(t, now, v.Time)
}

func TestGoInt64_InvalidReturnsZero(t *testing.T) {
	assert.Zero(t, goInt64(pgtype.Int8{}))
}

func TestGoInt64_ValidReturnsValue(t *testing.T) {
	assert.Equal(t, int64(7), goInt64(pgtype.Int8{Int64: 7, Valid: true}))
}

func TestGoInt64Ptr(t *testing.T) {
	assert.Nil(t, goInt64Ptr(pgtype.Int8{}))
	ptr := goInt64Ptr(pgtype.Int8{Int64: 5, Valid: true})
	require.NotNil(t, ptr)
	assert.Equal(t, int64(5), *ptr)
}

func TestPgNumeric_RoundTrip(t *testing.T) {
	n := pgNumeric(1.2345)
	got := goFloat64(n)
	assert.InDelta(t, 1.2345, got, 0.0001)
}

func TestGoFloat64_InvalidReturnsOne(t *testing.T) {
	assert.Equal(t, 1.0, goFloat64(pgtype.Numeric{}))
}

func TestIsDuplicateError_PgErr23505(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate key"}
	assert.True(t, isDuplicateError(pgErr))
}

func TestIsDuplicateError_OtherCode(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23502", Message: "not null"}
	assert.False(t, isDuplicateError(pgErr))
}

func TestIsDuplicateError_NonPgErr(t *testing.T) {
	assert.False(t, isDuplicateError(errors.New("plain error")))
}
