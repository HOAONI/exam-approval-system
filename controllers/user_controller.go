package controllers

import (
	"net/http"
	"strconv"

	"github.com/exam-approval-system/middlewares"
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	userService services.UserService
	authService services.AuthService
}

// NewUserController 创建用户控制器
func NewUserController(userService services.UserService, authService services.AuthService) *UserController {
	return &UserController{
		userService: userService,
		authService: authService,
	}
}

// RegisterRoutes 注册路由
func (c *UserController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		authenticated := api.Group("", middlewares.AuthMiddleware())
		{
			// 用户个人资料相关API
			authenticated.GET("/user/profile", c.GetProfile)
			authenticated.PUT("/user/profile", c.UpdateProfile)

			// 管理员专用API
			admin := authenticated.Group("/admin", middlewares.RoleMiddleware(models.RoleAdmin))
			{
				admin.GET("/users", c.ListUsers)
				admin.GET("/teachers", c.ListTeachers)
				admin.GET("/students", c.ListStudents)
				admin.GET("/user/:id", c.GetUser)
				admin.PUT("/user/:id", c.UpdateUser)
			}
		}
	}
}

// GetProfile 获取当前用户信息
func (c *UserController) GetProfile(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	user, err := c.userService.GetUserByID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// UpdateProfile 更新当前用户信息
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")

	var updateReq struct {
		Name string `json:"name"`
	}

	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, err := c.userService.GetUserByID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	user.Name = updateReq.Name

	if err := c.userService.UpdateUser(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// ListUsers 获取所有用户列表
func (c *UserController) ListUsers(ctx *gin.Context) {
	users, err := c.userService.ListUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// ListTeachers 获取所有教师列表
func (c *UserController) ListTeachers(ctx *gin.Context) {
	teachers, err := c.userService.ListTeachers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取教师列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, teachers)
}

// ListStudents 获取所有学生列表
func (c *UserController) ListStudents(ctx *gin.Context) {
	students, err := c.userService.ListStudents()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, students)
}

// GetUser 根据ID获取用户
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// UpdateUser 更新用户信息（管理员专用）
func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var updateReq struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}

	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	user.Name = updateReq.Name
	if updateReq.Role != "" {
		user.Role = updateReq.Role
	}

	if err := c.userService.UpdateUser(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
