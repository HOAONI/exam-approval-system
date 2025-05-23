package controllers

import (
	"net/http"
	"strconv"

	"github.com/exam-approval-system/middlewares"
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

// PaperController 试卷控制器
type PaperController struct {
	paperService services.PaperService
	examService  services.ExamService
	authService  services.AuthService
}

// NewPaperController 创建试卷控制器
func NewPaperController(paperService services.PaperService, examService services.ExamService, authService services.AuthService) *PaperController {
	return &PaperController{
		paperService: paperService,
		examService:  examService,
		authService:  authService,
	}
}

// RegisterRoutes 注册路由
func (c *PaperController) RegisterRoutes(router *gin.Engine) {
	paper := router.Group("/api/papers", middlewares.AuthMiddleware())
	{
		// 公共路由
		paper.GET("/:id", c.GetPaper)
		paper.GET("/exam/:exam_id", c.GetPapersByExam)
		paper.GET("/:id/verify", c.VerifyPaperSignature)

		// 教师路由
		teacher := paper.Group("/", middlewares.RoleMiddleware(models.RoleTeacher))
		{
			teacher.POST("", c.CreatePaper)
			teacher.PUT("/:id", c.UpdatePaper)
			teacher.DELETE("/:id", c.DeletePaper)
			teacher.POST("/:id/sign", c.SignPaper)
		}
	}
}

// CreatePaper 创建试卷
func (c *PaperController) CreatePaper(ctx *gin.Context) {
	var paperReq struct {
		ExamID       uint    `json:"exam_id" binding:"required"`
		Title        string  `json:"title" binding:"required"`
		Content      string  `json:"content"`
		Questions    string  `json:"questions"`
		Duration     int     `json:"duration" binding:"required"`
		TotalScore   float64 `json:"total_score" binding:"required"`
		PassingScore float64 `json:"passing_score" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&paperReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 检查关联的考试是否存在
	exam, err := c.examService.GetExamByID(paperReq.ExamID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查当前用户是否是考试的创建者
	userID, _ := ctx.Get("userID")
	if exam.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只有考试创建者才能添加试卷"})
		return
	}

	paper := &models.Paper{
		ExamID:       paperReq.ExamID,
		Title:        paperReq.Title,
		Content:      paperReq.Content,
		Questions:    paperReq.Questions,
		Duration:     paperReq.Duration,
		TotalScore:   paperReq.TotalScore,
		PassingScore: paperReq.PassingScore,
		Status:       models.StatusDraft,
	}

	if err := c.paperService.CreatePaper(paper); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, paper)
}

// GetPaper 获取试卷详情
func (c *PaperController) GetPaper(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的试卷ID"})
		return
	}

	paper, err := c.paperService.GetPaperByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "试卷不存在"})
		return
	}

	ctx.JSON(http.StatusOK, paper)
}

// GetPapersByExam 获取考试相关的试卷
func (c *PaperController) GetPapersByExam(ctx *gin.Context) {
	examIDStr := ctx.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试ID"})
		return
	}

	// 检查考试是否存在
	exam, err := c.examService.GetExamByID(uint(examID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查用户权限
	userID, _ := ctx.Get("userID")
	role, _ := ctx.Get("role")

	// 只有考试创建者、管理员或者已发布考试的学生才能查看试卷
	if exam.CreatorID != userID.(uint) &&
		role.(string) != models.RoleAdmin &&
		!(role.(string) == models.RoleStudent && exam.Status == models.StatusPublished) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限查看试卷"})
		return
	}

	papers, err := c.paperService.GetPapersByExamID(uint(examID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取试卷失败"})
		return
	}

	ctx.JSON(http.StatusOK, papers)
}

// UpdatePaper 更新试卷
func (c *PaperController) UpdatePaper(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的试卷ID"})
		return
	}

	var paperReq struct {
		Title        string  `json:"title"`
		Content      string  `json:"content"`
		Questions    string  `json:"questions"`
		Duration     int     `json:"duration"`
		TotalScore   float64 `json:"total_score"`
		PassingScore float64 `json:"passing_score"`
	}

	if err := ctx.ShouldBindJSON(&paperReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	paper, err := c.paperService.GetPaperByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "试卷不存在"})
		return
	}

	// 检查考试状态和用户权限
	exam, err := c.examService.GetExamByID(paper.ExamID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查当前用户是否是考试的创建者
	userID, _ := ctx.Get("userID")
	if exam.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只有考试创建者才能修改试卷"})
		return
	}

	// 更新字段
	if paperReq.Title != "" {
		paper.Title = paperReq.Title
	}
	if paperReq.Content != "" {
		paper.Content = paperReq.Content
	}
	if paperReq.Questions != "" {
		paper.Questions = paperReq.Questions
	}
	if paperReq.Duration > 0 {
		paper.Duration = paperReq.Duration
	}
	if paperReq.TotalScore > 0 {
		paper.TotalScore = paperReq.TotalScore
	}
	if paperReq.PassingScore > 0 {
		paper.PassingScore = paperReq.PassingScore
	}

	if err := c.paperService.UpdatePaper(paper); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, paper)
}

// DeletePaper 删除试卷
func (c *PaperController) DeletePaper(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的试卷ID"})
		return
	}

	paper, err := c.paperService.GetPaperByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "试卷不存在"})
		return
	}

	// 检查考试状态和用户权限
	exam, err := c.examService.GetExamByID(paper.ExamID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查当前用户是否是考试的创建者
	userID, _ := ctx.Get("userID")
	if exam.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只有考试创建者才能删除试卷"})
		return
	}

	if err := c.paperService.DeletePaper(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SignPaper 为试卷签名
func (c *PaperController) SignPaper(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的试卷ID"})
		return
	}

	// 获取当前用户ID
	userID, _ := ctx.Get("userID")

	// 为试卷签名
	err = c.paperService.SignPaper(uint(id), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "试卷签名成功",
	})
}

// VerifyPaperSignature 验证试卷签名
func (c *PaperController) VerifyPaperSignature(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的试卷ID"})
		return
	}

	// 验证签名
	isValid, err := c.paperService.VerifyPaperSignature(uint(id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"is_valid": isValid,
	})
}
