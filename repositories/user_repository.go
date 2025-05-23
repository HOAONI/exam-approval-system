package repositories

import (
	"github.com/exam-approval-system/configs"
	"github.com/exam-approval-system/models"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List() ([]models.User, error)
	ListByRole(role string) ([]models.User, error)
}

// userRepository 用户仓库实现
type userRepository struct{}

// NewUserRepository 创建用户仓库
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// Create 创建用户
func (r *userRepository) Create(user *models.User) error {
	return configs.DB.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := configs.DB.First(&user, id).Error
	return &user, err
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := configs.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}

// Update 更新用户
func (r *userRepository) Update(user *models.User) error {
	return configs.DB.Save(user).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	return configs.DB.Delete(&models.User{}, id).Error
}

// List 获取所有用户
func (r *userRepository) List() ([]models.User, error) {
	var users []models.User
	err := configs.DB.Find(&users).Error
	return users, err
}

// ListByRole 根据角色获取用户
func (r *userRepository) ListByRole(role string) ([]models.User, error) {
	var users []models.User
	err := configs.DB.Where("role = ?", role).Find(&users).Error
	return users, err
}
