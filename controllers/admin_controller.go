package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

// AdminController 管理员控制器
type AdminController struct {
	userService services.UserService
	authService services.AuthService
}

// NewAdminController 创建管理员控制器
func NewAdminController(userService services.UserService, authService services.AuthService) *AdminController {
	return &AdminController{
		userService: userService,
		authService: authService,
	}
}

// RegisterRoutes 注册管理员路由
func (c *AdminController) RegisterRoutes(router *gin.Engine) {
	admin := router.Group("/admin")
	{
		// 用户管理路由
		admin.GET("/users", c.ListUsers)
		admin.GET("/user/:id", c.GetUser)
		admin.POST("/user", c.CreateUser)
		admin.PUT("/user/:id", c.UpdateUser)
		admin.DELETE("/user/:id", c.DeleteUser)

		// 系统设置路由
		admin.GET("/settings", c.GetSettings)
		admin.POST("/settings", c.UpdateSettings)

		// 备份管理路由
		admin.POST("/backup", c.CreateBackup)
		admin.GET("/backups", c.ListBackups)
		admin.GET("/backup/:id", c.DownloadBackup)
	}
}

// ListUsers 获取用户列表
func (c *AdminController) ListUsers(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 此处直接检查用户角色，实际中应通过中间件进行身份验证
	user, err := c.userService.GetUserByUsername(username)
	if err != nil || user.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 获取筛选参数
	role := ctx.Query("role")
	status := ctx.Query("status")

	// 获取用户列表
	users, err := c.userService.ListUsersWithFilter(role, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   users,
	})
}

// GetUser 获取用户详情
func (c *AdminController) GetUser(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 获取用户ID
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 转换ID为uint
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取用户详情
	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
	})
}

// CreateUser 创建用户
func (c *AdminController) CreateUser(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 解析请求体
	var newUser models.User
	if err := ctx.ShouldBindJSON(&newUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户数据"})
		return
	}

	// 创建用户
	user, err := c.userService.CreateUser(&newUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"user":    user,
	})
}

// UpdateUser 更新用户
func (c *AdminController) UpdateUser(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 获取用户ID
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 解析请求体
	var updateData struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
		Password string `json:"password"`
		Status   string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户数据"})
		return
	}

	// 更新用户
	user, err := c.userService.UpdateUserDetails(
		userID,
		updateData.Name,
		updateData.Email,
		updateData.Phone,
		updateData.Role,
		updateData.Password,
		updateData.Status,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
	})
}

// DeleteUser 删除用户
func (c *AdminController) DeleteUser(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 获取用户ID
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 防止删除自己
	id, _ := strconv.ParseUint(userID, 10, 32)
	if uint(id) == currentUser.ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "不能删除当前登录的管理员账户"})
		return
	}

	// 删除用户
	err = c.userService.DeleteUser(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已成功删除",
	})
}

// GetSettings 获取系统设置
func (c *AdminController) GetSettings(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 这里应该从配置服务或数据库获取系统设置
	// 目前使用模拟数据
	settings := gin.H{
		"system_name":         "试卷审批管理系统",
		"admin_email":         "admin@example.com",
		"page_size":           10,
		"min_password_length": 8,
		"session_timeout":     30,
		"max_login_attempts":  5,
		"two_factor_auth":     false,
		"auto_backup":         true,
		"backup_frequency":    "weekly",
		"backup_count":        10,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

// UpdateSettings 更新系统设置
func (c *AdminController) UpdateSettings(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 解析请求体
	var settings map[string]interface{}
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的设置数据"})
		return
	}

	// 这里应该将设置保存到配置服务或数据库
	// 目前只返回成功响应
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置已成功更新",
	})
}

// CreateBackup 创建系统备份
func (c *AdminController) CreateBackup(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 这里应该触发系统备份过程
	// 目前只返回模拟的备份ID
	backupID := "backup_" + time.Now().Format("20060102_150405")

	ctx.JSON(http.StatusOK, gin.H{
		"success":   true,
		"backup_id": backupID,
		"message":   "备份已成功创建",
	})
}

// ListBackups 获取备份列表
func (c *AdminController) ListBackups(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 模拟备份列表
	now := time.Now()
	backups := []gin.H{
		{
			"id":         "backup_20230601_120000",
			"created_at": now.AddDate(0, 0, -7).Format(time.RFC3339),
			"size":       "2.5MB",
			"status":     "completed",
		},
		{
			"id":         "backup_20230608_120000",
			"created_at": now.AddDate(0, 0, -1).Format(time.RFC3339),
			"size":       "2.7MB",
			"status":     "completed",
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"backups": backups,
	})
}

// DownloadBackup 下载备份
func (c *AdminController) DownloadBackup(ctx *gin.Context) {
	// 身份验证检查
	username := ctx.GetHeader("X-Username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查用户权限
	currentUser, err := c.userService.GetUserByUsername(username)
	if err != nil || currentUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
		return
	}

	// 获取备份ID
	backupID := ctx.Param("id")
	if backupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "备份ID不能为空"})
		return
	}

	// 这里应该返回备份文件
	// 目前只返回下载链接
	ctx.JSON(http.StatusOK, gin.H{
		"success":      true,
		"download_url": "/backups/" + backupID + ".zip",
		"message":      "请点击链接下载备份文件",
	})
}
