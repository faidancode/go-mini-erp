package dbutil

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

//
// =======================
// UUID
// =======================
//

// pgtype.UUID -> *uuid.UUID
func PgUUIDToUUIDPtr(pg pgtype.UUID) *uuid.UUID {
	if !pg.Valid {
		return nil
	}
	id := uuid.UUID(pg.Bytes)
	return &id
}

// *uuid.UUID -> pgtype.UUID
func UUIDPtrToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: [16]byte(*id),
		Valid: true,
	}
}

//
// =======================
// STRING
// =======================
//

// string -> string (identity, untuk konsistensi)
func StringValue(s string) string {
	return s
}

// *string -> string (default "")
func StringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// *string -> sql.NullString
func StringPtrToNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

// string -> sql.NullString
func StringToNull(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// *string -> pgtype.Text
func StringPtrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

//
// =======================
// BOOL
// =======================
//

// bool -> *bool
func BoolPtr(v bool) *bool {
	return &v
}

// *bool -> bool (default false)
func BoolPtrValue(b *bool, defaultValue bool) bool {
	if b == nil {
		return defaultValue
	}
	return *b
}

// *bool -> sql.NullBool
func BoolPtrToNull(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{
		Bool:  *b,
		Valid: true,
	}
}

// bool -> pgtype.Bool
func BoolToPgBool(v bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  v,
		Valid: true,
	}
}

//
// =======================
// TIME
// =======================
//

// time.Time -> sql.NullTime
func TimeToNull(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: true,
	}
}

// *time.Time -> sql.NullTime
func TimePtrToNull(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}

// time.Time -> pgtype.Timestamptz
func TimeToPgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

// *time.Time -> pgtype.Timestamptz
func TimePtrToPgTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{
		Time:  *t,
		Valid: true,
	}
}

func PgTimeValue(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

//
// =======================
// INT32
// =======================
//

// *int32 -> int32
func Int32PtrValue(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// *int32 -> sql.NullInt32
func Int32PtrToNull(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{
		Int32: *i,
		Valid: true,
	}
}

//
// =======================
// FLOAT64
// =======================
//

// *float64 -> float64
func Float64PtrValue(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

//
// =======================
// DECIMAL
// =======================
//

// *decimal.Decimal -> decimal.Decimal
func DecimalPtrValue(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}

// float64 -> decimal (aman, tapi bukan exact)
func Float64ToDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// *float64 -> decimal
func Float64PtrToDecimal(f *float64) decimal.Decimal {
	if f == nil {
		return decimal.Zero
	}
	return decimal.NewFromFloat(*f)
}

// float64 -> decimal (EXACT, cocok uang)
func Float64ToDecimalExact(f float64) decimal.Decimal {
	return decimal.RequireFromString(
		strconv.FormatFloat(f, 'f', -1, 64),
	)
}

// *float64 -> decimal exact
func Float64PtrToDecimalExact(f *float64) decimal.Decimal {
	if f == nil {
		return decimal.Zero
	}
	return Float64ToDecimalExact(*f)
}

// decimal -> float64
func DecimalToFloat64(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

//
// =======================
// DECIMAL -> NULL
// =======================
//

// *float64 -> decimal.NullDecimal
func Float64PtrToNullDecimal(f *float64) decimal.NullDecimal {
	if f == nil {
		return decimal.NullDecimal{}
	}
	return decimal.NullDecimal{
		Decimal: decimal.NewFromFloat(*f),
		Valid:   true,
	}
}
