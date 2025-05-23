package controllers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"log"

	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

// 服务依赖
var (
	AuthService      services.AuthService
	DashboardService services.DashboardService
	ExamService      services.ExamService
)

// LoginPage 登录页面
func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "用户登录",
	})
}

// RegisterPage 注册页面
func RegisterPage(c *gin.Context) {
	// 强制清除可能存在的缓存
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0, post-check=0, pre-check=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "-1")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Surrogate-Control", "no-store")
	c.Header("Vary", "*")

	// 添加随机参数避免缓存
	timestamp := time.Now().UnixNano()

	// 渲染注册页面
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title":     "用户注册",
		"timestamp": timestamp,           // 添加纳秒级时间戳完全防止缓存
		"random":    timestamp % 1000000, // 额外随机数
	})
}

// ForceRegisterPage 强制跳转到注册页面
func ForceRegisterPage(c *gin.Context) {
	// 强制清除所有缓存
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0, post-check=0, pre-check=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "-1")
	c.Header("X-Content-Type-Options", "nosniff")

	// 直接渲染注册页面而不是重定向，防止浏览器缓存问题
	timestamp := time.Now().UnixNano()
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title":     "用户注册",
		"timestamp": timestamp,
		"random":    timestamp % 1000000,
	})
}

// Dashboard 控制面板页面 - 根据用户请求头和查询参数决定显示内容
func Dashboard(c *gin.Context) {
	// 尝试从请求头获取用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		// 如果请求头没有，尝试从查询参数获取
		username = c.Query("username")
		if username == "" {
			// 如果都没有提供，重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			return
		}
	}

	// 使用仓库直接查询用户
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 根据用户角色选择合适的控制面板模板
	var template string
	var title string
	var dashboardData gin.H

	// 基础数据
	dashboardData = gin.H{
		"title": "控制面板",
		"user":  user,
	}

	switch user.Role {
	case "student":
		template = "dashboard-student.html"
		title = "学生控制面板"

		// 获取学生仪表板数据
		if DashboardService != nil {
			stats, err := DashboardService.GetStudentDashboardStats(user.ID)
			if err == nil {
				dashboardData["stats"] = stats
				dashboardData["totalPapers"] = stats.TotalPapers
				dashboardData["approvedPapers"] = stats.ApprovedPapers
				dashboardData["pendingPapers"] = stats.PendingPapers
				dashboardData["rejectedPapers"] = stats.RejectedPapers
				dashboardData["recentPapers"] = stats.RecentPapers
				dashboardData["averageScore"] = stats.AverageScore
				dashboardData["examDataList"] = stats.ExamDataList
			}
		}

	case "teacher":
		template = "dashboard-teacher.html"
		title = "教师控制面板"

		// 获取教师仪表板数据
		if DashboardService != nil {
			stats, err := DashboardService.GetTeacherDashboardStats(user.ID)
			if err == nil {
				dashboardData["stats"] = stats
				dashboardData["totalPapers"] = stats.TotalPapers
				dashboardData["approvedPapers"] = stats.ApprovedPapers
				dashboardData["pendingPapers"] = stats.PendingPaperList
				dashboardData["rejectedPapers"] = stats.RejectedPapers
				dashboardData["recentPapers"] = stats.RecentPapers
				dashboardData["totalStudents"] = stats.TotalStudents
				dashboardData["examDataList"] = stats.ExamDataList
			}
		}

	case "admin":
		template = "dashboard-admin.html" // 使用管理员面板
		title = "管理员控制面板"

		// 获取管理员仪表板数据
		if DashboardService != nil {
			stats, err := DashboardService.GetAdminDashboardStats()
			if err == nil {
				dashboardData["stats"] = stats
				dashboardData["totalPapers"] = stats.TotalPapers
				dashboardData["approvedPapers"] = stats.ApprovedPapers
				dashboardData["pendingPapers"] = stats.PendingPapers
				dashboardData["rejectedPapers"] = stats.RejectedPapers
				dashboardData["recentPapers"] = stats.RecentPapers
				dashboardData["totalStudents"] = stats.TotalStudents
				dashboardData["examDataList"] = stats.ExamDataList
				dashboardData["pendingPapers"] = stats.RecentPapers // 暂时用最近的试卷作为待审批列表
			}
		}

		// 添加当前时间用于系统状态显示
		dashboardData["now"] = time.Now()

	default:
		template = "dashboard.html"
		title = "控制面板"
	}

	dashboardData["title"] = title

	// 渲染对应的控制面板模板
	c.HTML(http.StatusOK, template, dashboardData)
}

