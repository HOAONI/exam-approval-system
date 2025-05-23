package repositories

import (
	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
)

// CommentRepository 定义评论仓库接口
type CommentRepository interface {
	Create(comment *models.Comment) error
	GetCommentsByUserID(userID uint) ([]models.Comment, error)
	GetCommentsByExamID(examID uint) ([]models.Comment, error)
	ListByExamID(examID uint) ([]models.Comment, error)
}

// commentRepository 实现评论仓库接口
type commentRepository struct{}

// NewCommentRepository 创建一个新的评论仓库实例
func NewCommentRepository() CommentRepository {
	return &commentRepository{}
}

// Create 创建新的评论
func (r *commentRepository) Create(comment *models.Comment) error {
	return configs.DB.Create(comment).Error
}

// GetCommentsByUserID 获取指定用户的所有评论
func (r *commentRepository) GetCommentsByUserID(userID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := configs.DB.Where("user_id = ?", userID).Find(&comments).Error
	return comments, err
}

// GetCommentsByExamID 获取指定考试的所有评论
func (r *commentRepository) GetCommentsByExamID(examID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := configs.DB.Where("exam_id = ?", examID).Find(&comments).Error
	return comments, err
}

// ListByExamID 获取指定试卷的所有评论
func (r *commentRepository) ListByExamID(examID uint) ([]models.Comment, error) {
	var comments []models.Comment
	result := configs.DB.Where("exam_id = ?", examID).Find(&comments)
	return comments, result.Error
}
