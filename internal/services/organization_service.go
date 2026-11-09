package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"aacs-api/internal/models"
	"aacs-api/internal/repositories"
)

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrDuplicateCode        = errors.New("code already exists")
	ErrCompanyNotFound      = errors.New("company not found")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrEstateBlockNotFound  = errors.New("estate block not found")
	ErrInvalidRelation      = errors.New("invalid master data relation")
	ErrNoFieldsToUpdate     = errors.New("no fields to update")
)

type OrganizationService struct {
	Repository *repositories.OrganizationRepository
}

func NewOrganizationService(
	repository *repositories.OrganizationRepository,
) *OrganizationService {
	return &OrganizationService{
		Repository: repository,
	}
}

func normalizePaginationQuery(q models.PaginationQuery) models.PaginationQuery {
	if q.Page < 1 {
		q.Page = 1
	}

	if q.Limit < 1 {
		q.Limit = 20
	}

	if q.Limit > 100 {
		q.Limit = 100
	}

	q.Search = strings.TrimSpace(q.Search)
	q.Status = strings.ToLower(strings.TrimSpace(q.Status))

	return q
}

func validateRequired(value, field string, maxLength int) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidInput, field)
	}

	if len(value) > maxLength {
		return fmt.Errorf(
			"%w: %s maximum length is %d characters",
			ErrInvalidInput,
			field,
			maxLength,
		)
	}

	return nil
}

func normalizeStatus(status string, allowed ...string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	if status == "" {
		status = "active"
	}

	for _, item := range allowed {
		if status == item {
			return status, nil
		}
	}

	return "", fmt.Errorf("%w: invalid status", ErrInvalidInput)
}

func validateEmail(email *string) error {
	if email == nil {
		return nil
	}

	value := strings.TrimSpace(*email)

	if value == "" {
		return nil
	}

	if len(value) > 150 {
		return fmt.Errorf("%w: email maximum length is 150 characters", ErrInvalidInput)
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("%w: invalid email address", ErrInvalidInput)
	}

	*email = value

	return nil
}

func validatePhone(phone *string) error {
	if phone == nil {
		return nil
	}

	value := strings.TrimSpace(*phone)

	if len(value) > 30 {
		return fmt.Errorf("%w: phone maximum length is 30 characters", ErrInvalidInput)
	}

	*phone = value

	return nil
}

func validateOptionalString(value *string, field string, maxLength int) error {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)

	if len(trimmed) > maxLength {
		return fmt.Errorf(
			"%w: %s maximum length is %d characters",
			ErrInvalidInput,
			field,
			maxLength,
		)
	}

	*value = trimmed

	return nil
}

func validateOrganizationType(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	allowed := map[string]bool{
		"head_office": true,
		"estate":      true,
		"mill":        true,
		"department":  true,
		"other":       true,
	}

	if !allowed[value] {
		return "", fmt.Errorf("%w: invalid organization type", ErrInvalidInput)
	}

	return value, nil
}

func validateBlockStatus(value string) (string, error) {
	return normalizeStatus(value, "active", "inactive", "replanting")
}

func validatePositiveID(id uint64, field string) error {
	if id == 0 {
		return fmt.Errorf("%w: %s is required", ErrInvalidInput, field)
	}

	return nil
}

func (s *OrganizationService) checkCompany(
	ctx context.Context,
	companyID uint64,
) error {
	exists, err := s.Repository.CompanyExists(ctx, companyID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrCompanyNotFound
	}

	return nil
}

// ============================================================
// COMPANY
// ============================================================

