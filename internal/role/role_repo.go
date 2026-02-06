package role

import (
	"context"

	db "go-mini-erp/internal/shared/database/sqlc"

	"github.com/google/uuid"
)

//go:generate mockgen -source=repository.go -destination=mocks/role_repository_mock.go -package=mocks

type Repository interface {
	// Basic CRUD
	CreateRole(ctx context.Context, arg db.CreateRoleParams) (db.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (db.Role, error)
	GetRoleByCode(ctx context.Context, code string) (db.Role, error)
	ListRoles(ctx context.Context) ([]db.Role, error)
	UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	q db.Querier
}

func NewRepository(q db.Querier) Repository {
	return &repository{q: q}
}

func (r *repository) CreateRole(ctx context.Context, arg db.CreateRoleParams) (db.Role, error) {
	return r.q.CreateRole(ctx, arg)
}

func (r *repository) GetRoleByID(ctx context.Context, id uuid.UUID) (db.Role, error) {
	return r.q.GetRoleByID(ctx, id)
}

func (r *repository) GetRoleByCode(ctx context.Context, code string) (db.Role, error) {
	return r.q.GetRoleByCode(ctx, code)
}

func (r *repository) ListRoles(ctx context.Context) ([]db.Role, error) {
	return r.q.ListRoles(ctx)
}

func (r *repository) UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error) {
	return r.q.UpdateRole(ctx, arg)
}

func (r *repository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteRole(ctx, id)
}
