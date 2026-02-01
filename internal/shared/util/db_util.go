package util

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToBool mengonversi pgtype.Bool ke bool standar
func ToBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

// FromBool mengonversi bool ke pgtype.Bool
func FromBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// ToUUIDPtr mengonversi pgtype.UUID ke *uuid.UUID (Google)
func ToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

// ToUUID mengonversi pgtype.UUID ke uuid.UUID (Google)
func ToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return uuid.UUID(u.Bytes)
}

// ToTextPtr mengonversi pgtype.Text ke *string
func ToTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// ToTimePtr mengonversi pgtype.Timestamptz ke *time.Time
func ToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
