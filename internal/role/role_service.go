package role

import (
	"context"
	"errors"

	db "go-mini-erp/internal/shared/database/sqlc"
	"go-mini-erp/internal/shared/util/dbutil"

	"github.com/google/uuid"
)

type Service interface {
	CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*RoleResponse, error)
	ListRoles(ctx context.Context) ([]RoleResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req UpdateRoleRequest) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error) {
	// cek apakah code sudah ada
	_, err := s.repo.GetRoleByCode(ctx, req.Code)
	if err == nil {
		return nil, errors.New("role code already exists")
	}
	// err != nil bisa error db lain atau not found
	// jika error bukan not found return error
	// Asumsi GetRoleByCode pakai pgx.ErrNoRows pada not found

	// Buat role baru
	roleRow, err := s.repo.CreateRole(ctx, db.CreateRoleParams{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &RoleResponse{
		ID:          roleRow.ID,
		Code:        roleRow.Code,
		Name:        roleRow.Name,
		Description: roleRow.Description,
		IsActive:    dbutil.BoolPtrValue(roleRow.IsActive, false),
		CreatedAt:   dbutil.PgTimeValue(roleRow.CreatedAt),
		UpdatedAt:   dbutil.PgTimeValue(roleRow.UpdatedAt),
	}, nil
}

func (s *service) GetRoleByID(ctx context.Context, id uuid.UUID) (*RoleResponse, error) {
	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &RoleResponse{
		ID:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		IsActive:    dbutil.BoolPtrValue(role.IsActive, false),
		CreatedAt:   dbutil.PgTimeValue(role.CreatedAt),
		UpdatedAt:   dbutil.PgTimeValue(role.UpdatedAt),
	}, nil
}

func (s *service) ListRoles(ctx context.Context) ([]RoleResponse, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		result = append(result, RoleResponse{
			ID:          r.ID,
			Code:        r.Code,
			Name:        r.Name,
			Description: r.Description,
			IsActive:    dbutil.BoolPtrValue(r.IsActive, false),
			CreatedAt:   dbutil.PgTimeValue(r.CreatedAt),
			UpdatedAt:   dbutil.PgTimeValue(r.UpdatedAt),
		})
	}

	return result, nil
}

func (s *service) UpdateRole(
	ctx context.Context,
	id uuid.UUID,
	req UpdateRoleRequest,
) error {

	// ensure exists
	if _, err := s.repo.GetRoleByID(ctx, id); err != nil {
		return err
	}

	_, err := s.repo.UpdateRole(ctx, db.UpdateRoleParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	return err
}

func (s *service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRole(ctx, id)
}
