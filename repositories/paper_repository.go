package repositories

import (
	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
)

// PaperRepository 试卷仓库接口
type PaperRepository interface {
	Create(paper *models.Paper) error
	GetByID(id uint) (*models.Paper, error)
	GetByExamID(examID uint) ([]models.Paper, error)
	Update(paper *models.Paper) error
	Delete(id uint) error
}

// paperRepository 试卷仓库实现
type paperRepository struct{}

// NewPaperRepository 创建试卷仓库
func NewPaperRepository() PaperRepository {
	return &paperRepository{}
}

// Create 创建试卷
func (r *paperRepository) Create(paper *models.Paper) error {
	return configs.DB.Create(paper).Error
}

// GetByID 根据ID获取试卷
func (r *paperRepository) GetByID(id uint) (*models.Paper, error) {
	var paper models.Paper
	err := configs.DB.First(&paper, id).Error
	return &paper, err
}

// GetByExamID 根据考试ID获取试卷
func (r *paperRepository) GetByExamID(examID uint) ([]models.Paper, error) {
	var papers []models.Paper
	err := configs.DB.Where("exam_id = ?", examID).Find(&papers).Error
	return papers, err
}

// Update 更新试卷
func (r *paperRepository) Update(paper *models.Paper) error {
	return configs.DB.Save(paper).Error
}

// Delete 删除试卷
func (r *paperRepository) Delete(id uint) error {
	return configs.DB.Delete(&models.Paper{}, id).Error
}
