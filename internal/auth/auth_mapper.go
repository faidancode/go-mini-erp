package auth

import (
	db "go-mini-erp/internal/shared/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
uuidFromPg mengubah pgtype.UUID (nullable)
menjadi *uuid.UUID untuk DTO / response layer
*/
func uuidFromPg(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

/*
mapRoles mengubah hasil query role
ke RoleInfo tanpa logic tambahan
*/
func mapRoles(rows []db.GetUserRolesRow) []RoleInfo {
	roles := make([]RoleInfo, 0, len(rows))

	for _, r := range rows {
		roles = append(roles, RoleInfo{
			ID:   r.ID,
			Code: r.Code,
			Name: r.Name,
		})
	}

	return roles
}

/*
mapMenus mengubah hasil query menu
nullable UUID dipetakan via helper, *string langsung dipakai
*/
func mapMenus(rows []db.GetUserMenusRow) []MenuInfo {
	menus := make([]MenuInfo, 0, len(rows))

	for _, m := range rows {
		menus = append(menus, MenuInfo{
			ID:        m.ID,
			ParentID:  uuidFromPg(m.ParentID),
			Code:      m.Code,
			Name:      m.Name,
			Path:      m.Path,
			Icon:      m.Icon,
			CanCreate: m.CanCreate,
			CanRead:   m.CanRead,
			CanUpdate: m.CanUpdate,
			CanDelete: m.CanDelete,
		})
	}

	return menus
}
