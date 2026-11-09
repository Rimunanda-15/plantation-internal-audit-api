package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"aacs-api/internal/models"
	"aacs-api/internal/repositories"
)

type SOPService struct {
	Repo *repositories.SOPRepository
}

func NewSOPService(repo *repositories.SOPRepository) *SOPService {
	return &SOPService{
		Repo: repo,
	}
}

var (
	ErrSOPInvalidInput     = errors.New("input SOP tidak valid")
	ErrSOPNotFound         = errors.New("data SOP tidak ditemukan")
	ErrSOPDuplicateCode    = errors.New("kode sudah digunakan")
	ErrSOPInvalidStatus    = errors.New("status tidak valid")
	ErrSOPCategoryMismatch = errors.New("kategori tidak sesuai dengan perusahaan")
	ErrSOPApprovalAction   = errors.New("aksi approval tidak valid")
)

func sopNotFound(err error) error {
	if repositories.IsSOPNotFound(err) {
		return ErrSOPNotFound
	}

	return err
}

// ============================================================
// CATEGORY
// ============================================================

func (s *SOPService) ListCategories(
	ctx context.Context,
	query models.SOPPaginationQuery,
) ([]models.SOPCategory, int64, error) {
	return s.Repo.ListCategories(ctx, query)
}

func (s *SOPService) GetCategory(
	ctx context.Context,
	id uint64,
) (*models.SOPCategory, error) {
	if id == 0 {
		return nil, ErrSOPInvalidInput
	}

	category, err := s.Repo.GetCategory(ctx, id)
	if err != nil {
		return nil, sopNotFound(err)
	}

	return category, nil
}

func (s *SOPService) CreateCategory(
	ctx context.Context,
	input models.SOPCategoryCreate,
) (uint64, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)

	if input.Code == "" || input.Name == "" {
		return 0, fmt.Errorf("%w: code dan name wajib diisi", ErrSOPInvalidInput)
	}

	if input.Status == "" {
		input.Status = "active"
	}

	if !isValidCategoryStatus(input.Status) {
		return 0, ErrSOPInvalidStatus
	}

	exists, err := s.Repo.IsCategoryCodeExists(
		ctx,
		input.CompanyID,
		input.Code,
		0,
	)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrSOPDuplicateCode
	}

	return s.Repo.CreateCategory(ctx, input)
}

func (s *SOPService) UpdateCategory(
	ctx context.Context,
	id uint64,
	input models.SOPCategoryUpdate,
) error {
	if id == 0 {
		return ErrSOPInvalidInput
	}

	category, err := s.Repo.GetCategory(ctx, id)
	if err != nil {
		return sopNotFound(err)
	}

	if input.Code != nil {
		code := strings.TrimSpace(*input.Code)
		if code == "" {
			return fmt.Errorf("%w: code tidak boleh kosong", ErrSOPInvalidInput)
		}

		exists, err := s.Repo.IsCategoryCodeExists(
			ctx,
			category.CompanyID,
			code,
			id,
		)
		if err != nil {
			return err
		}

		if exists {
			return ErrSOPDuplicateCode
		}

		input.Code = &code
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return fmt.Errorf("%w: name tidak boleh kosong", ErrSOPInvalidInput)
		}

		input.Name = &name
	}

	if input.Status != nil && !isValidCategoryStatus(*input.Status) {
		return ErrSOPInvalidStatus
	}

	return s.Repo.UpdateCategory(ctx, id, input)
}

// ============================================================
// DOCUMENT
// ============================================================

func (s *SOPService) ListDocuments(
	ctx context.Context,
	query models.SOPPaginationQuery,
) ([]models.SOPDocument, int64, error) {
	return s.Repo.ListDocuments(ctx, query)
}

func (s *SOPService) GetDocument(
	ctx context.Context,
	id uint64,
) (*models.SOPDocument, error) {
	if id == 0 {
		return nil, ErrSOPInvalidInput
	}

	document, err := s.Repo.GetDocument(ctx, id)
	if err != nil {
		return nil, sopNotFound(err)
	}

	return document, nil
}

func (s *SOPService) CreateDocument(
	ctx context.Context,
	input models.SOPDocumentCreate,
	createdBy uint64,
) (uint64, error) {
	input.DocumentCode = strings.TrimSpace(input.DocumentCode)
	input.Title = strings.TrimSpace(input.Title)

	if input.CompanyID == 0 ||
		input.CategoryID == 0 ||
		input.OwnerUserID == 0 ||
		input.DocumentCode == "" ||
		input.Title == "" ||
		createdBy == 0 {
		return 0, ErrSOPInvalidInput
	}

	categoryValid, err := s.Repo.CategoryBelongsToCompany(
		ctx,
		input.CategoryID,
		input.CompanyID,
	)
	if err != nil {
		return 0, err
	}

	if !categoryValid {
		return 0, ErrSOPCategoryMismatch
	}

	exists, err := s.Repo.IsDocumentCodeExists(
		ctx,
		input.CompanyID,
		input.DocumentCode,
		0,
	)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrSOPDuplicateCode
	}

	return s.Repo.CreateDocument(ctx, input, createdBy)
}