func (s *OrganizationService) ListCompanies(
	ctx context.Context,
	q models.PaginationQuery,
) ([]models.Company, int64, error) {
	q = normalizePaginationQuery(q)

	data, total, err := s.Repository.ListCompanies(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (s *OrganizationService) GetCompany(
	ctx context.Context,
	id uint64,
) (*models.Company, error) {
	if err := validatePositiveID(id, "company_id"); err != nil {
		return nil, err
	}

	data, err := s.Repository.GetCompany(ctx, id)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *OrganizationService) CreateCompany(
	ctx context.Context,
	input models.CompanyCreate,
) (uint64, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)

	if err := validateRequired(input.Code, "code", 30); err != nil {
		return 0, err
	}

	if err := validateRequired(input.Name, "name", 150); err != nil {
		return 0, err
	}

	if err := validateEmail(input.Email); err != nil {
		return 0, err
	}

	if err := validatePhone(input.Phone); err != nil {
		return 0, err
	}

	status, err := normalizeStatus(input.Status, "active", "inactive")
	if err != nil {
		return 0, err
	}
	input.Status = status

	exists, err := s.Repository.IsCompanyCodeExists(ctx, input.Code, 0)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrDuplicateCode
	}

	return s.Repository.CreateCompany(ctx, input)
}

func (s *OrganizationService) UpdateCompany(
	ctx context.Context,
	id uint64,
	input models.CompanyUpdate,
) error {
	if err := validatePositiveID(id, "company_id"); err != nil {
		return err
	}

	if _, err := s.Repository.GetCompany(ctx, id); err != nil {
		return err
	}

	hasChanges := false

	if input.Code != nil {
		*input.Code = strings.ToUpper(strings.TrimSpace(*input.Code))

		if err := validateRequired(*input.Code, "code", 30); err != nil {
			return err
		}

		exists, err := s.Repository.IsCompanyCodeExists(ctx, *input.Code, id)
		if err != nil {
			return err
		}

		if exists {
			return ErrDuplicateCode
		}

		hasChanges = true
	}

	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)

		if err := validateRequired(*input.Name, "name", 150); err != nil {
			return err
		}

		hasChanges = true
	}

	if input.Address != nil {
		if err := validateOptionalString(input.Address, "address", 65535); err != nil {
			return err
		}
		hasChanges = true
	}

	if input.Phone != nil {
		if err := validatePhone(input.Phone); err != nil {
			return err
		}
		hasChanges = true
	}

	if input.Email != nil {
		if err := validateEmail(input.Email); err != nil {
			return err
		}
		hasChanges = true
	}

	if input.Status != nil {
		status, err := normalizeStatus(*input.Status, "active", "inactive")
		if err != nil {
			return err
		}

		*input.Status = status
		hasChanges = true
	}

	if !hasChanges {
		return ErrNoFieldsToUpdate
	}

	return s.Repository.UpdateCompany(ctx, id, input)
}

// ============================================================
// ORGANIZATION
// ============================================================

func (s *OrganizationService) ListOrganizations(
	ctx context.Context,
	q models.PaginationQuery,
	companyID uint64,
) ([]models.Organization, int64, error) {
	q = normalizePaginationQuery(q)

	if companyID > 0 {
		if err := s.checkCompany(ctx, companyID); err != nil {
			return nil, 0, err
		}
	}

	data, total, err := s.Repository.ListOrganizations(ctx, q, companyID)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (s *OrganizationService) GetOrganization(
	ctx context.Context,
	id uint64,
) (*models.Organization, error) {
	if err := validatePositiveID(id, "organization_id"); err != nil {
		return nil, err
	}

	return s.Repository.GetOrganization(ctx, id)
}

func (s *OrganizationService) CreateOrganization(
	ctx context.Context,
	input models.OrganizationCreate,
) (uint64, error) {
	if err := validatePositiveID(input.CompanyID, "company_id"); err != nil {
		return 0, err
	}

	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)

	if err := validateRequired(input.Code, "code", 40); err != nil {
		return 0, err
	}

	if err := validateRequired(input.Name, "name", 150); err != nil {
		return 0, err
	}

	orgType, err := validateOrganizationType(input.Type)
	if err != nil {
		return 0, err
	}
	input.Type = orgType

	status, err := normalizeStatus(input.Status, "active", "inactive")
	if err != nil {
		return 0, err
	}
	input.Status = status

	if err := s.checkCompany(ctx, input.CompanyID); err != nil {
		return 0, err
	}

	if input.ParentID != nil {
		if *input.ParentID == 0 {
			return 0, fmt.Errorf("%w: invalid parent_id", ErrInvalidInput)
		}

		valid, err := s.Repository.IsOrganizationParentValid(
			ctx,
			*input.ParentID,
			input.CompanyID,
			0,
		)
		if err != nil {
			return 0, err
		}

		if !valid {
			return 0, ErrInvalidRelation
		}
	}

	exists, err := s.Repository.IsOrganizationCodeExists(
		ctx,
		input.CompanyID,
		input.Code,
		0,
	)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrDuplicateCode
	}

	return s.Repository.CreateOrganization(ctx, input)
}

