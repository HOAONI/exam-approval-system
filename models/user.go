package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// 用户角色类型
const (
	RoleStudent = "student" // 学生
	RoleTeacher = "teacher" // 教师
	RoleAdmin   = "admin"   // 管理员
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Username  string    `gorm:"size:50;unique;not null" json:"username"`
	Password  string    `gorm:"size:100;not null" json:"password"` // 允许在API中读取密码
	Name      string    `gorm:"size:50;not null" json:"name"`
	Role      string    `gorm:"size:20;not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// 关联关系 - 将在数据库迁移时创建
	TeacherID uint   `json:"teacher_id"` // 学生关联的教师ID（仅对学生有效）
	Teacher   *User  `gorm:"foreignkey:TeacherID" json:"teacher,omitempty"`
	Students  []User `gorm:"foreignkey:TeacherID" json:"students,omitempty"` // 教师关联的学生（仅对教师有效）
}

// CheckOldPassword 验证旧的加密密码（兼容旧版本）
func (u *User) CheckOldPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) error {
	// 使用bcrypt验证密码
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// SetPassword 设置密码（使用bcrypt加密）
func (u *User) SetPassword(password string) error {
	// 使用bcrypt加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// BeforeCreate 创建记录前的钩子函数
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now())
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// BeforeUpdate 更新记录前的钩子函数
func (u *User) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}
