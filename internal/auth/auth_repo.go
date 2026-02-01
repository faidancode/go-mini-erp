package auth

import (
	"context"

	db "go-mini-erp/internal/shared/database/sqlc"

	"github.com/google/uuid"
)

//go:generate mockgen -source=auth_repo.go -destination=mocks/auth_repository_mock.go -package=mocks

// Repository defines auth data access contract
type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (db.GetUserByUsernameRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)

	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error

	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]db.GetUserRolesRow, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]db.GetUserMenusRow, error)

	AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) (db.AssignRoleToUserRow, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error

	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

// repository is concrete implementation
// depends on sqlc interface, NOT concrete Queries
type repository struct {
	q db.Querier
}

// NewRepository creates auth repository
func NewRepository(q db.Querier) Repository {
	return &repository{
		q: q,
	}
}

// ==========================
// User
// ==========================

func (r *repository) GetUserByUsername(
	ctx context.Context,
	username string,
) (db.GetUserByUsernameRow, error) {
	return r.q.GetUserByUsername(ctx, username)
}

func (r *repository) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (db.GetUserByIDRow, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *repository) GetUserByEmail(
	ctx context.Context,
	email string,
) (db.GetUserByEmailRow, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *repository) CreateUser(
	ctx context.Context,
	arg db.CreateUserParams,
) (db.CreateUserRow, error) {
	return r.q.CreateUser(ctx, arg)
}

func (r *repository) UpdateUserLastLogin(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.q.UpdateUserLastLogin(ctx, id)
}

// ==========================
// Role & Menu
// ==========================

func (r *repository) GetUserRoles(
	ctx context.Context,
	userID uuid.UUID,
) ([]db.GetUserRolesRow, error) {
	return r.q.GetUserRoles(ctx, userID)
}

func (r *repository) GetUserMenus(
	ctx context.Context,
	userID uuid.UUID,
) ([]db.GetUserMenusRow, error) {
	return r.q.GetUserMenus(ctx, userID)
}

func (r *repository) AssignRoleToUser(
	ctx context.Context,
	arg db.AssignRoleToUserParams,
) (db.AssignRoleToUserRow, error) {
	return r.q.AssignRoleToUser(ctx, arg)
}

func (r *repository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	arg := db.RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return r.q.RemoveRoleFromUser(ctx, arg)
}

// ==========================
// Validation helpers
// ==========================

func (r *repository) CheckUsernameExists(
	ctx context.Context,
	username string,
) (bool, error) {
	return r.q.CheckUsernameExists(ctx, username)
}

func (r *repository) CheckEmailExists(
	ctx context.Context,
	email string,
) (bool, error) {
	return r.q.CheckEmailExists(ctx, email)
}