func (s *OrganizationService) UpdateOrganization(
	ctx context.Context,
	id uint64,
	input models.OrganizationUpdate,
) error {
	if err := validatePositiveID(id, "organization_id"); err != nil {
		return err
	}

	current, err := s.Repository.GetOrganization(ctx, id)
	if err != nil {
		return err
	}

	hasChanges := false

	if input.ParentID != nil {
		if *input.ParentID == id {
			return ErrInvalidRelation
		}

		if *input.ParentID == 0 {
			return fmt.Errorf("%w: invalid parent_id", ErrInvalidInput)
		}

		valid, err := s.Repository.IsOrganizationParentValid(
			ctx,
			*input.ParentID,
			current.CompanyID,
			id,
		)
		if err != nil {
			return err
		}

		if !valid {
			return ErrInvalidRelation
		}

		hasChanges = true
	}

	if input.Code != nil {
		*input.Code = strings.ToUpper(strings.TrimSpace(*input.Code))

		if err := validateRequired(*input.Code, "code", 40); err != nil {
			return err
		}

		exists, err := s.Repository.IsOrganizationCodeExists(
			ctx,
			current.CompanyID,
			*input.Code,
			id,
		)
		if err != nil {
			return err
		}

		if exists {
			return ErrDuplicateCode
		}

		hasChanges = true
	}

	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)

		if err := validateRequired(*input.Name, "name", 150); err != nil {
			return err
		}

		hasChanges = true
	}

	if input.Type != nil {
		value, err := validateOrganizationType(*input.Type)
		if err != nil {
			return err
		}

		*input.Type = value
		hasChanges = true
	}

	if input.Status != nil {
		value, err := normalizeStatus(*input.Status, "active", "inactive")
		if err != nil {
			return err
		}

		*input.Status = value
		hasChanges = true
	}

	if !hasChanges {
		return ErrNoFieldsToUpdate
	}

	return s.Repository.UpdateOrganization(ctx, id, input)
}

// ============================================================
// DEPARTMENT
// ============================================================

func (s *OrganizationService) ListDepartments(
	ctx context.Context,
	q models.PaginationQuery,
	companyID uint64,
) ([]models.Department, int64, error) {
	q = normalizePaginationQuery(q)

	if companyID > 0 {
		if err := s.checkCompany(ctx, companyID); err != nil {
			return nil, 0, err
		}
	}

	data, total, err := s.Repository.ListDepartments(ctx, q, companyID)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (s *OrganizationService) GetDepartment(
	ctx context.Context,
	id uint64,
) (*models.Department, error) {
	if err := validatePositiveID(id, "department_id"); err != nil {
		return nil, err
	}

	return s.Repository.GetDepartment(ctx, id)
}

func (s *OrganizationService) CreateDepartment(
	ctx context.Context,
	input models.DepartmentCreate,
) (uint64, error) {
	if err := validatePositiveID(input.CompanyID, "company_id"); err != nil {
		return 0, err
	}

	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)

	if err := validateRequired(input.Code, "code", 40); err != nil {
		return 0, err
	}

	if err := validateRequired(input.Name, "name", 150); err != nil {
		return 0, err
	}

	status, err := normalizeStatus(input.Status, "active", "inactive")
	if err != nil {
		return 0, err
	}
	input.Status = status

	if err := s.checkCompany(ctx, input.CompanyID); err != nil {
		return 0, err
	}

	if input.OrganizationID != nil {
		valid, err := s.Repository.OrganizationBelongsToCompany(
			ctx,
			*input.OrganizationID,
			input.CompanyID,
		)
		if err != nil {
			return 0, err
		}

		if !valid {
			return 0, ErrInvalidRelation
		}
	}

	exists, err := s.Repository.IsDepartmentCodeExists(
		ctx,
		input.CompanyID,
		input.Code,
		0,
	)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrDuplicateCode
	}

	return s.Repository.CreateDepartment(ctx, input)
}