// DashboardStudent 学生控制面板页面
func DashboardStudent(c *gin.Context) {
	// 尝试从请求头获取用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		// 如果请求头没有，尝试从查询参数获取
		username = c.Query("username")
		if username == "" {
			// 如果都没有提供，显示错误页面
			c.HTML(http.StatusOK, "dashboard-student.html", gin.H{
				"title": "学生控制面板",
				"error": "未登录或会话已过期",
			})
			return
		}
	}

	// 使用仓库直接查询用户
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.HTML(http.StatusOK, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "用户不存在",
		})
		return
	}

	// 准备仪表板数据
	dashboardData := gin.H{
		"title": "学生控制面板",
		"user":  user,
	}

	// 如果用户不是学生，显示错误信息
	if user.Role != "student" {
		dashboardData["error"] = "您没有权限访问学生控制面板"
		c.HTML(http.StatusOK, "dashboard-student.html", dashboardData)
		return
	}

	// 获取所有试卷，不再使用学生仪表板统计数据
	examRepo := repositories.NewExamRepository()

	// 打印当前学生信息
	log.Printf("学生信息: ID=%d, 用户名=%s", user.ID, user.Username)

	// 首先检查数据库中是否有已发布的试卷
	allExams, err := examRepo.List()
	log.Printf("数据库中所有试卷数量: %d", len(allExams))
	for i, exam := range allExams {
		log.Printf("试卷 #%d: ID=%d, 标题=%s, 状态=%s", i+1, exam.ID, exam.Title, exam.Status)
	}

	// 获取已发布的试卷
	publishedExams, err := examRepo.ListPublished()
	if err != nil {
		log.Printf("获取已发布试卷时出错: %v", err)
		dashboardData["error"] = "获取试卷列表失败: " + err.Error()
	} else {
		log.Printf("已发布试卷数量: %d", len(publishedExams))
		dashboardData["recentPapers"] = publishedExams
		dashboardData["totalPapers"] = len(publishedExams)

		// 默认统计数值
		dashboardData["approvedPapers"] = 0
		dashboardData["pendingPapers"] = 0
		dashboardData["rejectedPapers"] = 0
		dashboardData["averageScore"] = nil
	}

	// 获取学生已完成的试卷
	var completedExams []uint
	examDataRepo := repositories.NewExamDataRepository()
	allExamDataList, err := examDataRepo.GetExamsByStudentID(user.ID)

	// 确保所有已发布的试卷对学生可见，即使没有为其创建ExamData记录
	if len(publishedExams) > 0 {
		// 记录已经有ExamData记录的试卷ID，避免重复
		existingExamIDs := make(map[uint]bool)
		for _, data := range allExamDataList {
			existingExamIDs[data.ExamID] = true
		}

		// 为没有ExamData记录的已发布试卷创建记录并保存到数据库
		for _, exam := range publishedExams {
			if !existingExamIDs[exam.ID] {
				log.Printf("为学生创建并保存ExamData记录，试卷ID: %d, 创建者ID: %d", exam.ID, exam.CreatorID)
				examData := models.ExamData{
					ExamID:    exam.ID,
					StudentID: user.ID,
					Status:    "assigned",
					Title:     exam.Title,
					Course:    exam.Course,
				}

				// 保存到数据库
				err := examDataRepo.Create(&examData)
				if err != nil {
					log.Printf("保存ExamData记录失败, 试卷ID: %d, 创建者ID: %d, 错误: %v", exam.ID, exam.CreatorID, err)
				} else {
					// 添加到列表中
					allExamDataList = append(allExamDataList, examData)
					existingExamIDs[exam.ID] = true

					log.Printf("成功为学生ID=%d创建试卷ID=%d的ExamData记录，创建者为教师ID=%d", user.ID, exam.ID, exam.CreatorID)
				}
			}
		}
	}

	// 过滤出关联试卷仍然存在的ExamData记录
	var examDataList []models.ExamData
	if err == nil && len(allExamDataList) > 0 {
		for _, data := range allExamDataList {
			// 查询试卷是否存在
			exam, err := examRepo.GetByID(data.ExamID)
			if err == nil && exam != nil && exam.ID > 0 {
				// 试卷存在，添加到有效列表
				examDataList = append(examDataList, data)
			} else {
				log.Printf("过滤掉关联到不存在试卷的ExamData记录: ID=%d, 试卷ID=%d", data.ID, data.ExamID)
			}
		}
		log.Printf("过滤前 ExamData 数量: %d, 过滤后数量: %d", len(allExamDataList), len(examDataList))
	}

	if err == nil && len(examDataList) > 0 {
		dashboardData["examDataList"] = examDataList

		// 计算统计数据
		var approvedCount, pendingCount, rejectedCount int
		var totalScore float64
		var scoredCount int

		// 提取已完成的试卷ID列表并计算统计数据
		for _, examData := range examDataList {
			completedExams = append(completedExams, examData.ExamID)

			// 按状态分类
			switch examData.Status {
			case models.StatusApproved:
				approvedCount++
				totalScore += examData.TotalScore
				scoredCount++
			case models.StatusPending:
				pendingCount++
			case models.StatusRejected:
				rejectedCount++
			}
		}

		// 更新统计数据
		dashboardData["completedExams"] = completedExams
		dashboardData["approvedPapers"] = approvedCount
		dashboardData["pendingPapers"] = pendingCount
		dashboardData["rejectedPapers"] = rejectedCount

		// 计算平均分
		if scoredCount > 0 {
			averageScore := totalScore / float64(scoredCount)
			dashboardData["averageScore"] = averageScore
			log.Printf("学生 %s 的平均分: %.2f (共 %d 份已评分试卷)", user.Username, averageScore, scoredCount)
		} else {
			dashboardData["averageScore"] = nil
		}

		// 获取这些已完成试卷的详细信息，用于审批进度页面
		var submittedExams []models.Exam
		for _, examID := range completedExams {
			exam, err := examRepo.GetByID(examID)
			if err == nil {
				submittedExams = append(submittedExams, *exam) // 解引用指针
			}
		}
		dashboardData["submittedExams"] = submittedExams
	} else {
		// 如果没有完成任何试卷，则设置为空
		dashboardData["completedExams"] = []uint{}
		dashboardData["submittedExams"] = []models.Exam{}
		dashboardData["examDataList"] = []models.ExamData{}
	}

	// 确保已发布的试卷数据加载到模板中
	c.HTML(http.StatusOK, "dashboard-student.html", dashboardData)
}

// DashboardTeacher 教师控制面板页面
func DashboardTeacher(c *gin.Context) {
	// 尝试从请求头获取用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		// 如果请求头没有，尝试从查询参数获取
		username = c.Query("username")
		if username == "" {
			// 如果都没有提供，显示错误页面
			c.HTML(http.StatusOK, "dashboard-teacher.html", gin.H{
				"title": "教师控制面板",
				"error": "未登录或会话已过期",
				"stats": &services.DashboardStats{}, // 添加默认值
				"user":  &models.User{},             // 添加默认值
			})
			return
		}
	}

	// 使用仓库直接查询用户
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.HTML(http.StatusOK, "dashboard-teacher.html", gin.H{
			"title": "教师控制面板",
			"error": "用户不存在",
			"stats": &services.DashboardStats{}, // 添加默认值
			"user":  &models.User{},             // 添加默认值
		})
		return
	}

	// 准备仪表板数据
	dashboardData := gin.H{
		"title": "教师控制面板",
		"user":  user,
		"stats": &services.DashboardStats{}, // 添加默认值
	}

	// 如果用户不是教师，显示错误信息
	if user.Role != "teacher" {
		dashboardData["error"] = "您没有权限访问教师控制面板"
		c.HTML(http.StatusOK, "dashboard-teacher.html", dashboardData)
		return
	}

	// 获取教师关联的学生列表
	var students []models.User
	configs.DB.Where("teacher_id = ?", user.ID).Find(&students)

	// 查询所有学生并确保关联到当前教师（临时解决方案）
	var allStudents []models.User
	configs.DB.Where("role = ?", models.RoleStudent).Find(&allStudents)

	log.Printf("找到 %d 名学生待关联", len(allStudents))

	// 将这些学生关联到当前教师
	for _, student := range allStudents {
		if student.TeacherID != user.ID {
			student.TeacherID = user.ID
			configs.DB.Save(&student)
		}
	}

	// 更新仪表板数据中的学生列表
	dashboardData["students"] = allStudents

	// 获取教师仪表板数据
	if DashboardService != nil {
		stats, err := DashboardService.GetTeacherDashboardStats(user.ID)
		if err == nil {
			dashboardData["stats"] = stats // 覆盖默认值
			dashboardData["totalPapers"] = stats.TotalPapers
			dashboardData["approvedPapers"] = stats.ApprovedPapers
			dashboardData["pendingPapers"] = stats.PendingPaperList
			dashboardData["rejectedPapers"] = stats.RejectedPapers
			dashboardData["recentPapers"] = stats.RecentPapers
			dashboardData["totalStudents"] = len(students) // 使用关联学生数量
			dashboardData["examDataList"] = stats.ExamDataList
		} else {
			// 记录错误，但仍然使用默认的空 stats
			log.Printf("Error getting teacher dashboard stats for user %d: %v", user.ID, err)
			dashboardData["error"] = "获取仪表板数据失败"
		}
	} else {
		log.Println("DashboardService is nil")
		dashboardData["error"] = "仪表板服务未初始化"
	}

	// 获取教师的学生提交的试卷
	examDataRepo := repositories.NewExamDataRepository()
	pendingExamData, err := examDataRepo.ListByStatus(models.StatusPending)
	if err == nil {
		// 过滤出该教师负责的学生提交的试卷
		var teacherPendingExams []models.ExamData
		for _, data := range pendingExamData {
			// 如果学生关联到当前教师，或者教师是试卷创建者
			studentValid := data.Student.ID != 0 && data.Student.TeacherID == user.ID
			examValid := data.Exam.ID != 0 && data.Exam.CreatorID == user.ID
			if studentValid || examValid {
				teacherPendingExams = append(teacherPendingExams, data)
			}
		}

		dashboardData["pendingExamData"] = teacherPendingExams
		dashboardData["pendingExamsCount"] = len(teacherPendingExams)

		// 打印日志，帮助调试
		log.Printf("找到 %d 份待审批试卷", len(pendingExamData))
		log.Printf("当前教师负责的待审批试卷: %d 份", len(teacherPendingExams))
	} else {
		log.Printf("获取待审批试卷失败: %v", err)
	}

	c.HTML(http.StatusOK, "dashboard-teacher.html", dashboardData)
}

