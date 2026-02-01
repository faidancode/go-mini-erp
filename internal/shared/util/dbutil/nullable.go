package dbutil

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

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
