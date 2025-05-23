package services

import (
	"errors"

	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
)

// ExamService 考试服务接口
type ExamService interface {
	CreateExam(exam *models.Exam) error
	GetExamByID(id uint) (*models.Exam, error)
	UpdateExam(exam *models.Exam) error
	DeleteExam(id uint) error
	ListExams() ([]models.Exam, error)
	ListExamsByCreator(creatorID uint) ([]models.Exam, error)
	ListExamsByStatus(status string) ([]models.Exam, error)
	ListPendingExams() ([]models.Exam, error)
	SubmitForApproval(examID uint) error
	ApproveExam(examID, approverID uint, comment string) error
	RejectExam(examID, approverID uint, comment string) error
	PublishExam(examID uint) error
	AddComment(comment *models.Comment) error
	GetCommentsByExamID(examID uint) ([]models.Comment, error)
}

// examService 考试服务实现
type examService struct {
	examRepository repositories.ExamRepository
	userRepository repositories.UserRepository
}

// NewExamService 创建考试服务
func NewExamService(examRepo repositories.ExamRepository, userRepo repositories.UserRepository) ExamService {
	return &examService{
		examRepository: examRepo,
		userRepository: userRepo,
	}
}

// CreateExam 创建考试
func (s *examService) CreateExam(exam *models.Exam) error {
	// 验证创建者是教师
	creator, err := s.userRepository.GetByID(exam.CreatorID)
	if err != nil {
		return errors.New("创建者不存在")
	}
	if creator.Role != models.RoleTeacher {
		return errors.New("只有教师才能创建考试")
	}

	// 创建试卷
	if err := s.examRepository.Create(exam); err != nil {
		return err
	}

	// 如果试卷状态为已发布，则自动分配给所有学生
	if exam.Status == models.StatusPublished {
		// 获取所有学生
		students, err := s.userRepository.ListByRole(models.RoleStudent)
		if err != nil {
			return err
		}

		// 为每个学生创建ExamData记录
		for _, student := range students {
			// 先检查是否已存在相同的记录
			existingExamData, err := s.examRepository.GetExamDataByExamAndStudent(exam.ID, student.ID)
			if err == nil && existingExamData != nil {
				// 如果记录已存在，跳过创建
				continue
			}

			examData := &models.ExamData{
				ExamID:     exam.ID,
				StudentID:  student.ID,
				Title:      exam.Title,
				Course:     exam.Course,
				TotalScore: exam.TotalScore,
				Status:     "assigned", // 直接设置为已分配状态
			}
			if err := s.examRepository.CreateExamData(examData); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetExamByID 根据ID获取考试
func (s *examService) GetExamByID(id uint) (*models.Exam, error) {
	return s.examRepository.GetByID(id)
}

// UpdateExam 更新考试
func (s *examService) UpdateExam(exam *models.Exam) error {
	// 只能修改草稿状态的考试
	currentExam, err := s.examRepository.GetByID(exam.ID)
	if err != nil {
		return err
	}
	if currentExam.Status != models.StatusDraft && currentExam.Status != models.StatusRejected {
		return errors.New("只能修改草稿或被拒绝状态的考试")
	}

	return s.examRepository.Update(exam)
}

// DeleteExam 删除考试
func (s *examService) DeleteExam(id uint) error {
	// 只能删除草稿状态的考试
	exam, err := s.examRepository.GetByID(id)
	if err != nil {
		return err
	}
	if exam.Status != models.StatusDraft {
		return errors.New("只能删除草稿状态的考试")
	}

	return s.examRepository.Delete(id)
}

// ListExams 获取所有考试
func (s *examService) ListExams() ([]models.Exam, error) {
	return s.examRepository.List()
}

// ListExamsByCreator 根据创建者获取考试
func (s *examService) ListExamsByCreator(creatorID uint) ([]models.Exam, error) {
	return s.examRepository.ListByCreator(creatorID)
}

// ListExamsByStatus 根据状态获取考试
func (s *examService) ListExamsByStatus(status string) ([]models.Exam, error) {
	return s.examRepository.ListByStatus(status)
}

// ListPendingExams 获取待审批的考试
func (s *examService) ListPendingExams() ([]models.Exam, error) {
	return s.examRepository.ListPendingApproval()
}

// SubmitForApproval 提交考试审批
func (s *examService) SubmitForApproval(examID uint) error {
	exam, err := s.examRepository.GetByID(examID)
	if err != nil {
		return err
	}
	if exam.Status != models.StatusDraft && exam.Status != models.StatusRejected {
		return errors.New("只能提交草稿或被拒绝状态的考试")
	}

	exam.Status = models.StatusPending
	return s.examRepository.Update(exam)
}

// ApproveExam 审批通过考试
func (s *examService) ApproveExam(examID, approverID uint, comment string) error {
	// 验证审批者是管理员
	approver, err := s.userRepository.GetByID(approverID)
	if err != nil {
		return errors.New("审批者不存在")
	}
	if approver.Role != models.RoleAdmin {
		return errors.New("只有管理员才能审批考试")
	}

	exam, err := s.examRepository.GetByID(examID)
	if err != nil {
		return err
	}
	if exam.Status != models.StatusPending {
		return errors.New("只能审批待审批状态的考试")
	}

	// 添加评论
	if comment != "" {
		commentObj := &models.Comment{
			ExamID:  examID,
			UserID:  approverID,
			Content: comment,
		}
		if err := s.examRepository.AddComment(commentObj); err != nil {
			return err
		}
	}

	// 更新考试状态
	exam.Status = models.StatusApproved
	exam.ApproverID = approverID
	return s.examRepository.Update(exam)
}

// RejectExam 拒绝考试
func (s *examService) RejectExam(examID, approverID uint, comment string) error {
	// 验证审批者是管理员
	approver, err := s.userRepository.GetByID(approverID)
	if err != nil {
		return errors.New("审批者不存在")
	}
	if approver.Role != models.RoleAdmin {
		return errors.New("只有管理员才能拒绝考试")
	}

	exam, err := s.examRepository.GetByID(examID)
	if err != nil {
		return err
	}
	if exam.Status != models.StatusPending {
		return errors.New("只能拒绝待审批状态的考试")
	}

	// 添加评论
	if comment != "" {
		commentObj := &models.Comment{
			ExamID:  examID,
			UserID:  approverID,
			Content: comment,
		}
		if err := s.examRepository.AddComment(commentObj); err != nil {
			return err
		}
	}

	// 更新考试状态
	exam.Status = models.StatusRejected
	exam.ApproverID = approverID
	return s.examRepository.Update(exam)
}

// PublishExam 发布考试
func (s *examService) PublishExam(examID uint) error {
	exam, err := s.examRepository.GetByID(examID)
	if err != nil {
		return err
	}
	if exam.Status != models.StatusApproved {
		return errors.New("只能发布已审批通过的考试")
	}

	exam.Status = models.StatusPublished
	return s.examRepository.Update(exam)
}

// AddComment 添加评论
func (s *examService) AddComment(comment *models.Comment) error {
	return s.examRepository.AddComment(comment)
}

// GetCommentsByExamID 根据考试ID获取评论
func (s *examService) GetCommentsByExamID(examID uint) ([]models.Comment, error) {
	return s.examRepository.GetCommentsByExamID(examID)
}