// DashboardAdmin 管理员控制面板页面
func DashboardAdmin(c *gin.Context) {
	// 尝试从请求头获取用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		// 如果请求头没有，尝试从查询参数获取
		username = c.Query("username")
		if username == "" {
			// 如果都没有提供，显示错误页面
			c.HTML(http.StatusOK, "dashboard-admin.html", gin.H{
				"title": "管理员控制面板",
				"error": "未登录或会话已过期",
				"now":   time.Now(),
			})
			return
		}
	}

	// 使用仓库直接查询用户
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.HTML(http.StatusOK, "dashboard-admin.html", gin.H{
			"title": "管理员控制面板",
			"error": "用户不存在",
			"now":   time.Now(),
		})
		return
	}

	// 准备仪表板数据
	dashboardData := gin.H{
		"title": "管理员控制面板",
		"user":  user,
		"now":   time.Now(),
	}

	// 如果用户不是管理员，显示错误信息
	if user.Role != "admin" {
		dashboardData["error"] = "您没有权限访问管理员控制面板"
		c.HTML(http.StatusOK, "dashboard-admin.html", dashboardData)
		return
	}

	// 获取管理员仪表板数据
	if DashboardService != nil {
		stats, err := DashboardService.GetAdminDashboardStats()
		if err == nil {
			dashboardData["stats"] = stats
			dashboardData["totalPapers"] = stats.TotalPapers
			dashboardData["approvedPapers"] = stats.ApprovedPapers
			dashboardData["totalStudents"] = stats.TotalStudents
		}
	}

	// 获取所有用户列表，用于用户管理
	allUsers, err := userRepo.List()
	if err == nil {
		dashboardData["allUsers"] = allUsers

		// 计算教师、学生和总人数
		var teacherCount, studentCount int
		for _, u := range allUsers {
			if u.Role == "teacher" {
				teacherCount++
			} else if u.Role == "student" {
				studentCount++
			}
		}

		dashboardData["teacherCount"] = teacherCount
		dashboardData["studentCount"] = studentCount
		dashboardData["totalUserCount"] = len(allUsers)
	}

	// 获取所有试卷列表，用于试卷管理
	examRepo := repositories.NewExamRepository()
	allPapers, err := examRepo.List()
	if err == nil {
		dashboardData["allPapers"] = allPapers
	}

	c.HTML(http.StatusOK, "dashboard-admin.html", dashboardData)
}

// HandleCreatePaper 处理创建试卷的请求
func HandleCreatePaper(c *gin.Context) {
	// 获取表单数据
	title := c.PostForm("title")
	course := c.PostForm("course")
	description := c.PostForm("description")

	// 获取当前用户名 (用于创建者ID和重定向时的会话保持)
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.Query("username") // 尝试从查询参数获取
		if username == "" {
			username = c.PostForm("username") // 尝试从表单数据获取
		}
	}

	// 是否请求JSON响应
	wantJSON := c.GetHeader("Accept") == "application/json" || c.GetHeader("X-Requested-With") == "XMLHttpRequest"

	// 尝试获取用户信息，无论后续操作是否成功，都可能需要在渲染错误页面时使用
	var currentUser *models.User
	if username != "" {
		userRepo := repositories.NewUserRepository() // 临时的repo实例
		currentUser, _ = userRepo.GetByUsername(username)
	}

	// 验证必填字段
	if title == "" || course == "" {
		if wantJSON {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "试卷标题和科目不能为空",
			})
		} else {
			c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
				"title": "管理员控制面板",
				"user":  currentUser, // 传递用户信息
				"error": "试卷标题和科目不能为空",
			})
		}
		return
	}

	// 如果此时仍没有获取到用户名，说明用户未认证或会话丢失
	if username == "" {
		if wantJSON {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户会话已过期或未登录",
			})
		} else {
			c.Redirect(http.StatusFound, "/login?error=session_expired_or_missing_username")
		}
		return
	}

	// 再次确认用户信息，并获取用户ID
	if currentUser == nil || currentUser.Username != username { // currentUser可能是之前获取失败的
		userRepo := repositories.NewUserRepository()
		var err error
		currentUser, err = userRepo.GetByUsername(username)
		if err != nil {
			if wantJSON {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "用户不存在",
				})
			} else {
				c.Redirect(http.StatusFound, "/login?error=user_not_found")
			}
			return
		}
	}

	// 创建试卷对象 (Exam)
	exam := &models.Exam{
		Title:       title,
		Description: description,
		Course:      course,
		CreatorID:   currentUser.ID,                     // 使用认证用户的ID
		Status:      models.StatusPublished,             // 直接设置为已发布状态
		StartTime:   time.Now(),                         // 可根据需求调整
		EndTime:     time.Now().Add(time.Hour * 24 * 7), // 默认有效期一周，可调整
		// CreatedAt 和 UpdatedAt 会由GORM的钩子自动处理
	}

	// 检查 ExamService 是否已初始化
	if ExamService == nil {
		if wantJSON {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "内部服务器错误: ExamService 未初始化",
			})
		} else {
			c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
				"title": "管理员控制面板",
				"user":  currentUser,
				"error": "内部服务器错误: ExamService 未初始化",
			})
		}
		return
	}

	// 使用 ExamService 保存到数据库
	err := ExamService.CreateExam(exam) // 假设CreateExam返回类型为 error
	if err != nil {
		if wantJSON {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "创建试卷失败: " + err.Error(),
			})
		} else {
			c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
				"title": "管理员控制面板",
				"user":  currentUser,
				"error": "创建试卷失败: " + err.Error(),
			})
		}
		return
	}

	// 根据请求类型返回响应
	if wantJSON {
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "试卷创建成功",
			"data": gin.H{
				"id":    exam.ID,
				"title": exam.Title,
			},
		})
	} else {
		// 重定向回管理员仪表板的试卷管理模块，并带上用户名和成功提示
		redirectURL := "/admin/dashboard?username=" + currentUser.Username + "&success=paper_created#papers"
		c.Redirect(http.StatusFound, redirectURL)
	}
}

