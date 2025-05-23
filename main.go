package main

import (
	"log"
	"net/http"

	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/controllers"
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
	"github.com/exam-approval-system/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库连接
	configs.InitDB()
	defer configs.DB.Close()

	// 自动迁移数据库表结构
	configs.DB.AutoMigrate(&models.User{}, &models.Exam{}, &models.Paper{}, &models.Comment{}, &models.ExamData{})

	// 手动添加外键约束
	configs.DB.Model(&models.User{}).AddForeignKey("teacher_id", "users(id)", "SET NULL", "CASCADE")

	// 更新所有draft状态的试卷为published状态
	var draftExams []models.Exam
	result := configs.DB.Where("status = ?", "draft").Find(&draftExams)
	if result.Error == nil {
		log.Printf("找到 %d 份草稿状态的试卷，正在更新为已发布状态", len(draftExams))
		for _, exam := range draftExams {
			exam.Status = "published"
			configs.DB.Save(&exam)
		}
		log.Printf("已将所有草稿试卷更新为已发布状态")
	} else {
		log.Printf("查询草稿试卷时出错: %v", result.Error)
	}

	// 创建Gin路由引擎
	router := gin.Default()

	// 设置静态文件路径
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// 初始化存储库
	userRepo := repositories.NewUserRepository()
	examRepo := repositories.NewExamRepository()
	paperRepo := repositories.NewPaperRepository()
	examDataRepo := repositories.NewExamDataRepository()

	// 初始化服务
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	examService := services.NewExamService(examRepo, userRepo)
	paperService := services.NewPaperService(paperRepo, examRepo)
	dashboardService := services.NewDashboardService(examRepo, userRepo, paperRepo, examDataRepo)

	// 设置页面控制器的依赖项
	controllers.AuthService = authService
	controllers.DashboardService = dashboardService
	controllers.ExamService = examService

	// 初始化控制器
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService, authService)
	examController := controllers.NewExamController(examService, authService)
	paperController := controllers.NewPaperController(paperService, examService, authService)
	adminController := controllers.NewAdminController(userService, authService)

	// 注册API路由
	authController.RegisterRoutes(router)
	userController.RegisterRoutes(router)
	examController.RegisterRoutes(router)
	paperController.RegisterRoutes(router)
	adminController.RegisterRoutes(router)

	// 注册前端路由
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})
	router.GET("/login", controllers.LoginPage)
	router.GET("/register", controllers.RegisterPage)
	router.GET("/dashboard", controllers.Dashboard)

	// 添加基于角色的控制面板路由
	router.GET("/dashboard-student", controllers.DashboardStudent)
	router.GET("/dashboard-teacher", controllers.DashboardTeacher)
	router.GET("/dashboard-admin", controllers.DashboardAdmin)

	// 添加重定向路由
	router.GET("/redirect-to-register", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/register")
	})

	// 强制注册路由
	router.GET("/new-account", controllers.ForceRegisterPage)

	// 添加调试页面
	router.GET("/debug", func(c *gin.Context) {
		c.HTML(http.StatusOK, "debug.html", gin.H{
			"title": "调试页面",
		})
	})

	// 注册管理员相关路由
	adminRouterGroup := router.Group("/admin")
	adminRouterGroup.GET("/dashboard", controllers.DashboardAdmin)

	// 试卷管理路由
	adminRouterGroup.POST("/papers/create", controllers.HandleCreatePaper)
	adminRouterGroup.GET("/papers/delete/:id", controllers.HandleDeletePaper)

	// 用户管理路由
	adminRouterGroup.POST("/users/create", controllers.HandleCreateUser)
	adminRouterGroup.GET("/users/delete/:id", controllers.HandleDeleteUser)

	// 个人中心路由
	adminRouterGroup.POST("/profile/change-password", controllers.HandleChangePassword)

	// 添加教师专用路由组
	teacherRouterGroup := router.Group("/teacher")
	teacherRouterGroup.GET("/dashboard", controllers.DashboardTeacher)

	// 教师试卷管理路由
	teacherRouterGroup.POST("/papers/create", controllers.HandleCreatePaper)
	teacherRouterGroup.GET("/papers/delete/:id", controllers.HandleDeletePaper)
	teacherRouterGroup.GET("/papers/view/:id", controllers.HandleViewPaper)
	teacherRouterGroup.POST("/papers/update/:id", controllers.HandleUpdatePaper)

	// 教师批阅管理路由
	teacherRouterGroup.GET("/examdata/:id", controllers.HandleGetExamData)
	teacherRouterGroup.POST("/grade-exam", controllers.HandleGradeExam)

	// 教师学生管理路由
	teacherRouterGroup.GET("/student-exams/:id", controllers.HandleGetStudentExams)

	// 教师审批管理路由
	teacherRouterGroup.POST("/approve/:id", controllers.HandleApprovePaper)
	teacherRouterGroup.POST("/reject/:id", controllers.HandleRejectPaper)

	// 教师个人中心路由
	teacherRouterGroup.POST("/profile/change-password", controllers.HandleChangePassword)

	// 添加学生专用路由组
	studentRouterGroup := router.Group("/student")
	studentRouterGroup.GET("/dashboard", controllers.DashboardStudent)
	studentRouterGroup.GET("/exam/:id", controllers.HandleExamView)
	studentRouterGroup.POST("/submit-exam/:id", controllers.HandleExamSubmit)
	studentRouterGroup.GET("/exam-result/:id", controllers.HandleExamResult)

	// 启动服务器
	router.Run(":8080")
}