func (s *OrganizationService) UpdateDepartment(
	ctx context.Context,
	id uint64,
	input models.DepartmentUpdate,
) error {
	if err := validatePositiveID(id, "department_id"); err != nil {
		return err
	}

	current, err := s.Repository.GetDepartment(ctx, id)
	if err != nil {
		return err
	}

	hasChanges := false

	if input.OrganizationID != nil {
		valid, err := s.Repository.OrganizationBelongsToCompany(
			ctx,
			*input.OrganizationID,
			current.CompanyID,
		)
		if err != nil {
			return err
		}

		if !valid {
			return ErrInvalidRelation
		}

		hasChanges = true
	}

	if input.Code != nil {
		*input.Code = strings.ToUpper(strings.TrimSpace(*input.Code))

		if err := validateRequired(*input.Code, "code", 40); err != nil {
			return err
		}

		exists, err := s.Repository.IsDepartmentCodeExists(
			ctx,
			current.CompanyID,
			*input.Code,
			id,
		)
		if err != nil {
			return err
		}

		if exists {
			return ErrDuplicateCode
		}

		hasChanges = true
	}

	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)

		if err := validateRequired(*input.Name, "name", 150); err != nil {
			return err
		}

		hasChanges = true
	}

	if input.Status != nil {
		value, err := normalizeStatus(*input.Status, "active", "inactive")
		if err != nil {
			return err
		}

		*input.Status = value
		hasChanges = true
	}

	if !hasChanges {
		return ErrNoFieldsToUpdate
	}

	return s.Repository.UpdateDepartment(ctx, id, input)
}

// ============================================================
// ESTATE BLOCK
// ============================================================

