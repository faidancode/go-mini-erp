package auth

import (
	"context"
	db "go-mini-erp/internal/shared/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:generate mockgen -source=auth_repo.go -destination=mocks/auth_repository_mock.go -package=mocks
type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (db.GetUserByUsernameRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]db.GetUserRolesRow, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]db.GetUserMenusRow, error)
	AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) (db.AssignRoleToUserRow, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (db.GetUserByUsernameRow, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (db.GetUserByIDRow, error) {
	return r.queries.GetUserByID(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *repository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *repository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateUserLastLogin(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
}

func (r *repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]db.GetUserRolesRow, error) {
	return r.queries.GetUserRoles(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
}

func (r *repository) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]db.GetUserMenusRow, error) {
	return r.queries.GetUserMenus(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
}

func (r *repository) AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) (db.AssignRoleToUserRow, error) {
	return r.queries.AssignRoleToUser(ctx, arg)
}

func (r *repository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	result, err := r.queries.CheckUsernameExists(ctx, username)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *repository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	result, err := r.queries.CheckEmailExists(ctx, email)
	if err != nil {
		return false, err
	}
	return result, nil
}
