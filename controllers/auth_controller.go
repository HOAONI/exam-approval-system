package controllers

import (
	"net/http"

	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	authService services.AuthService
	userRepo    repositories.UserRepository
}

// NewAuthController 创建认证控制器
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
		userRepo:    repositories.NewUserRepository(),
	}
}

// RegisterRoutes 注册路由
func (c *AuthController) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", c.Login)
		auth.POST("/register", c.Register)
		auth.GET("/logout", c.Logout)
		auth.GET("/check", c.CheckAuth)
	}
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}

	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, err := c.authService.Login(loginReq.Username, loginReq.Password, loginReq.Role)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 返回用户信息，不设置cookie，由前端管理会话
	ctx.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"role":     user.Role,
		},
	})
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
	var registerReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&registerReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证角色值是否有效
	validRoles := map[string]bool{
		models.RoleStudent: true,
		models.RoleTeacher: true,
		models.RoleAdmin:   true,
	}

	if !validRoles[registerReq.Role] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户角色，必须是student、teacher或admin"})
		return
	}

	user := &models.User{
		Username: registerReq.Username,
		Password: registerReq.Password,
		Name:     registerReq.Name,
		Role:     registerReq.Role,
	}

	if err := c.authService.Register(user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果注册的是教师，自动将所有学生与该教师关联
	if user.Role == models.RoleTeacher {
		var allStudents []models.User
		configs.DB.Where("role = ?", models.RoleStudent).Find(&allStudents)

		if len(allStudents) > 0 {
			// 将所有学生关联到当前教师
			for _, student := range allStudents {
				student.TeacherID = user.ID
				configs.DB.Save(&student)
			}
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"role":     user.Role,
		},
	})
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	// 客户端处理登出逻辑，这里只返回成功消息
	ctx.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// CheckAuth 检查认证状态
func (c *AuthController) CheckAuth(ctx *gin.Context) {
	// 基于请求头的简单认证
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"authenticated": false,
			"error":         "未提供用户名",
		})
		return
	}

	// 根据用户名获取用户信息
	user, err := c.userRepo.GetByUsername(username)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"authenticated": false,
			"error":         "用户不存在",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"role":     user.Role,
		},
	})
}