func (s *OrganizationService) ListEstateBlocks(
	ctx context.Context,
	q models.PaginationQuery,
	companyID uint64,
	estateID uint64,
) ([]models.EstateBlock, int64, error) {
	q = normalizePaginationQuery(q)

	if companyID > 0 {
		if err := s.checkCompany(ctx, companyID); err != nil {
			return nil, 0, err
		}
	}

	if estateID > 0 {
		if companyID == 0 {
			return nil, 0, fmt.Errorf("%w: company_id is required with estate_id", ErrInvalidInput)
		}

		valid, err := s.Repository.EstateBelongsToCompany(
			ctx,
			estateID,
			companyID,
		)
		if err != nil {
			return nil, 0, err
		}

		if !valid {
			return nil, 0, ErrInvalidRelation
		}
	}

	data, total, err := s.Repository.ListEstateBlocks(
		ctx,
		q,
		companyID,
		estateID,
	)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (s *OrganizationService) GetEstateBlock(
	ctx context.Context,
	id uint64,
) (*models.EstateBlock, error) {
	if err := validatePositiveID(id, "estate_block_id"); err != nil {
		return nil, err
	}

	return s.Repository.GetEstateBlock(ctx, id)
}

func (s *OrganizationService) CreateEstateBlock(
	ctx context.Context,
	input models.EstateBlockCreate,
) (uint64, error) {
	if err := validatePositiveID(input.CompanyID, "company_id"); err != nil {
		return 0, err
	}

	if err := validatePositiveID(input.EstateID, "estate_id"); err != nil {
		return 0, err
	}

	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)

	if err := validateRequired(input.Code, "code", 50); err != nil {
		return 0, err
	}

	if err := validateRequired(input.Name, "name", 150); err != nil {
		return 0, err
	}

	if err := validateOptionalString(input.Afdeling, "afdeling", 100); err != nil {
		return 0, err
	}

	if err := validateOptionalString(input.PlantType, "plant_type", 80); err != nil {
		return 0, err
	}

	status, err := validateBlockStatus(input.Status)
	if err != nil {
		return 0, err
	}
	input.Status = status

	if input.AreaHectare != nil && *input.AreaHectare < 0 {
		return 0, fmt.Errorf("%w: area_hectare cannot be negative", ErrInvalidInput)
	}

	if input.PlantingYear != nil {
		currentYear := uint16(time.Now().Year())

		if *input.PlantingYear < 1900 || *input.PlantingYear > currentYear+1 {
			return 0, fmt.Errorf("%w: invalid planting_year", ErrInvalidInput)
		}
	}

	valid, err := s.Repository.EstateBelongsToCompany(
		ctx,
		input.EstateID,
		input.CompanyID,
	)
	if err != nil {
		return 0, err
	}

	if !valid {
		return 0, ErrInvalidRelation
	}

	exists, err := s.Repository.IsEstateBlockCodeExists(
		ctx,
		input.EstateID,
		input.Code,
		0,
	)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrDuplicateCode
	}

	return s.Repository.CreateEstateBlock(ctx, input)
}

func (s *OrganizationService) UpdateEstateBlock(
	ctx context.Context,
	id uint64,
	input models.EstateBlockUpdate,
) error {
	if err := validatePositiveID(id, "estate_block_id"); err != nil {
		return err
	}

	current, err := s.Repository.GetEstateBlock(ctx, id)
	if err != nil {
		return err
	}

	hasChanges := false

	if input.Code != nil {
		*input.Code = strings.ToUpper(strings.TrimSpace(*input.Code))

		if err := validateRequired(*input.Code, "code", 50); err != nil {
			return err
		}

		exists, err := s.Repository.IsEstateBlockCodeExists(
			ctx,
			current.EstateID,
			*input.Code,
			id,
		)
		if err != nil {
			return err
		}

		if exists {
			return ErrDuplicateCode
		}

		hasChanges = true
	}

	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)

		if err := validateRequired(*input.Name, "name", 150); err != nil {
			return err
		}

		hasChanges = true
	}

	if input.Afdeling != nil {
		if err := validateOptionalString(input.Afdeling, "afdeling", 100); err != nil {
			return err
		}
		hasChanges = true
	}

	if input.AreaHectare != nil {
		if *input.AreaHectare < 0 {
			return fmt.Errorf("%w: area_hectare cannot be negative", ErrInvalidInput)
		}
		hasChanges = true
	}

	if input.PlantingYear != nil {
		currentYear := uint16(time.Now().Year())

		if *input.PlantingYear < 1900 || *input.PlantingYear > currentYear+1 {
			return fmt.Errorf("%w: invalid planting_year", ErrInvalidInput)
		}
		hasChanges = true
	}

	if input.PlantType != nil {
		if err := validateOptionalString(input.PlantType, "plant_type", 80); err != nil {
			return err
		}
		hasChanges = true
	}

	if input.Status != nil {
		value, err := validateBlockStatus(*input.Status)
		if err != nil {
			return err
		}

		*input.Status = value
		hasChanges = true
	}

	if !hasChanges {
		return ErrNoFieldsToUpdate
	}

	return s.Repository.UpdateEstateBlock(ctx, id, input)
}
