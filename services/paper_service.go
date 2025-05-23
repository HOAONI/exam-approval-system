package services

import (
	"errors"
	"time"

	"github.com/exam-approval-system/models"
	"github.com/exam-approval-system/repositories"
	"github.com/exam-approval-system/utils"
)

// PaperService 试卷服务接口
type PaperService interface {
	CreatePaper(paper *models.Paper) error
	GetPaperByID(id uint) (*models.Paper, error)
	GetPapersByExamID(examID uint) ([]models.Paper, error)
	UpdatePaper(paper *models.Paper) error
	DeletePaper(id uint) error
	SignPaper(paperID uint, signerID uint) error
	VerifyPaperSignature(paperID uint) (bool, error)
}

// paperService 试卷服务实现
type paperService struct {
	paperRepository repositories.PaperRepository
	examRepository  repositories.ExamRepository
}

// NewPaperService 创建试卷服务
func NewPaperService(paperRepo repositories.PaperRepository, examRepo repositories.ExamRepository) PaperService {
	return &paperService{
		paperRepository: paperRepo,
		examRepository:  examRepo,
	}
}

// CreatePaper 创建试卷
func (s *paperService) CreatePaper(paper *models.Paper) error {
	// 检查关联的考试是否存在以及状态是否为草稿
	exam, err := s.examRepository.GetByID(paper.ExamID)
	if err != nil {
		return errors.New("考试不存在")
	}
	if exam.Status != models.StatusDraft && exam.Status != models.StatusRejected {
		return errors.New("只能为草稿或被拒绝状态的考试创建试卷")
	}

	// 设置试卷状态为草稿
	paper.Status = models.StatusDraft
	return s.paperRepository.Create(paper)
}

// GetPaperByID 根据ID获取试卷
func (s *paperService) GetPaperByID(id uint) (*models.Paper, error) {
	return s.paperRepository.GetByID(id)
}

// GetPapersByExamID 根据考试ID获取试卷
func (s *paperService) GetPapersByExamID(examID uint) ([]models.Paper, error) {
	return s.paperRepository.GetByExamID(examID)
}

// UpdatePaper 更新试卷
func (s *paperService) UpdatePaper(paper *models.Paper) error {
	// 检查关联的考试状态
	exam, err := s.examRepository.GetByID(paper.ExamID)
	if err != nil {
		return errors.New("考试不存在")
	}
	if exam.Status != models.StatusDraft && exam.Status != models.StatusRejected {
		return errors.New("只能修改草稿或被拒绝状态的考试相关试卷")
	}

	return s.paperRepository.Update(paper)
}

// DeletePaper 删除试卷
func (s *paperService) DeletePaper(id uint) error {
	paper, err := s.paperRepository.GetByID(id)
	if err != nil {
		return err
	}

	// 检查关联的考试状态
	exam, err := s.examRepository.GetByID(paper.ExamID)
	if err != nil {
		return errors.New("考试不存在")
	}
	if exam.Status != models.StatusDraft && exam.Status != models.StatusRejected {
		return errors.New("只能删除草稿或被拒绝状态的考试相关试卷")
	}

	return s.paperRepository.Delete(id)
}

// SignPaper 为试卷签名
func (s *paperService) SignPaper(paperID uint, signerID uint) error {
	// 获取试卷信息
	paper, err := s.paperRepository.GetByID(paperID)
	if err != nil {
		return errors.New("试卷不存在")
	}

	// 检查试卷状态
	if paper.Status != models.StatusApproved {
		return errors.New("只有已审批的试卷才能签名")
	}

	// 生成签名
	now := time.Now()
	signature := utils.GeneratePaperSignature(
		paper.ID,
		paper.Title,
		paper.Content,
		paper.Questions,
		now,
	)

	// 更新试卷签名信息
	paper.Signature = signature
	paper.SignedAt = now
	paper.SignedBy = signerID

	// 保存更新
	return s.paperRepository.Update(paper)
}

// VerifyPaperSignature 验证试卷签名
func (s *paperService) VerifyPaperSignature(paperID uint) (bool, error) {
	// 获取试卷信息
	paper, err := s.paperRepository.GetByID(paperID)
	if err != nil {
		return false, errors.New("试卷不存在")
	}

	// 检查是否有签名
	if paper.Signature == "" {
		return false, errors.New("试卷未签名")
	}

	// 验证签名
	isValid := utils.VerifyPaperSignature(
		paper.ID,
		paper.Title,
		paper.Content,
		paper.Questions,
		paper.SignedAt,
		paper.Signature,
	)

	return isValid, nil
}
