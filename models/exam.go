package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 试卷状态常量
const (
	StatusDraft     = "draft"     // 草稿
	StatusPending   = "pending"   // 待审批
	StatusApproved  = "approved"  // 已审批
	StatusRejected  = "rejected"  // 已拒绝
	StatusPublished = "published" // 已发布
)

// Exam 考试模型 - 试卷数据表
type Exam struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	Description string    `gorm:"size:1000" json:"description"`
	Course      string    `gorm:"size:100;not null" json:"course"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatorID   uint      `json:"creator_id"`
	Creator     User      `gorm:"foreignkey:CreatorID" json:"creator"`
	Status      string    `gorm:"size:20;not null;default:'draft'" json:"status"`
	ApproverID  uint      `json:"approver_id"`
	Approver    User      `gorm:"foreignkey:ApproverID" json:"approver"`
	Papers      []Paper   `gorm:"foreignkey:ExamID" json:"papers"`
	TotalScore  float64   `json:"total_score"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExamData 试卷数据表 - 用于专门存储试卷数据
type ExamData struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	ExamID     uint      `json:"exam_id"`
	Exam       Exam      `gorm:"foreignkey:ExamID" json:"exam"`
	StudentID  uint      `json:"student_id"`
	Student    User      `gorm:"foreignkey:StudentID" json:"student"`
	Title      string    `gorm:"size:100;not null" json:"title"`
	Course     string    `gorm:"size:100;not null" json:"course"`
	TotalScore float64   `json:"total_score"`
	Status     string    `gorm:"size:20;not null;default:'draft'" json:"status"`
	ApproverID uint      `json:"approver_id"`
	Approver   User      `gorm:"foreignkey:ApproverID" json:"approver"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Paper 试卷模型
type Paper struct {
	ID           uint      `gorm:"primary_key" json:"id"`
	ExamID       uint      `json:"exam_id"`
	Title        string    `gorm:"size:100;not null" json:"title"`
	Content      string    `gorm:"type:text" json:"content"`
	Questions    string    `gorm:"type:text" json:"questions"` // JSON格式存储题目
	Duration     int       `json:"duration"`                   // 考试时长（分钟）
	TotalScore   float64   `json:"total_score"`
	PassingScore float64   `json:"passing_score"`
	Status       string    `gorm:"size:20;not null;default:'draft'" json:"status"`
	Signature    string    `gorm:"size:256" json:"signature"` // 试卷签名
	SignedAt     time.Time `json:"signed_at"`                 // 签名时间
	SignedBy     uint      `json:"signed_by"`                 // 签名人ID
	Signer       User      `gorm:"foreignkey:SignedBy" json:"signer"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Comment 审批评论
type Comment struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	ExamID    uint      `json:"exam_id"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignkey:UserID" json:"user"`
	Content   string    `gorm:"size:1000;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate 创建记录前的钩子函数
func (e *Exam) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now())
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// BeforeUpdate 更新记录前的钩子函数
func (e *Exam) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// BeforeCreate 创建记录前的钩子函数
func (p *Paper) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now())
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// BeforeUpdate 更新记录前的钩子函数
func (p *Paper) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// BeforeCreate 创建记录前的钩子函数
func (c *Comment) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now())
	return nil
}

// BeforeCreate 创建记录前的钩子函数
func (ed *ExamData) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now())
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// BeforeUpdate 更新记录前的钩子函数
func (ed *ExamData) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}