// HandleDeletePaper 处理删除试卷的请求
func HandleDeletePaper(c *gin.Context) {
	// 获取试卷ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的试卷ID",
			})
		} else {
			c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
				"error": "无效的试卷ID",
			})
		}
		return
	}

	// 获取发起操作的用户名，用于重定向
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.Query("username") // 尝试从查询参数获取
		if username == "" {
			// 尝试从cookie中获取
			usernameCookie, err := c.Cookie("username")
			if err == nil && usernameCookie != "" {
				username = usernameCookie
			}
		}
		if username == "" {
			// 仍然没有找到用户名
			log.Printf("试卷删除操作未提供用户名，重定向到登录页面")
			// 判断请求类型
			if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "用户未登录或会话已过期",
				})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			return
		}
	}

	// 查找用户信息验证身份
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户未登录或会话已过期",
			})
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
		return
	}

	// 检查用户权限，只有管理员和教师可以删除试卷
	if user.Role != "admin" && user.Role != "teacher" {
		log.Printf("用户 %s 没有删除试卷的权限", username)
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "您没有删除试卷的权限",
			})
		} else {
			dashboardPath := "/dashboard"
			if user.Role == "student" {
				dashboardPath = "/dashboard-student"
			}
			c.Redirect(http.StatusFound, dashboardPath+"?username="+username+"&error=您没有删除试卷的权限")
		}
		return
	}

	// 获取试卷信息
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		log.Printf("查找试卷失败: %v", err)
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "试卷不存在",
			})
		} else {
			c.HTML(http.StatusNotFound, "dashboard-admin.html", gin.H{
				"error": "试卷不存在",
			})
		}
		return
	}

	// 删除与试卷相关的所有数据
	log.Printf("准备删除试卷ID: %d, 标题: %s", id, exam.Title)

	// 1. 查找和删除相关的ExamData记录
	var examDataList []models.ExamData
	result := configs.DB.Where("exam_id = ?", id).Find(&examDataList)
	if result.Error != nil {
		log.Printf("查询试卷相关ExamData记录失败: %v", result.Error)
	} else {
		log.Printf("找到%d条相关ExamData记录需要删除", len(examDataList))
		// 记录每一条要删除的ExamData信息用于调试
		for _, data := range examDataList {
			log.Printf("删除ExamData记录: ID=%d, 学生ID=%d, 试卷ID=%d, 标题=%s",
				data.ID, data.StudentID, data.ExamID, data.Title)
		}
	}

	// 执行删除
	deleteResult := configs.DB.Where("exam_id = ?", id).Delete(&models.ExamData{})
	log.Printf("删除ExamData结果: 影响行数=%d", deleteResult.RowsAffected)

	// 2. 删除相关的Comment记录
	result = configs.DB.Where("exam_id = ?", id).Delete(&models.Comment{})
	log.Printf("删除Comment结果: 影响行数=%d", result.RowsAffected)

	// 3. 删除相关的Paper记录
	result = configs.DB.Where("exam_id = ?", id).Delete(&models.Paper{})
	log.Printf("删除Paper结果: 影响行数=%d", result.RowsAffected)

	// 4. 最后删除试卷本身
	err = examRepo.Delete(uint(id))
	if err != nil {
		log.Printf("删除试卷失败: %v", err)
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "删除试卷失败: " + err.Error(),
			})
		} else {
			c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
				"error": "删除试卷失败: " + err.Error(),
			})
		}
		return
	}

	log.Printf("成功删除试卷ID: %d, 标题: %s", id, exam.Title)

	// 根据请求类型返回响应
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "试卷删除成功",
		})
	} else {
		// 根据用户角色选择合适的重定向URL
		redirectURL := "/dashboard-admin?username=" + username + "#papers"
		if user.Role == "teacher" {
			redirectURL = "/dashboard-teacher?username=" + username + "#papers"
		}
		// 添加成功消息
		redirectURL += "&success=1"
		c.Redirect(http.StatusFound, redirectURL)
	}
}

// HandleCreateUser 处理创建用户的请求
func HandleCreateUser(c *gin.Context) {
	// 获取表单数据
	username := c.PostForm("username")
	password := c.PostForm("password")
	name := c.PostForm("name")
	role := c.PostForm("role")

	// 验证必填字段
	if username == "" || password == "" || name == "" || role == "" {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "用户名、密码、姓名和角色不能为空",
		})
		return
	}

	// 验证角色值是否有效
	validRoles := map[string]bool{
		models.RoleStudent: true,
		models.RoleTeacher: true,
		models.RoleAdmin:   true,
	}

	if !validRoles[role] {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "无效的用户角色，必须是student、teacher或admin",
		})
		return
	}

	// 创建用户对象
	user := &models.User{
		Username: username,
		Password: password,
		Name:     name,
		Role:     role,
	}

	// 保存到数据库
	userRepo := repositories.NewUserRepository()
	err := userRepo.Create(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
			"error": "创建用户失败: " + err.Error(),
		})
		return
	}

	// 获取发起操作的管理员用户名，用于重定向
	actingAdminUsername := c.GetHeader("X-Username")
	if actingAdminUsername == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.Redirect(http.StatusFound, "/admin/dashboard?username="+actingAdminUsername+"#users")
}

// HandleApprovePaper 处理批准试卷的请求
func HandleApprovePaper(c *gin.Context) {
	// 获取试卷ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "无效的试卷ID",
		})
		return
	}

	// 获取当前用户ID
	username := c.GetHeader("X-Username")
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 获取试卷
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "dashboard-admin.html", gin.H{
			"error": "试卷不存在",
		})
		return
	}

	// 更新试卷状态为已审批
	exam.Status = models.StatusApproved
	exam.ApproverID = user.ID

	// 保存到数据库
	err = examRepo.Update(exam)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
			"error": "审批试卷失败: " + err.Error(),
		})
		return
	}

	// 添加审批评论（可选）
	comment := c.PostForm("comment")
	if comment != "" {
		commentObj := &models.Comment{
			ExamID:  exam.ID,
			UserID:  user.ID,
			Content: comment,
		}
		examRepo.AddComment(commentObj)
	}

	// 重定向回管理员仪表板，并显示审批管理模块
	c.Redirect(http.StatusFound, "/admin/dashboard?username="+username+"#approval")
}

// HandleRejectPaper 处理拒绝试卷的请求
func HandleRejectPaper(c *gin.Context) {
	// 获取试卷ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "无效的试卷ID",
		})
		return
	}

	// 获取当前用户ID
	username := c.GetHeader("X-Username")
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 获取试卷
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "dashboard-admin.html", gin.H{
			"error": "试卷不存在",
		})
		return
	}

	// 更新试卷状态为已拒绝
	exam.Status = models.StatusRejected
	exam.ApproverID = user.ID

	// 保存到数据库
	err = examRepo.Update(exam)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
			"error": "拒绝试卷失败: " + err.Error(),
		})
		return
	}

	// 添加拒绝理由（可选）
	comment := c.PostForm("comment")
	if comment != "" {
		commentObj := &models.Comment{
			ExamID:  exam.ID,
			UserID:  user.ID,
			Content: comment,
		}
		examRepo.AddComment(commentObj)
	}

	// 重定向回管理员仪表板，并显示审批管理模块
	c.Redirect(http.StatusFound, "/admin/dashboard?username="+username+"#approval")
}

// HandleChangePassword 处理修改密码的请求
func HandleChangePassword(c *gin.Context) {
	// 获取表单数据
	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	// 验证必填字段
	if oldPassword == "" || newPassword == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "所有密码字段均不能为空",
		})
		return
	}

	// 验证新密码和确认密码是否一致
	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "新密码和确认密码不一致",
		})
		return
	}

	// 获取当前用户
	username := c.GetHeader("X-Username")
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 验证旧密码
	if user.Password != oldPassword {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "旧密码不正确",
		})
		return
	}

	// 更新密码
	user.Password = newPassword
	err = userRepo.Update(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
			"error": "修改密码失败: " + err.Error(),
		})
		return
	}

	// 重定向回管理员仪表板，并显示个人中心模块
	c.Redirect(http.StatusFound, "/admin/dashboard?username="+username+"#profile&success=true")
}

