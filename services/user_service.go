package services

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	ListUsers() ([]models.User, error)
	ListTeachers() ([]models.User, error)
	ListStudents() ([]models.User, error)
	ListAdmins() ([]models.User, error)
	GetUsersByRole(role string) ([]models.User, error)
	GetAllUsers() ([]models.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	// 为管理员功能添加的新方法
	ListUsersWithFilter(role string, status string) ([]models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	UpdateUserDetails(idStr string, name string, email string, phone string, role string, password string, status string) (*models.User, error)
	DeleteUser(idStr string) error
}

// userService 用户服务实现
type userService struct {
	userRepository repositories.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepository: userRepo,
	}
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepository.GetByID(id)
}

// GetUserByUsername 根据用户名获取用户
func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepository.GetByUsername(username)
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(user *models.User) error {
	return s.userRepository.Update(user)
}

// ListUsers 获取所有用户
func (s *userService) ListUsers() ([]models.User, error) {
	return s.userRepository.List()
}

// ListTeachers 获取所有教师
func (s *userService) ListTeachers() ([]models.User, error) {
	return s.userRepository.ListByRole(models.RoleTeacher)
}

// ListStudents 获取所有学生
func (s *userService) ListStudents() ([]models.User, error) {
	return s.userRepository.ListByRole(models.RoleStudent)
}

// ListAdmins 获取所有管理员
func (s *userService) ListAdmins() ([]models.User, error) {
	return s.userRepository.ListByRole(models.RoleAdmin)
}

// GetUsersByRole 根据角色获取用户列表
func (s *userService) GetUsersByRole(role string) ([]models.User, error) {
	return s.userRepository.ListByRole(role)
}

// GetAllUsers 获取所有用户
func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.userRepository.List()
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := user.CheckPassword(oldPassword); err != nil {
		return errors.New("旧密码不正确")
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 保存用户
	return s.userRepository.Update(user)
}

// ListUsersWithFilter 根据角色和状态获取用户列表
func (s *userService) ListUsersWithFilter(role string, status string) ([]models.User, error) {
	// 如果没有指定角色或状态，返回所有用户
	if role == "all" && (status == "all" || status == "") {
		return s.userRepository.List()
	}

	// 只按角色筛选
	if role != "all" && (status == "all" || status == "") {
		return s.userRepository.ListByRole(role)
	}

	// 需要按状态筛选 (当前实现仅按角色筛选，状态筛选待扩展)
	// 此处应该调用 userRepository 中的方法根据状态筛选
	// 目前先简单实现为只按角色筛选
	if role != "all" {
		return s.userRepository.ListByRole(role)
	}

	// 默认返回所有用户
	return s.userRepository.List()
}

// CreateUser 创建新用户
func (s *userService) CreateUser(user *models.User) (*models.User, error) {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepository.GetByUsername(user.Username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 设置默认值 (如果需要)
	if user.Role == "" {
		user.Role = models.RoleStudent // 默认为学生角色
	}

	// 保存用户
	err := s.userRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserDetails 更新用户详细信息
func (s *userService) UpdateUserDetails(idStr string, name string, email string, phone string, role string, password string, status string) (*models.User, error) {
	// 将字符串ID转换为uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	// 获取用户
	user, err := s.userRepository.GetByID(uint(id))
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 更新信息
	if name != "" {
		user.Name = name
	}
	if role != "" {
		user.Role = role
	}
	if password != "" {
		user.Password = password
	}
	// 注意：此模型中没有Email、Phone和Status字段

	// 保存更新
	err = s.userRepository.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(idStr string) error {
	// 将字符串ID转换为uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 检查用户是否存在
	_, err = s.userRepository.GetByID(uint(id))
	if err != nil {
		return errors.New("用户不存在")
	}

	// 删除用户
	return s.userRepository.Delete(uint(id))
}
