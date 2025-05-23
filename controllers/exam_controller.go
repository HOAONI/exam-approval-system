package controllers

import (
	"net/http"
	"strconv"

	"github.com/exam-approval-system/middlewares"
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

// ExamController 考试控制器
type ExamController struct {
	examService services.ExamService
	authService services.AuthService
}

// NewExamController 创建考试控制器
func NewExamController(examService services.ExamService, authService services.AuthService) *ExamController {
	return &ExamController{
		examService: examService,
		authService: authService,
	}
}

// RegisterRoutes 注册路由
func (c *ExamController) RegisterRoutes(router *gin.Engine) {
	exam := router.Group("/api/exams", middlewares.AuthMiddleware())
	{
		// 公共路由
		exam.GET("/:id", c.GetExam)
		exam.GET("/:id/comments", c.GetExamComments)

		// 学生路由
		student := exam.Group("/", middlewares.RoleMiddleware(models.RoleStudent))
		{
			student.GET("/published", c.ListPublishedExams)
		}

		// 教师路由
		teacher := exam.Group("/", middlewares.RoleMiddleware(models.RoleTeacher))
		{
			teacher.GET("/my", c.ListMyExams)
			teacher.POST("", c.CreateExam)
			teacher.PUT("/:id", c.UpdateExam)
			teacher.DELETE("/:id", c.DeleteExam)
			teacher.POST("/:id/submit", c.SubmitExam)
			teacher.POST("/:id/comment", c.AddComment)
		}

		// 管理员路由
		admin := exam.Group("/", middlewares.RoleMiddleware(models.RoleAdmin))
		{
			admin.GET("", c.ListAllExams)
			admin.GET("/pending", c.ListPendingExams)
			admin.POST("/:id/approve", c.ApproveExam)
			admin.POST("/:id/reject", c.RejectExam)
			admin.POST("/:id/publish", c.PublishExam)
			admin.POST("/:id/admincomment", c.AddComment)
		}
	}
}

// CreateExam 创建考试
func (c *ExamController) CreateExam(ctx *gin.Context) {
	var examReq struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Course      string `json:"course" binding:"required"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&examReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID, _ := ctx.Get("userID")

	// 解析时间字符串为time.Time
	startTime, err := parseTime(examReq.StartTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "开始时间格式错误"})
		return
	}
	endTime, err := parseTime(examReq.EndTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "结束时间格式错误"})
		return
	}

	exam := &models.Exam{
		Title:       examReq.Title,
		Description: examReq.Description,
		Course:      examReq.Course,
		StartTime:   startTime,
		EndTime:     endTime,
		CreatorID:   userID.(uint),
		Status:      models.StatusDraft,
	}

	if err := c.examService.CreateExam(exam); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, exam)
}

// GetExam 获取考试详情
func (c *ExamController) GetExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	exam, err := c.examService.GetExamByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	ctx.JSON(http.StatusOK, exam)
}

// UpdateExam 更新考试
func (c *ExamController) UpdateExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	var examReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Course      string `json:"course"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
	}

	if err := ctx.ShouldBindJSON(&examReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	exam, err := c.examService.GetExamByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查是否是考试的创建者
	userID, _ := ctx.Get("userID")
	if exam.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只有创建者才能修改考试"})
		return
	}

	// 更新字段
	if examReq.Title != "" {
		exam.Title = examReq.Title
	}
	if examReq.Description != "" {
		exam.Description = examReq.Description
	}
	if examReq.Course != "" {
		exam.Course = examReq.Course
	}
	if examReq.StartTime != "" {
		startTime, err := parseTime(examReq.StartTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "开始时间格式错误"})
			return
		}
		exam.StartTime = startTime
	}
	if examReq.EndTime != "" {
		endTime, err := parseTime(examReq.EndTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "结束时间格式错误"})
			return
		}
		exam.EndTime = endTime
	}

	if err := c.examService.UpdateExam(exam); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, exam)
}

// DeleteExam 删除考试
func (c *ExamController) DeleteExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	exam, err := c.examService.GetExamByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查是否是考试的创建者
	userID, _ := ctx.Get("userID")
	if exam.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只有创建者才能删除考试"})
		return
	}

	if err := c.examService.DeleteExam(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListAllExams 获取所有考试
func (c *ExamController) ListAllExams(ctx *gin.Context) {
	exams, err := c.examService.ListExams()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取考试列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, exams)
}

// ListMyExams 获取我创建的考试
func (c *ExamController) ListMyExams(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	exams, err := c.examService.ListExamsByCreator(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取考试列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, exams)
}

// ListPendingExams 获取待审批的考试
func (c *ExamController) ListPendingExams(ctx *gin.Context) {
	exams, err := c.examService.ListPendingExams()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取待审批考试列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, exams)
}

// ListPublishedExams 获取已发布的考试
func (c *ExamController) ListPublishedExams(ctx *gin.Context) {
	exams, err := c.examService.ListExamsByStatus(models.StatusPublished)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取已发布考试列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, exams)
}

// SubmitExam 提交考试审批
func (c *ExamController) SubmitExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	exam, err := c.examService.GetExamByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查是否是考试的创建者
	userID, _ := ctx.Get("userID")
	if exam.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只有创建者才能提交考试审批"})
		return
	}

	if err := c.examService.SubmitForApproval(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "提交成功"})
}

// ApproveExam 审批通过考试
func (c *ExamController) ApproveExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	var approveReq struct {
		Comment string `json:"comment"`
	}

	if err := ctx.ShouldBindJSON(&approveReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID, _ := ctx.Get("userID")
	if err := c.examService.ApproveExam(uint(id), userID.(uint), approveReq.Comment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "审批通过"})
}

// RejectExam 拒绝考试
func (c *ExamController) RejectExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	var rejectReq struct {
		Comment string `json:"comment" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&rejectReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID, _ := ctx.Get("userID")
	if err := c.examService.RejectExam(uint(id), userID.(uint), rejectReq.Comment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "已拒绝"})
}

// PublishExam 发布考试
func (c *ExamController) PublishExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	if err := c.examService.PublishExam(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "发布成功"})
}

// AddComment 添加评论
func (c *ExamController) AddComment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	var commentReq struct {
		Content string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&commentReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID, _ := ctx.Get("userID")
	comment := &models.Comment{
		ExamID:  uint(id),
		UserID:  userID.(uint),
		Content: commentReq.Content,
	}

	if err := c.examService.AddComment(comment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, comment)
}

// GetExamComments 获取考试评论
func (c *ExamController) GetExamComments(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	comments, err := c.examService.GetCommentsByExamID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论失败"})
		return
	}

	ctx.JSON(http.StatusOK, comments)
}
