package middlewares

import (
	"net/http"

	"github.com/exam-approval-system/repositories"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 简单认证中间件，基于请求头检查用户名
func AuthMiddleware() gin.HandlerFunc {
	userRepo := repositories.NewUserRepository()

	return func(c *gin.Context) {
		// 从请求头获取用户名
		username := c.GetHeader("X-Username")
		if username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户名"})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := userRepo.GetByUsername(username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", user.ID)
		c.Set("role", user.Role)
		c.Next()
	}
}

// RoleMiddleware 角色中间件，验证用户角色
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户角色
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		roleStr := role.(string)
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		// 如果角色不匹配，返回权限错误
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限访问该资源"})
		c.Abort()
	}
}
