package dbutil

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PgUUIDToUUIDPtr(pg pgtype.UUID) *uuid.UUID {
	if !pg.Valid {
		return nil
	}

	// Konversi [16]byte ke uuid.UUID (Google)
	id := uuid.UUID(pg.Bytes)
	return &id
}

// Dari *uuid.UUID ke pgtype.UUID
func UUIDPtrToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}

	return pgtype.UUID{
		Bytes: [16]byte(*id),
		Valid: true,
	}
}

// Bool helpers

func BoolPtr(v bool) *bool {
	return &v
}

func PgBool(v bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  v,
		Valid: true,
	}
}

// Time helpers
func NullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: true,
	}
}

func PgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