// HandleDeleteUser 处理删除用户的请求
func HandleDeleteUser(c *gin.Context) {
	// 获取用户ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	// 获取当前管理员用户名（从URL查询参数获取，这是重点修复）
	username := c.Query("username")
	if username == "" {
		// 尝试从请求头获取
		username = c.GetHeader("X-Username")
		if username == "" {
			// 尝试从cookie中获取
			usernameCookie, err := c.Cookie("username")
			if err == nil && usernameCookie != "" {
				username = usernameCookie
			}
		}
		if username == "" {
			// 仍然没有找到用户名
			log.Printf("用户删除操作未提供用户名，重定向到登录页面")
			c.Redirect(http.StatusFound, "/login")
			return
		}
	}

	log.Printf("尝试删除用户ID: %d，操作人: %s", id, username)

	userRepo := repositories.NewUserRepository()
	adminUser, err := userRepo.GetByUsername(username)
	if err != nil {
		log.Printf("获取管理员用户信息失败: %v", err)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 验证当前用户是否为管理员
	if adminUser.Role != models.RoleAdmin {
		log.Printf("用户 %s 不是管理员，没有删除用户的权限", username)
		c.HTML(http.StatusForbidden, "dashboard-admin.html", gin.H{
			"error": "您没有删除用户的权限，此操作仅限管理员执行",
			"user":  adminUser,
		})
		return
	}

	// 检查用户是否存在
	targetUser, err := userRepo.GetByID(uint(id))
	if err != nil {
		log.Printf("要删除的用户(ID:%d)不存在: %v", id, err)
		c.HTML(http.StatusNotFound, "dashboard-admin.html", gin.H{
			"error": "用户不存在",
			"user":  adminUser,
		})
		return
	}

	// 防止管理员删除自己
	if adminUser.ID == targetUser.ID {
		log.Printf("管理员(%s)尝试删除自己的账户", username)
		c.HTML(http.StatusBadRequest, "dashboard-admin.html", gin.H{
			"error": "不能删除当前登录的管理员账户",
			"user":  adminUser,
		})
		return
	}

	log.Printf("准备删除用户ID: %d, 用户名: %s, 角色: %s", id, targetUser.Username, targetUser.Role)

	// 如果用户是学生，需要清除与该学生相关的考试数据
	if targetUser.Role == models.RoleStudent {
		result := configs.DB.Where("student_id = ?", id).Delete(&models.ExamData{})
		log.Printf("删除学生相关考试数据，影响行数: %d", result.RowsAffected)

		// 删除与该学生相关的评论
		result = configs.DB.Where("user_id = ?", id).Delete(&models.Comment{})
		log.Printf("删除学生评论数据，影响行数: %d", result.RowsAffected)
	}

	// 如果用户是教师，确保没有学生与该教师关联
	if targetUser.Role == models.RoleTeacher {
		// 查找关联到该教师的学生
		var students []models.User
		result := configs.DB.Where("teacher_id = ?", id).Find(&students)
		if result.Error == nil && len(students) > 0 {
			log.Printf("该教师有 %d 名关联学生，将学生的teacher_id设为0", len(students))
			// 更新关联的学生，将其teacher_id设为0
			result = configs.DB.Model(&models.User{}).Where("teacher_id = ?", id).Update("teacher_id", 0)
			log.Printf("更新学生的teacher_id，影响行数: %d", result.RowsAffected)
		}

		// 删除该教师创建的试卷以及相关数据
		var exams []models.Exam
		result = configs.DB.Where("creator_id = ?", id).Find(&exams)
		if result.Error == nil && len(exams) > 0 {
			log.Printf("该教师创建了 %d 份试卷，将删除相关数据", len(exams))

			for _, exam := range exams {
				// 删除试卷相关的ExamData
				result = configs.DB.Where("exam_id = ?", exam.ID).Delete(&models.ExamData{})
				log.Printf("删除试卷(%d)的ExamData，影响行数: %d", exam.ID, result.RowsAffected)

				// 删除试卷相关的Comment
				result = configs.DB.Where("exam_id = ?", exam.ID).Delete(&models.Comment{})
				log.Printf("删除试卷(%d)的Comment，影响行数: %d", exam.ID, result.RowsAffected)

				// 删除试卷相关的Paper
				result = configs.DB.Where("exam_id = ?", exam.ID).Delete(&models.Paper{})
				log.Printf("删除试卷(%d)的Paper，影响行数: %d", exam.ID, result.RowsAffected)
			}

			// 删除试卷
			result = configs.DB.Delete(&models.Exam{}, "creator_id = ?", id)
			log.Printf("删除教师创建的试卷，影响行数: %d", result.RowsAffected)
		}

		// 删除该教师的评论
		result = configs.DB.Where("user_id = ?", id).Delete(&models.Comment{})
		log.Printf("删除教师评论，影响行数: %d", result.RowsAffected)
	}

	// 删除用户
	err = userRepo.Delete(uint(id))
	if err != nil {
		log.Printf("删除用户失败: %v", err)
		c.HTML(http.StatusInternalServerError, "dashboard-admin.html", gin.H{
			"error": "删除用户失败: " + err.Error(),
			"user":  adminUser,
		})
		return
	}

	log.Printf("成功删除用户ID: %d, 用户名: %s, 角色: %s", id, targetUser.Username, targetUser.Role)

	// 重定向回管理员仪表板，并显示用户管理模块
	c.Redirect(http.StatusFound, "/dashboard-admin?username="+adminUser.Username+"&success=1#users")
}

// HandleViewPaper 处理查看试卷的请求
func HandleViewPaper(c *gin.Context) {
	// 获取试卷ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的试卷ID",
			})
		} else {
			c.HTML(http.StatusBadRequest, "dashboard-teacher.html", gin.H{
				"error": "无效的试卷ID",
			})
		}
		return
	}

	// 获取当前用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.Query("username") // 尝试从查询参数获取
	}

	if username == "" {
		// 如果都没有提供用户名，返回错误
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户未登录或会话已过期",
			})
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
		return
	}

	// 获取试卷信息
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "试卷不存在",
			})
		} else {
			c.HTML(http.StatusNotFound, "dashboard-teacher.html", gin.H{
				"error": "试卷不存在",
			})
		}
		return
	}

	// 返回试卷信息
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		c.JSON(http.StatusOK, exam)
	} else {
		c.HTML(http.StatusOK, "dashboard-teacher.html", gin.H{
			"title": "查看试卷",
			"exam":  exam,
		})
	}
}

