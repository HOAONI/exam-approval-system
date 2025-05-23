package services

import (
	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
)

// 仪表板统计数据
type DashboardStats struct {
	TotalPapers      int               `json:"total_papers"`
	ApprovedPapers   int               `json:"approved_papers"`
	PendingPapers    int               `json:"pending_papers"`
	RejectedPapers   int               `json:"rejected_papers"`
	RecentPapers     []models.Exam     `json:"recent_papers"`
	PendingPaperList []models.Exam     `json:"pending_paper_list"`
	PapersBySubject  map[string]int    `json:"papers_by_subject"`
	AverageScore     float64           `json:"average_score"`
	TotalStudents    int               `json:"total_students"`
	ExamDataList     []models.ExamData `json:"exam_data_list"`
}

// DashboardService 仪表板服务接口
type DashboardService interface {
	GetStudentDashboardStats(userID uint) (*DashboardStats, error)
	GetTeacherDashboardStats(userID uint) (*DashboardStats, error)
	GetAdminDashboardStats() (*DashboardStats, error)
}

// dashboardService 仪表板服务实现
type dashboardService struct {
	examRepository     repositories.ExamRepository
	userRepository     repositories.UserRepository
	paperRepository    repositories.PaperRepository
	examDataRepository repositories.ExamDataRepository
}

// NewDashboardService 创建仪表板服务
func NewDashboardService(
	examRepo repositories.ExamRepository,
	userRepo repositories.UserRepository,
	paperRepo repositories.PaperRepository,
	examDataRepo repositories.ExamDataRepository,
) DashboardService {
	return &dashboardService{
		examRepository:     examRepo,
		userRepository:     userRepo,
		paperRepository:    paperRepo,
		examDataRepository: examDataRepo,
	}
}

// GetStudentDashboardStats 获取学生仪表板统计数据
func (s *dashboardService) GetStudentDashboardStats(userID uint) (*DashboardStats, error) {
	// 获取所有与学生相关的考试
	exams, err := s.examRepository.ListByCreator(userID)
	if err != nil {
		return nil, err
	}

	// 获取该学生的试卷数据
	examDataList, err := s.examDataRepository.ListByStudent(userID)
	if err != nil {
		return nil, err
	}

	var approvedCount, pendingCount, rejectedCount int
	var totalScore float64
	var scoreCount int
	papersBySubject := make(map[string]int)

	// 计算统计数据
	for _, exam := range exams {
		switch exam.Status {
		case models.StatusApproved, models.StatusPublished:
			approvedCount++
		case models.StatusPending:
			pendingCount++
		case models.StatusRejected:
			rejectedCount++
		}

		// 计算总分和平均分
		if exam.TotalScore > 0 {
			totalScore += exam.TotalScore
			scoreCount++
		}

		// 按科目统计
		if _, exists := papersBySubject[exam.Course]; exists {
			papersBySubject[exam.Course]++
		} else {
			papersBySubject[exam.Course] = 1
		}
	}

	// 计算平均分
	var averageScore float64
	if scoreCount > 0 {
		averageScore = totalScore / float64(scoreCount)
	}

	// 获取最近的试卷，最多5份
	var recentPapers []models.Exam
	if len(exams) > 0 {
		end := len(exams)
		if end > 5 {
			end = 5
		}
		recentPapers = exams[:end]
	}

	return &DashboardStats{
		TotalPapers:     len(exams),
		ApprovedPapers:  approvedCount,
		PendingPapers:   pendingCount,
		RejectedPapers:  rejectedCount,
		RecentPapers:    recentPapers,
		PapersBySubject: papersBySubject,
		AverageScore:    averageScore,
		ExamDataList:    examDataList,
	}, nil
}

// GetTeacherDashboardStats 获取教师仪表板统计数据
func (s *dashboardService) GetTeacherDashboardStats(userID uint) (*DashboardStats, error) {
	// 获取所有与教师相关的考试
	exams, err := s.examRepository.ListByCreator(userID)
	if err != nil {
		return nil, err
	}

	// 获取所有学生用户数量
	students, err := s.userRepository.ListByRole(models.RoleStudent)
	if err != nil {
		return nil, err
	}

	// 获取所有试卷数据
	examDataList, err := s.examDataRepository.List()
	if err != nil {
		examDataList = []models.ExamData{}
	}

	var approvedCount, pendingCount, rejectedCount int
	var pendingPaperList []models.Exam
	papersBySubject := make(map[string]int)

	// 计算统计数据并收集待审批试卷
	for _, exam := range exams {
		switch exam.Status {
		case models.StatusApproved, models.StatusPublished:
			approvedCount++
		case models.StatusPending:
			pendingCount++
			pendingPaperList = append(pendingPaperList, exam)
		case models.StatusRejected:
			rejectedCount++
		}

		// 按科目统计
		if _, exists := papersBySubject[exam.Course]; exists {
			papersBySubject[exam.Course]++
		} else {
			papersBySubject[exam.Course] = 1
		}
	}

	// 获取最近的试卷，最多5份
	var recentPapers []models.Exam
	if len(exams) > 0 {
		end := len(exams)
		if end > 5 {
			end = 5
		}
		recentPapers = exams[:end]
	}

	return &DashboardStats{
		TotalPapers:      len(exams),
		ApprovedPapers:   approvedCount,
		PendingPapers:    pendingCount,
		RejectedPapers:   rejectedCount,
		RecentPapers:     recentPapers,
		PendingPaperList: pendingPaperList,
		PapersBySubject:  papersBySubject,
		TotalStudents:    len(students),
		ExamDataList:     examDataList,
	}, nil
}

// GetAdminDashboardStats 获取管理员仪表板统计数据
func (s *dashboardService) GetAdminDashboardStats() (*DashboardStats, error) {
	// 获取所有考试
	exams, err := s.examRepository.List()
	if err != nil {
		return nil, err
	}

	// 获取所有学生用户数量
	students, err := s.userRepository.ListByRole(models.RoleStudent)
	if err != nil {
		return nil, err
	}

	// 获取所有试卷数据
	examDataList, err := s.examDataRepository.List()
	if err != nil {
		return nil, err
	}

	var approvedCount, pendingCount, rejectedCount int
	papersBySubject := make(map[string]int)

	// 计算统计数据
	for _, exam := range exams {
		switch exam.Status {
		case models.StatusApproved, models.StatusPublished:
			approvedCount++
		case models.StatusPending:
			pendingCount++
		case models.StatusRejected:
			rejectedCount++
		}

		// 按科目统计
		if _, exists := papersBySubject[exam.Course]; exists {
			papersBySubject[exam.Course]++
		} else {
			papersBySubject[exam.Course] = 1
		}
	}

	// 获取最近的试卷，最多5份
	var recentPapers []models.Exam
	if len(exams) > 0 {
		end := len(exams)
		if end > 5 {
			end = 5
		}
		recentPapers = exams[:end]
	}

	return &DashboardStats{
		TotalPapers:     len(exams),
		ApprovedPapers:  approvedCount,
		PendingPapers:   pendingCount,
		RejectedPapers:  rejectedCount,
		RecentPapers:    recentPapers,
		PapersBySubject: papersBySubject,
		TotalStudents:   len(students),
		ExamDataList:    examDataList,
	}, nil
}