func (s *SOPService) UpdateDocument(
	ctx context.Context,
	id uint64,
	input models.SOPDocumentUpdate,
	updatedBy uint64,
) error {
	if id == 0 || updatedBy == 0 {
		return ErrSOPInvalidInput
	}

	document, err := s.Repo.GetDocument(ctx, id)
	if err != nil {
		return sopNotFound(err)
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return fmt.Errorf("%w: title tidak boleh kosong", ErrSOPInvalidInput)
		}

		input.Title = &title
	}

	if input.CategoryID != nil {
		categoryValid, err := s.Repo.CategoryBelongsToCompany(
			ctx,
			*input.CategoryID,
			document.CompanyID,
		)
		if err != nil {
			return err
		}

		if !categoryValid {
			return ErrSOPCategoryMismatch
		}
	}

	return s.Repo.UpdateDocument(ctx, id, input, updatedBy)
}

func (s *SOPService) DeleteDocument(
	ctx context.Context,
	id uint64,
	deletedBy uint64,
) error {
	if id == 0 || deletedBy == 0 {
		return ErrSOPInvalidInput
	}

	exists, err := s.Repo.DocumentExists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return ErrSOPNotFound
	}

	return s.Repo.SoftDeleteDocument(ctx, id, deletedBy)
}

// ============================================================
// VERSION
// ============================================================

func (s *SOPService) ListVersions(
	ctx context.Context,
	documentID uint64,
) ([]models.SOPVersion, error) {
	if documentID == 0 {
		return nil, ErrSOPInvalidInput
	}

	exists, err := s.Repo.DocumentExists(ctx, documentID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrSOPNotFound
	}

	return s.Repo.ListVersions(ctx, documentID)
}

func (s *SOPService) GetVersion(
	ctx context.Context,
	versionID uint64,
) (*models.SOPVersion, error) {
	if versionID == 0 {
		return nil, ErrSOPInvalidInput
	}

	version, err := s.Repo.GetVersion(ctx, versionID)
	if err != nil {
		return nil, sopNotFound(err)
	}

	return version, nil
}

func (s *SOPService) CreateVersion(
	ctx context.Context,
	documentID uint64,
	input models.SOPVersionCreate,
	createdBy uint64,
) (uint64, error) {
	if documentID == 0 || createdBy == 0 {
		return 0, ErrSOPInvalidInput
	}

	exists, err := s.Repo.DocumentExists(ctx, documentID)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, ErrSOPNotFound
	}

	if input.ChangeSummary != nil {
		summary := strings.TrimSpace(*input.ChangeSummary)
		input.ChangeSummary = &summary
	}

	return s.Repo.CreateVersion(ctx, documentID, input, createdBy)
}

// ============================================================
// APPROVAL
// ============================================================

func (s *SOPService) ListApprovals(
	ctx context.Context,
	versionID uint64,
) ([]models.SOPApproval, error) {
	if versionID == 0 {
		return nil, ErrSOPInvalidInput
	}

	_, err := s.Repo.GetVersion(ctx, versionID)
	if err != nil {
		return nil, sopNotFound(err)
	}

	return s.Repo.ListApprovals(ctx, versionID)
}

func (s *SOPService) CreateApproval(
	ctx context.Context,
	versionID uint64,
	input models.SOPApprovalCreate,
) (uint64, error) {
	if versionID == 0 ||
		input.ApproverUserID == 0 ||
		input.SequenceNumber == 0 {
		return 0, ErrSOPInvalidInput
	}

	_, err := s.Repo.GetVersion(ctx, versionID)
	if err != nil {
		return 0, sopNotFound(err)
	}

	return s.Repo.CreateApproval(ctx, versionID, input)
}

func (s *SOPService) ActOnApproval(
	ctx context.Context,
	approvalID uint64,
	action string,
	input models.SOPApprovalAction,
) error {
	if approvalID == 0 {
		return ErrSOPInvalidInput
	}

	action = strings.ToLower(strings.TrimSpace(action))

	switch action {
	case "approved", "rejected", "returned":
	default:
		return ErrSOPApprovalAction
	}

	return s.Repo.UpdateApprovalStatus(
		ctx,
		approvalID,
		action,
		input.Comments,
	)
}

// ============================================================
// ACKNOWLEDGEMENT
// ============================================================

func (s *SOPService) ListAcknowledgements(
	ctx context.Context,
	versionID uint64,
) ([]models.SOPAcknowledgement, error) {
	if versionID == 0 {
		return nil, ErrSOPInvalidInput
	}

	_, err := s.Repo.GetVersion(ctx, versionID)
	if err != nil {
		return nil, sopNotFound(err)
	}

	return s.Repo.ListAcknowledgements(ctx, versionID)
}

func (s *SOPService) Acknowledge(
	ctx context.Context,
	versionID uint64,
	userID uint64,
) (uint64, error) {
	if versionID == 0 || userID == 0 {
		return 0, ErrSOPInvalidInput
	}

	_, err := s.Repo.GetVersion(ctx, versionID)
	if err != nil {
		return 0, sopNotFound(err)
	}

	return s.Repo.CreateAcknowledgement(ctx, versionID, userID)
}

// ============================================================
// VALIDATION
// ============================================================

func isValidCategoryStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "inactive":
		return true
	default:
		return false
	}
}