// HandleUpdatePaper 处理更新试卷的请求
func HandleUpdatePaper(c *gin.Context) {
	// 获取试卷ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的试卷ID",
			})
		} else {
			c.HTML(http.StatusBadRequest, "dashboard-teacher.html", gin.H{
				"error": "无效的试卷ID",
			})
		}
		return
	}

	// 获取当前用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.Query("username") // 尝试从查询参数获取
	}

	if username == "" {
		// 如果都没有提供用户名，返回错误
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户未登录或会话已过期",
			})
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
		return
	}

	// 获取试卷信息
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "试卷不存在",
			})
		} else {
			c.HTML(http.StatusNotFound, "dashboard-teacher.html", gin.H{
				"error": "试卷不存在",
			})
		}
		return
	}

	// 从请求体获取更新数据
	var updateData struct {
		Title       string `json:"title"`
		Course      string `json:"course"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "请求格式错误",
			})
		} else {
			c.HTML(http.StatusBadRequest, "dashboard-teacher.html", gin.H{
				"error": "请求格式错误",
			})
		}
		return
	}

	// 更新试卷信息
	if updateData.Title != "" {
		exam.Title = updateData.Title
	}
	if updateData.Course != "" {
		exam.Course = updateData.Course
	}
	exam.Description = updateData.Description // 可以为空

	// 保存更新
	err = examRepo.Update(exam)
	if err != nil {
		// 判断请求类型
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "更新试卷失败: " + err.Error(),
			})
		} else {
			c.HTML(http.StatusInternalServerError, "dashboard-teacher.html", gin.H{
				"error": "更新试卷失败: " + err.Error(),
			})
		}
		return
	}

	// 返回成功响应
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "试卷更新成功",
			"data":    exam,
		})
	} else {
		redirectURL := "/dashboard-teacher?username=" + username + "#papers"
		c.Redirect(http.StatusFound, redirectURL)
	}
}

// HandleDistributePaper 处理教师分发试卷给学生的请求
func HandleDistributePaper(c *gin.Context) {
	// 从请求头获取教师用户名
	teacherUsername := c.GetHeader("X-Username")
	if teacherUsername == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录或会话已过期",
		})
		return
	}

	// 获取请求体中的数据
	var req struct {
		ExamID     uint   `json:"examId"`
		StudentIds []uint `json:"studentIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 验证试卷是否存在
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(req.ExamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "试卷不存在: " + err.Error(),
		})
		return
	}

	// 验证请求的用户是否是试卷的创建者或管理员
	userRepo := repositories.NewUserRepository()
	teacher, err := userRepo.GetByUsername(teacherUsername)
	if err != nil || (teacher.ID != exam.CreatorID && teacher.Role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权分发此试卷",
		})
		return
	}

	// 将试卷状态更新为已发布
	exam.Status = models.StatusPublished
	err = examRepo.Update(exam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新试卷状态失败: " + err.Error(),
		})
		return
	}

	// 如果没有选择特定学生，则分发给所有学生
	if len(req.StudentIds) == 0 {
		students, err := userRepo.ListByRole("student")
		if err == nil && len(students) > 0 {
			log.Printf("未选择特定学生，试卷(ID:%d)将自动分配给所有%d名学生", exam.ID, len(students))
			for _, student := range students {
				req.StudentIds = append(req.StudentIds, student.ID)
			}
		}
	}

	// 为每个选定的学生创建试卷分配记录
	for _, studentID := range req.StudentIds {
		// 验证学生是否存在
		_, err := userRepo.GetByID(studentID)
		if err != nil {
			continue // 跳过不存在的学生
		}

		// 创建ExamData记录，表示试卷分配给了学生
		examData := &models.ExamData{
			ExamID:    req.ExamID,
			StudentID: studentID,
			Status:    "assigned", // 已分配状态
		}

		examDataRepo := repositories.NewExamDataRepository()
		err = examDataRepo.Create(examData)
		if err != nil {
			// 记录错误但继续处理其他学生
			log.Printf("为学生 ID %d 分配试卷 ID %d 失败: %v", studentID, req.ExamID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "试卷已成功分发给所选学生",
	})
}

// HandleListStudents 获取所有学生列表
func HandleListStudents(c *gin.Context) {
	// 从请求头获取教师用户名
	teacherUsername := c.GetHeader("X-Username")
	if teacherUsername == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录或会话已过期",
		})
		return
	}

	// 验证请求者是否是教师或管理员
	userRepo := repositories.NewUserRepository()
	teacher, err := userRepo.GetByUsername(teacherUsername)
	if err != nil || (teacher.Role != "teacher" && teacher.Role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权访问学生列表",
		})
		return
	}

	// 获取所有学生
	allUsers, err := userRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户列表失败: " + err.Error(),
		})
		return
	}

	// 过滤出学生角色的用户
	var students []gin.H
	for _, user := range allUsers {
		if user.Role == "student" {
			students = append(students, gin.H{
				"id":       user.ID,
				"username": user.Username,
				"name":     user.Name,
			})
		}
	}

	c.JSON(http.StatusOK, students)
}

// HandleGetAssignedPapers 获取分配给学生的试卷列表
func HandleGetAssignedPapers(c *gin.Context) {
	// 从请求头获取学生用户名
	studentUsername := c.GetHeader("X-Username")
	if studentUsername == "" {
		studentUsername = c.Query("username")
		if studentUsername == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户未登录或会话已过期",
			})
			return
		}
	}

	// 验证请求者是否是学生
	userRepo := repositories.NewUserRepository()
	student, err := userRepo.GetByUsername(studentUsername)
	if err != nil || student.Role != "student" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "只有学生才能查看分配的试卷",
		})
		return
	}

	// 获取分配给该学生的所有试卷
	examDataRepo := repositories.NewExamDataRepository()
	assignedExams, err := examDataRepo.GetExamsByStudentID(student.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取分配试卷失败: " + err.Error(),
		})
		return
	}

	// 格式化响应数据
	var examList []gin.H
	for _, examData := range assignedExams {
		examList = append(examList, gin.H{
			"id":          examData.Exam.ID,
			"title":       examData.Exam.Title,
			"description": examData.Exam.Description,
			"course":      examData.Exam.Course,
			"status":      examData.Status,
			"createdAt":   examData.Exam.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    examList,
	})
}

// HandleExamView 处理学生查看考试页面的请求
func HandleExamView(c *gin.Context) {
	// 获取考试ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "无效的考试ID",
		})
		return
	}

	// 获取当前学生用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.Query("username") // 尝试从查询参数获取
	}

	if username == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 获取学生信息
	userRepo := repositories.NewUserRepository()
	student, err := userRepo.GetByUsername(username)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 验证用户是否是学生角色
	if student.Role != "student" {
		c.HTML(http.StatusForbidden, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "您没有权限参加考试",
		})
		return
	}

	// 获取考试信息
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "考试不存在",
		})
		return
	}

	// 验证考试是否处于已发布状态
	if exam.Status != "published" {
		c.HTML(http.StatusForbidden, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "该考试尚未发布，无法参加",
		})
		return
	}

	// 确保有ExamData记录
	var examDataId uint

	// 尝试从查询参数获取examDataId
	examDataIdStr := c.Query("examDataId")
	if examDataIdStr != "" {
		examDataIdUint, parseErr := strconv.ParseUint(examDataIdStr, 10, 32)
		if parseErr == nil {
			examDataId = uint(examDataIdUint)
		}
	}

	// 如果没有提供有效的examDataId，尝试查找或创建一个
	if examDataId == 0 {
		// 使用事务来确保原子性
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("开始事务失败: %v", tx.Error)
			c.HTML(http.StatusInternalServerError, "dashboard-student.html", gin.H{
				"title": "学生控制面板",
				"error": "系统错误，请稍后重试",
			})
			return
		}

		// 在事务中查找记录
		var existingExamData models.ExamData
		if err := tx.Where("exam_id = ? AND student_id = ?", exam.ID, student.ID).First(&existingExamData).Error; err == nil {
			// 如果找到记录，使用它
			examDataId = existingExamData.ID
			log.Printf("找到已存在的ExamData记录，ID: %d", examDataId)
			tx.Commit()
		} else {
			// 如果没有找到记录，创建一个新的
			newExamData := &models.ExamData{
				ExamID:    exam.ID,
				StudentID: student.ID,
				Status:    "assigned",
				Title:     exam.Title,
				Course:    exam.Course,
			}

			if err := tx.Create(newExamData).Error; err != nil {
				tx.Rollback()
				log.Printf("创建ExamData记录失败: %v", err)
				c.HTML(http.StatusInternalServerError, "dashboard-student.html", gin.H{
					"title": "学生控制面板",
					"error": "创建考试记录失败，请稍后重试",
				})
				return
			}

			examDataId = newExamData.ID
			log.Printf("创建新的ExamData记录，ID: %d", examDataId)
			tx.Commit()
		}
	}

	// 渲染考试页面，传递examDataId
	c.HTML(http.StatusOK, "exam.html", gin.H{
		"title":      exam.Title + " - 在线考试",
		"exam":       exam,
		"user":       student,
		"examDataId": examDataId,
	})
}

