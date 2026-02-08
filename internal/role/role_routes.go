package role

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	routes := r.Group("/roles")
	{
		routes.POST("", h.CreateRole)
		routes.GET("", h.ListRoles)
		routes.GET("/:id", h.GetRoleByID)
		routes.PUT("/:id", h.UpdateRole)
		routes.DELETE("/:id", h.DeleteRole)
	}
}
