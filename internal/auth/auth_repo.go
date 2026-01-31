package auth

import (
	"context"
	"go-mini-erp/internal/dbgen"

	"github.com/google/uuid"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (dbgen.GetUserByUsernameRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (dbgen.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (dbgen.GetUserByEmailRow, error)
	CreateUser(ctx context.Context, arg dbgen.CreateUserParams) (dbgen.CreateUserRow, error)
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]dbgen.GetUserRolesRow, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]dbgen.GetUserMenusRow, error)
	AssignRoleToUser(ctx context.Context, arg dbgen.AssignRoleToUserParams) (dbgen.AssignRoleToUserRow, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

type repository struct {
	queries *dbgen.Queries
}

func NewRepository(queries *dbgen.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (dbgen.GetUserByUsernameRow, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (dbgen.GetUserByIDRow, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (dbgen.GetUserByEmailRow, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *repository) CreateUser(ctx context.Context, arg dbgen.CreateUserParams) (dbgen.CreateUserRow, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *repository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateUserLastLogin(ctx, id)
}

func (r *repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]dbgen.GetUserRolesRow, error) {
	return r.queries.GetUserRoles(ctx, userID)
}

func (r *repository) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]dbgen.GetUserMenusRow, error) {
	return r.queries.GetUserMenus(ctx, userID)
}

func (r *repository) AssignRoleToUser(ctx context.Context, arg dbgen.AssignRoleToUserParams) (dbgen.AssignRoleToUserRow, error) {
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
