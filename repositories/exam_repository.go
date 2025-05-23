package repositories

import (
	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
)

// ExamRepository 考试仓库接口
type ExamRepository interface {
	Create(exam *models.Exam) error
	GetByID(id uint) (*models.Exam, error)
	Update(exam *models.Exam) error
	Delete(id uint) error
	List() ([]models.Exam, error)
	ListByCreator(creatorID uint) ([]models.Exam, error)
	ListByStatus(status string) ([]models.Exam, error)
	ListPendingApproval() ([]models.Exam, error)
	ListPublished() ([]models.Exam, error)
	AddComment(comment *models.Comment) error
	GetCommentsByExamID(examID uint) ([]models.Comment, error)
	CreateExamData(examData *models.ExamData) error
	GetExamDataByExamAndStudent(examID, studentID uint) (*models.ExamData, error)
}

// examRepository 考试仓库实现
type examRepository struct{}

// NewExamRepository 创建考试仓库
func NewExamRepository() ExamRepository {
	return &examRepository{}
}

// Create 创建考试
func (r *examRepository) Create(exam *models.Exam) error {
	return configs.DB.Create(exam).Error
}

// GetByID 根据ID获取考试
func (r *examRepository) GetByID(id uint) (*models.Exam, error) {
	var exam models.Exam
	err := configs.DB.Preload("Creator").Preload("Approver").Preload("Papers").First(&exam, id).Error
	return &exam, err
}

// Update 更新考试
func (r *examRepository) Update(exam *models.Exam) error {
	return configs.DB.Save(exam).Error
}

// Delete 删除考试
func (r *examRepository) Delete(id uint) error {
	return configs.DB.Delete(&models.Exam{}, id).Error
}

// List 获取所有考试
func (r *examRepository) List() ([]models.Exam, error) {
	var exams []models.Exam
	err := configs.DB.Preload("Creator").Find(&exams).Error
	return exams, err
}

// ListByCreator 根据创建者获取考试
func (r *examRepository) ListByCreator(creatorID uint) ([]models.Exam, error) {
	var exams []models.Exam
	err := configs.DB.Where("creator_id = ?", creatorID).Preload("Creator").Find(&exams).Error
	return exams, err
}

// ListByStatus 根据状态获取考试
func (r *examRepository) ListByStatus(status string) ([]models.Exam, error) {
	var exams []models.Exam
	err := configs.DB.Where("status = ?", status).Preload("Creator").Find(&exams).Error
	return exams, err
}

// ListPendingApproval 获取待审批的考试
func (r *examRepository) ListPendingApproval() ([]models.Exam, error) {
	var exams []models.Exam
	err := configs.DB.Where("status = ?", models.StatusPending).Preload("Creator").Find(&exams).Error
	return exams, err
}

// ListPublished 获取已发布的考试
func (r *examRepository) ListPublished() ([]models.Exam, error) {
	var exams []models.Exam
	err := configs.DB.Where("status = ?", models.StatusPublished).Preload("Creator").Find(&exams).Error
	return exams, err
}

// AddComment 添加评论
func (r *examRepository) AddComment(comment *models.Comment) error {
	return configs.DB.Create(comment).Error
}

// GetCommentsByExamID 根据考试ID获取评论
func (r *examRepository) GetCommentsByExamID(examID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := configs.DB.Where("exam_id = ?", examID).Preload("User").Order("created_at desc").Find(&comments).Error
	return comments, err
}

// CreateExamData 创建试卷数据
func (r *examRepository) CreateExamData(examData *models.ExamData) error {
	return configs.DB.Create(examData).Error
}

// GetExamDataByExamAndStudent 根据考试ID和学生ID获取试卷数据
func (r *examRepository) GetExamDataByExamAndStudent(examID, studentID uint) (*models.ExamData, error) {
	var examData models.ExamData
	err := configs.DB.Where("exam_id = ? AND student_id = ?", examID, studentID).First(&examData).Error
	if err != nil {
		return nil, err
	}
	return &examData, nil
}
