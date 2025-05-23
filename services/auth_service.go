package services

import (
	"errors"
	"fmt"

	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(username, password, role string) (*models.User, error)
	Register(user *models.User) error
	GetUserProfile(userID uint) (*models.User, error)
}

// authService 认证服务实现
type authService struct {
	userRepository repositories.UserRepository
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{
		userRepository: userRepo,
	}
}

// Login 用户登录
func (s *authService) Login(username, password, role string) (*models.User, error) {
	user, err := s.userRepository.GetByUsername(username)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 验证密码
	if err := user.CheckPassword(password); err != nil {
		return nil, errors.New("密码错误")
	}

	// 验证角色（仅当用户提供了角色时才验证）
	if role != "" && user.Role != role {
		return nil, errors.New("用户角色不匹配")
	}

	return user, nil
}

// Register 用户注册
func (s *authService) Register(user *models.User) error {
	// 检查用户名是否已存在
	existingUser, err := s.userRepository.GetByUsername(user.Username)
	if err == nil && existingUser.ID > 0 {
		return errors.New("用户名已存在")
	}

	// 加密密码
	if err := user.SetPassword(user.Password); err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 创建用户
	return s.userRepository.Create(user)
}

// GetUserProfile 获取用户信息
func (s *authService) GetUserProfile(userID uint) (*models.User, error) {
	return s.userRepository.GetByID(userID)
}