// HandleExamSubmit 处理学生提交考试答案的请求
func HandleExamSubmit(c *gin.Context) {
	// 获取考试ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "无效的考试ID",
		})
		return
	}

	// 获取当前学生用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.PostForm("username") // 从表单获取
		if username == "" {
			username = c.Query("username") // 尝试从查询参数获取
		}
	}

	if username == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 获取学生信息
	userRepo := repositories.NewUserRepository()
	student, err := userRepo.GetByUsername(username)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 验证用户是否是学生角色
	if student.Role != "student" {
		c.HTML(http.StatusForbidden, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "您没有权限提交考试答案",
		})
		return
	}

	// 获取考试信息
	examRepo := repositories.NewExamRepository()
	exam, err := examRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "考试不存在",
		})
		return
	}

	// 获取学生提交的答案
	answer := c.PostForm("answer")
	if answer == "" {
		c.HTML(http.StatusBadRequest, "exam.html", gin.H{
			"title": exam.Title + " - 在线考试",
			"exam":  exam,
			"user":  student,
			"error": "答案不能为空",
		})
		return
	}

	// 检查是否有已存在的ExamData记录
	examDataRepo := repositories.NewExamDataRepository()
	var examData *models.ExamData

	// 尝试从表单获取examDataId
	examDataIdStr := c.PostForm("examDataId")
	if examDataIdStr != "" {
		examDataIdUint, parseErr := strconv.ParseUint(examDataIdStr, 10, 32)
		if parseErr == nil {
			// 如果提供了有效的examDataId，尝试获取该记录
			existingExamData, getErr := examDataRepo.GetByID(uint(examDataIdUint))
			if getErr == nil && existingExamData.StudentID == student.ID && existingExamData.ExamID == exam.ID {
				examData = existingExamData
				log.Printf("使用已存在的ExamData记录(ID: %d)提交答案", examData.ID)
			}
		}
	}

	// 如果没有找到有效的ExamData记录，创建一个新的
	if examData == nil {
		examData = &models.ExamData{
			ExamID:     exam.ID,
			StudentID:  student.ID,
			Title:      exam.Title,
			Course:     exam.Course,
			TotalScore: 0.0,                  // 初始分数为0，等待教师批阅
			Status:     models.StatusPending, // 设置为待批阅状态
		}

		// 保存到数据库
		if err := configs.DB.Create(examData).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "dashboard-student.html", gin.H{
				"title": "学生控制面板",
				"error": "创建答题记录失败: " + err.Error(),
			})
			return
		}

		log.Printf("创建新的ExamData记录(ID: %d)提交答案", examData.ID)
	} else {
		// 如果是使用已有记录，更新状态为pending
		examData.Status = models.StatusPending
		if err := configs.DB.Save(examData).Error; err != nil {
			log.Printf("更新ExamData状态失败: %v", err)
		}
	}

	// 添加答案到Comment表中，关联到ExamData
	comment := &models.Comment{
		ExamID:    exam.ID,
		UserID:    student.ID,
		Content:   answer,
		CreatedAt: time.Now(),
	}

	// 保存答案内容
	if err := configs.DB.Create(comment).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard-student.html", gin.H{
			"title": "学生控制面板",
			"error": "提交答案失败: " + err.Error(),
		})
		return
	}

	// 添加日志记录
	log.Printf("学生 %s (ID: %d) 提交了考试 %s (ID: %d) 的答案，ExamData ID: %d",
		student.Username, student.ID, exam.Title, exam.ID, examData.ID)

	// 重定向回学生控制面板
	c.Redirect(http.StatusFound, "/dashboard-student?username="+username)
}

// HandleGetExamData 获取试卷数据和学生答案
func HandleGetExamData(c *gin.Context) {
	// 获取试卷数据ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的试卷数据ID",
		})
		return
	}

	// 验证教师身份
	username := c.GetHeader("X-Username")
	if username == "" {
		username = c.Query("username")
		if username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "未登录或会话已过期",
			})
			return
		}
	}

	userRepo := repositories.NewUserRepository()
	teacher, err := userRepo.GetByUsername(username)
	if err != nil || teacher.Role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权访问该资源",
		})
		return
	}

	// 获取试卷数据
	examDataRepo := repositories.NewExamDataRepository()
	examData, err := examDataRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "试卷数据不存在",
		})
		return
	}

	// 验证权限（只能查看自己负责的学生或自己创建的试卷）
	if examData.Student.TeacherID != teacher.ID && examData.Exam.CreatorID != teacher.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权查看该试卷数据",
		})
		return
	}

	// 查询学生答案
	commentRepo := repositories.NewCommentRepository()
	comments, err := commentRepo.GetCommentsByExamID(examData.ExamID)

	var answer string
	if err == nil && len(comments) > 0 {
		// 找到该学生提交的最新答案
		for i := len(comments) - 1; i >= 0; i-- {
			if comments[i].UserID == examData.StudentID {
				answer = comments[i].Content
				break
			}
		}
	}

	// 返回试卷数据和学生答案
	c.JSON(http.StatusOK, gin.H{
		"id":      examData.ID,
		"title":   examData.Title,
		"course":  examData.Course,
		"student": examData.Student,
		"exam":    examData.Exam,
		"status":  examData.Status,
		"answer":  answer,
		"score":   examData.TotalScore,
	})
}

