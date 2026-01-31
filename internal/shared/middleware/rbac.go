package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireMenu(menuCode string, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetRoles(c)

		if len(roles) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		// TODO: Query database to check role_menus table
		// hasAccess := checkMenuAccess(roles, menuCode, permission)

		c.Next()
	}
}

// RequireRole checks if user has specific role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetRoles(c)

		hasRole := false
		for _, role := range roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
