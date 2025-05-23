package repositories

import (
	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
)

// ExamDataRepository 试卷数据仓库接口
type ExamDataRepository interface {
	Create(examData *models.ExamData) error
	GetByID(id uint) (*models.ExamData, error)
	Update(examData *models.ExamData) error
	Delete(id uint) error
	List() ([]models.ExamData, error)
	ListByStudent(studentID uint) ([]models.ExamData, error)
	ListByExam(examID uint) ([]models.ExamData, error)
	ListByStatus(status string) ([]models.ExamData, error)
	GetExamsByStudentID(studentID uint) ([]models.ExamData, error)
}

// examDataRepository 试卷数据仓库实现
type examDataRepository struct{}

// NewExamDataRepository 创建试卷数据仓库
func NewExamDataRepository() ExamDataRepository {
	return &examDataRepository{}
}

// Create 创建试卷数据
func (r *examDataRepository) Create(examData *models.ExamData) error {
	return configs.DB.Create(examData).Error
}

// GetByID 根据ID获取试卷数据
func (r *examDataRepository) GetByID(id uint) (*models.ExamData, error) {
	var examData models.ExamData
	err := configs.DB.Preload("Exam").Preload("Student").Preload("Approver").First(&examData, id).Error
	return &examData, err
}

// Update 更新试卷数据
func (r *examDataRepository) Update(examData *models.ExamData) error {
	return configs.DB.Save(examData).Error
}

// Delete 删除试卷数据
func (r *examDataRepository) Delete(id uint) error {
	return configs.DB.Delete(&models.ExamData{}, id).Error
}

// List 获取所有试卷数据
func (r *examDataRepository) List() ([]models.ExamData, error) {
	var examDataList []models.ExamData
	err := configs.DB.Preload("Student").Preload("Exam").Find(&examDataList).Error
	return examDataList, err
}

// ListByStudent 根据学生ID获取试卷数据
func (r *examDataRepository) ListByStudent(studentID uint) ([]models.ExamData, error) {
	var examDataList []models.ExamData
	err := configs.DB.Where("student_id = ?", studentID).Preload("Exam").Find(&examDataList).Error
	return examDataList, err
}

// ListByExam 根据考试ID获取试卷数据
func (r *examDataRepository) ListByExam(examID uint) ([]models.ExamData, error) {
	var examDataList []models.ExamData
	err := configs.DB.Where("exam_id = ?", examID).Preload("Student").Find(&examDataList).Error
	return examDataList, err
}

// ListByStatus 根据状态获取试卷数据
func (r *examDataRepository) ListByStatus(status string) ([]models.ExamData, error) {
	var examDataList []models.ExamData
	err := configs.DB.Where("status = ?", status).
		Preload("Student").
		Preload("Exam").
		Preload("Exam.Creator").
		Find(&examDataList).Error
	return examDataList, err
}

// GetExamsByStudentID 获取分配给学生的所有试卷
func (r *examDataRepository) GetExamsByStudentID(studentID uint) ([]models.ExamData, error) {
	var examDataList []models.ExamData
	err := configs.DB.Where("student_id = ?", studentID).
		Preload("Exam").
		Preload("Exam.Creator").
		Find(&examDataList).Error
	return examDataList, err
}