// HandleGradeExam 处理教师评分
func HandleGradeExam(c *gin.Context) {
	// 打印原始请求体，帮助调试
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Printf("收到评分请求体: %s", string(bodyBytes))

	// 获取请求数据
	var req struct {
		ExamDataID uint    `json:"examDataId"` // 保持与前端字段名一致：examDataId
		Score      float64 `json:"score"`
		Comment    string  `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("评分请求绑定错误: %v, 请求体: %s", err, string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求数据: " + err.Error(),
		})
		return
	}

	log.Printf("成功解析评分请求: examDataId=%d, score=%.1f", req.ExamDataID, req.Score)

	// 验证评分是否在有效范围
	if req.Score < 0 || req.Score > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "评分必须在0-100之间",
		})
		return
	}

	// 验证教师身份
	username := c.GetHeader("X-Username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录或会话已过期",
		})
		return
	}

	userRepo := repositories.NewUserRepository()
	teacher, err := userRepo.GetByUsername(username)
	if err != nil || teacher.Role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权进行评分操作",
		})
		return
	}

	// 获取试卷数据
	examDataRepo := repositories.NewExamDataRepository()
	examData, err := examDataRepo.GetByID(req.ExamDataID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "试卷数据不存在",
		})
		return
	}

	// 验证权限
	if examData.Student.TeacherID != teacher.ID && examData.Exam.CreatorID != teacher.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权评分该试卷",
		})
		return
	}

	// 更新试卷数据的分数和状态
	examData.TotalScore = req.Score
	examData.Status = models.StatusApproved // 设置为已批阅状态
	examData.ApproverID = teacher.ID

	if err := examDataRepo.Update(examData); err != nil {
		log.Printf("更新试卷评分失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新试卷评分失败: " + err.Error(),
		})
		return
	}

	// 如果有评语，保存为评论
	if req.Comment != "" {
		commentRepo := repositories.NewCommentRepository()
		comment := &models.Comment{
			ExamID:    examData.ExamID,
			UserID:    teacher.ID,
			Content:   req.Comment,
			CreatedAt: time.Now(),
		}

		if err := commentRepo.Create(comment); err != nil {
			log.Printf("保存评语失败: %v", err)
			// 即使评语保存失败，也继续流程，不返回错误
		}
	}

	// 记录日志
	log.Printf("教师 %s(ID:%d) 对学生 %s(ID:%d) 的试卷 %s(ID:%d) 评分为 %.1f 分",
		teacher.Username, teacher.ID,
		examData.Student.Username, examData.Student.ID,
		examData.Title, examData.ExamID, req.Score)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "试卷评分成功",
	})
}

// HandleExamResult 处理查看试卷评分详情
func HandleExamResult(c *gin.Context) {
	// 获取试卷数据ID
	examDataID := c.Param("id")
	if examDataID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "试卷ID不能为空",
		})
		return
	}

	// 尝试从请求头获取用户名
	username := c.GetHeader("X-Username")
	if username == "" {
		// 如果请求头没有，尝试从查询参数获取
		username = c.Query("username")
		if username == "" {
			// 如果都没有提供，返回错误
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "用户名不能为空",
			})
			return
		}
	}

	// 验证用户身份
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户验证失败",
		})
		return
	}

	// 检查用户是否是学生
	if user.Role != "student" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "只有学生可以查看评分详情",
		})
		return
	}

	// 获取试卷数据
	examDataIDUint, err := strconv.ParseUint(examDataID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的试卷ID",
		})
		return
	}

	examDataRepo := repositories.NewExamDataRepository()
	examData, err := examDataRepo.GetByID(uint(examDataIDUint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "试卷数据不存在",
		})
		return
	}

	// 检查试卷是否属于当前学生
	if examData.StudentID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "您无权查看此试卷",
		})
		return
	}

	// u83b7u53d6u8bd5u5377u6570u636e
	examRepo := repositories.NewExamRepository()
	exam, examErr := examRepo.GetByID(examData.ExamID)
	if examErr != nil || exam == nil || exam.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "u8be5u8bd5u5377u5df2u4e0du5b58u5728uff0cu53efu80fdu5df2u88abu5220u9664",
		})
		return
	}

	// u68c0u67e5u8bd5u5377u662fu5426u5c5eu4e8eu5f53u524du5b66u751f
	if exam.Status != "published" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "该试卷尚未发布，无法查看评分详情",
		})
		return
	}

	// 准备基本的返回数据
	responseData := gin.H{
		"success":    true,
		"id":         examData.ID,
		"exam_id":    examData.ExamID,
		"title":      examData.Title,
		"course":     examData.Course,
		"status":     examData.Status,
		"created_at": examData.CreatedAt,
		"updated_at": examData.UpdatedAt,
	}

	// 仅当试卷已批阅时才返回评分和评语
	if examData.Status == models.StatusApproved {
		// 获取教师评语
		commentRepo := repositories.NewCommentRepository()
		comments, _ := commentRepo.ListByExamID(examData.ExamID)

		// 查找教师评语（教师的评论，非学生自己的评论）
		var comment string
		var teacherID uint = examData.ApproverID // 批阅教师ID

		if len(comments) > 0 {
			// 查找教师的最新评语
			var newestTeacherComment *models.Comment
			for i := range comments {
				// 只查找教师评论（非学生自己提交的答案）
				if comments[i].UserID == teacherID {
					if newestTeacherComment == nil || comments[i].CreatedAt.After(newestTeacherComment.CreatedAt) {
						newestTeacherComment = &comments[i]
					}
				}
			}

			if newestTeacherComment != nil {
				comment = newestTeacherComment.Content
			}
		}

		// 添加评分和评语信息
		responseData["score"] = examData.TotalScore
		responseData["comment"] = comment
	} else {
		// 未批阅或待批阅试卷不返回分数和评语
		responseData["score"] = 0
		responseData["comment"] = "试卷尚未批阅，请耐心等待"
	}

	// 返回试卷详情
	c.JSON(http.StatusOK, responseData)
}

// HandleGetStudentExams 获取学生提交并已批阅的试卷列表
func HandleGetStudentExams(c *gin.Context) {
	// 获取学生ID参数
	studentIDStr := c.Param("id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的学生ID",
		})
		return
	}

	// 获取教师用户名，验证权限
	teacherUsername := c.GetHeader("X-Username")
	if teacherUsername == "" {
		teacherUsername = c.Query("username")
		if teacherUsername == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户未登录或会话已过期",
			})
			return
		}
	}

	// 验证请求者是否是教师或管理员
	userRepo := repositories.NewUserRepository()
	teacher, err := userRepo.GetByUsername(teacherUsername)
	if err != nil || (teacher.Role != "teacher" && teacher.Role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权访问学生试卷",
		})
		return
	}

	// 验证学生是否存在
	student, err := userRepo.GetByID(uint(studentID))
	if err != nil || student.Role != "student" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "找不到指定的学生",
		})
		return
	}

	// 获取该学生的所有试卷数据
	examDataRepo := repositories.NewExamDataRepository()
	examDataList, err := examDataRepo.ListByStudent(uint(studentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取学生试卷失败: " + err.Error(),
		})
		return
	}

	// 获取评论数据
	commentRepo := repositories.NewCommentRepository()

	// 格式化试卷数据
	var formattedExams []gin.H
	for _, examData := range examDataList {
		// 获取试卷的评论
		comments, _ := commentRepo.GetCommentsByExamID(examData.ExamID)
		var answerText, commentText string

		// 查找学生的答案和教师评语
		for _, comment := range comments {
			// 通过用户ID查询用户角色
			user, _ := userRepo.GetByID(comment.UserID)
			if user != nil {
				if user.Role == "student" {
					answerText = comment.Content
				} else if user.Role == "teacher" {
					commentText = comment.Content
				}
			}
		}

		formattedExams = append(formattedExams, gin.H{
			"id":          examData.ID,
			"examId":      examData.ExamID,
			"title":       examData.Title,
			"course":      examData.Course,
			"studentId":   examData.StudentID,
			"studentName": student.Name,
			"status":      examData.Status,
			"score":       examData.TotalScore,
			"answer":      answerText,
			"comment":     commentText,
			"createdAt":   examData.CreatedAt.Format("2006-01-02 15:04:05"),
			"submittedAt": examData.CreatedAt.Format("2006-01-02 15:04:05"),
			"reviewedAt":  examData.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    formattedExams,
		"student": gin.H{
			"id":       student.ID,
			"name":     student.Name,
			"username": student.Username,
		},
	})
}
