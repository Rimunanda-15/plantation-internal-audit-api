package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"aacs-api/internal/models"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type OrganizationRepository struct {
	DB *sqlx.DB
}

func NewOrganizationRepository(db *sqlx.DB) *OrganizationRepository {
	return &OrganizationRepository{
		DB: db,
	}
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	return limit, (page - 1) * limit
}

func buildTotalPages(total int64, limit int) int64 {
	if total == 0 || limit <= 0 {
		return 0
	}

	return (total + int64(limit) - 1) / int64(limit)
}

func validateStatusFilter(status string, allowed ...string) error {
	if status == "" {
		return nil
	}

	for _, value := range allowed {
		if status == value {
			return nil
		}
	}

	return fmt.Errorf("invalid status filter: %s", status)
}

func (r *OrganizationRepository) CompanyExists(
	ctx context.Context,
	id uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM companies
		WHERE id = ?
	`, id)

	return count > 0, err
}

func (r *OrganizationRepository) OrganizationExists(
	ctx context.Context,
	id uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM organizations
		WHERE id = ?
	`, id)

	return count > 0, err
}

func (r *OrganizationRepository) OrganizationBelongsToCompany(
	ctx context.Context,
	organizationID uint64,
	companyID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM organizations
		WHERE id = ?
		  AND company_id = ?
	`, organizationID, companyID)

	return count > 0, err
}

func (r *OrganizationRepository) IsOrganizationParentValid(
	ctx context.Context,
	parentID uint64,
	companyID uint64,
	currentID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM organizations
		WHERE id = ?
		  AND company_id = ?
		  AND id <> ?
	`, parentID, companyID, currentID)

	return count > 0, err
}

func (r *OrganizationRepository) DepartmentBelongsToCompany(
	ctx context.Context,
	departmentID uint64,
	companyID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM departments
		WHERE id = ?
		  AND company_id = ?
	`, departmentID, companyID)

	return count > 0, err
}

func (r *OrganizationRepository) EstateBelongsToCompany(
	ctx context.Context,
	estateID uint64,
	companyID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM organizations
		WHERE id = ?
		  AND company_id = ?
		  AND type = 'estate'
	`, estateID, companyID)

	return count > 0, err
}

func (r *OrganizationRepository) IsCompanyCodeExists(
	ctx context.Context,
	code string,
	excludeID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM companies
		WHERE code = ?
		  AND id <> ?
	`, code, excludeID)

	return count > 0, err
}

func (r *OrganizationRepository) IsOrganizationCodeExists(
	ctx context.Context,
	companyID uint64,
	code string,
	excludeID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM organizations
		WHERE company_id = ?
		  AND code = ?
		  AND id <> ?
	`, companyID, code, excludeID)

	return count > 0, err
}

func (r *OrganizationRepository) IsDepartmentCodeExists(
	ctx context.Context,
	companyID uint64,
	code string,
	excludeID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM departments
		WHERE company_id = ?
		  AND code = ?
		  AND id <> ?
	`, companyID, code, excludeID)

	return count > 0, err
}

func (r *OrganizationRepository) IsEstateBlockCodeExists(
	ctx context.Context,
	estateID uint64,
	code string,
	excludeID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM estate_blocks
		WHERE estate_id = ?
		  AND code = ?
		  AND id <> ?
	`, estateID, code, excludeID)

	return count > 0, err
}

// ============================================================
// COMPANY
// ============================================================

func (r *OrganizationRepository) ListCompanies(
	ctx context.Context,
	q models.PaginationQuery,
) ([]models.Company, int64, error) {
	limit, offset := normalizePagination(q.Page, q.Limit)

	where := []string{"1 = 1"}
	args := []interface{}{}

	if q.Search != "" {
		where = append(where, "(code LIKE ? OR name LIKE ?)")
		search := "%" + strings.TrimSpace(q.Search) + "%"
		args = append(args, search, search)
	}

	if q.Status != "" {
		if err := validateStatusFilter(q.Status, "active", "inactive"); err != nil {
			return nil, 0, err
		}

		where = append(where, "status = ?")
		args = append(args, q.Status)
	}

	condition := strings.Join(where, " AND ")

	var total int64

	countQuery := `
		SELECT COUNT(*)
		FROM companies
		WHERE ` + condition

	if err := r.DB.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id,
			code,
			name,
			address,
			phone,
			email,
			status,
			created_at,
			updated_at
		FROM companies
		WHERE ` + condition + `
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`

	queryArgs := append(append([]interface{}{}, args...), limit, offset)

	var data []models.Company

	if err := r.DB.SelectContext(ctx, &data, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (r *OrganizationRepository) GetCompany(
	ctx context.Context,
	id uint64,
) (*models.Company, error) {
	var data models.Company

	err := r.DB.GetContext(ctx, &data, `
		SELECT
			id,
			code,
			name,
			address,
			phone,
			email,
			status,
			created_at,
			updated_at
		FROM companies
		WHERE id = ?
	`, id)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *OrganizationRepository) CreateCompany(
	ctx context.Context,
	input models.CompanyCreate,
) (uint64, error) {
	result, err := r.DB.ExecContext(ctx, `
		INSERT INTO companies (
			code,
			name,
			address,
			phone,
			email,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		input.Code,
		input.Name,
		input.Address,
		input.Phone,
		input.Email,
		input.Status,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (r *OrganizationRepository) UpdateCompany(
	ctx context.Context,
	id uint64,
	input models.CompanyUpdate,
) error {
	sets := []string{}
	args := []interface{}{}

	if input.Code != nil {
		sets = append(sets, "code = ?")
		args = append(args, *input.Code)
	}

	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}

	if input.Address != nil {
		sets = append(sets, "address = ?")
		args = append(args, *input.Address)
	}

	if input.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, *input.Phone)
	}

	if input.Email != nil {
		sets = append(sets, "email = ?")
		args = append(args, *input.Email)
	}

	if input.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *input.Status)
	}

	if len(sets) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := `
		UPDATE companies
		SET ` + strings.Join(sets, ", ") + `
		WHERE id = ?
	`

	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

// ============================================================
// ORGANIZATION
// ============================================================

func (r *OrganizationRepository) ListOrganizations(
	ctx context.Context,
	q models.PaginationQuery,
	companyID uint64,
) ([]models.Organization, int64, error) {
	limit, offset := normalizePagination(q.Page, q.Limit)

	where := []string{"1 = 1"}
	args := []interface{}{}

	if companyID > 0 {
		where = append(where, "company_id = ?")
		args = append(args, companyID)
	}

	if q.Search != "" {
		where = append(where, "(code LIKE ? OR name LIKE ?)")
		search := "%" + strings.TrimSpace(q.Search) + "%"
		args = append(args, search, search)
	}

	if q.Status != "" {
		if err := validateStatusFilter(q.Status, "active", "inactive"); err != nil {
			return nil, 0, err
		}

		where = append(where, "status = ?")
		args = append(args, q.Status)
	}

	condition := strings.Join(where, " AND ")

	var total int64

	if err := r.DB.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM organizations
		WHERE `+condition, args...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id,
			company_id,
			parent_id,
			code,
			name,
			type,
			status,
			created_at,
			updated_at
		FROM organizations
		WHERE ` + condition + `
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`

	queryArgs := append(append([]interface{}{}, args...), limit, offset)

	var data []models.Organization

	if err := r.DB.SelectContext(ctx, &data, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (r *OrganizationRepository) GetOrganization(
	ctx context.Context,
	id uint64,
) (*models.Organization, error) {
	var data models.Organization

	err := r.DB.GetContext(ctx, &data, `
		SELECT
			id,
			company_id,
			parent_id,
			code,
			name,
			type,
			status,
			created_at,
			updated_at
		FROM organizations
		WHERE id = ?
	`, id)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *OrganizationRepository) CreateOrganization(
	ctx context.Context,
	input models.OrganizationCreate,
) (uint64, error) {
	result, err := r.DB.ExecContext(ctx, `
		INSERT INTO organizations (
			company_id,
			parent_id,
			code,
			name,
			type,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		input.CompanyID,
		input.ParentID,
		input.Code,
		input.Name,
		input.Type,
		input.Status,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (r *OrganizationRepository) UpdateOrganization(
	ctx context.Context,
	id uint64,
	input models.OrganizationUpdate,
) error {
	sets := []string{}
	args := []interface{}{}

	if input.ParentID != nil {
		sets = append(sets, "parent_id = ?")
		args = append(args, *input.ParentID)
	}

	if input.Code != nil {
		sets = append(sets, "code = ?")
		args = append(args, *input.Code)
	}

	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}

	if input.Type != nil {
		sets = append(sets, "type = ?")
		args = append(args, *input.Type)
	}

	if input.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *input.Status)
	}

	if len(sets) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := `
		UPDATE organizations
		SET ` + strings.Join(sets, ", ") + `
		WHERE id = ?
	`

	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

// ============================================================
// DEPARTMENT
// ============================================================

func (r *OrganizationRepository) ListDepartments(
	ctx context.Context,
	q models.PaginationQuery,
	companyID uint64,
) ([]models.Department, int64, error) {
	limit, offset := normalizePagination(q.Page, q.Limit)

	where := []string{"1 = 1"}
	args := []interface{}{}

	if companyID > 0 {
		where = append(where, "company_id = ?")
		args = append(args, companyID)
	}

	if q.Search != "" {
		where = append(where, "(code LIKE ? OR name LIKE ?)")
		search := "%" + strings.TrimSpace(q.Search) + "%"
		args = append(args, search, search)
	}

	if q.Status != "" {
		if err := validateStatusFilter(q.Status, "active", "inactive"); err != nil {
			return nil, 0, err
		}

		where = append(where, "status = ?")
		args = append(args, q.Status)
	}

	condition := strings.Join(where, " AND ")

	var total int64

	if err := r.DB.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM departments
		WHERE `+condition, args...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id,
			company_id,
			organization_id,
			code,
			name,
			status,
			created_at
		FROM departments
		WHERE ` + condition + `
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`

	queryArgs := append(append([]interface{}{}, args...), limit, offset)

	var data []models.Department

	if err := r.DB.SelectContext(ctx, &data, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (r *OrganizationRepository) GetDepartment(
	ctx context.Context,
	id uint64,
) (*models.Department, error) {
	var data models.Department

	err := r.DB.GetContext(ctx, &data, `
		SELECT
			id,
			company_id,
			organization_id,
			code,
			name,
			status,
			created_at
		FROM departments
		WHERE id = ?
	`, id)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *OrganizationRepository) CreateDepartment(
	ctx context.Context,
	input models.DepartmentCreate,
) (uint64, error) {
	result, err := r.DB.ExecContext(ctx, `
		INSERT INTO departments (
			company_id,
			organization_id,
			code,
			name,
			status
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		input.CompanyID,
		input.OrganizationID,
		input.Code,
		input.Name,
		input.Status,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (r *OrganizationRepository) UpdateDepartment(
	ctx context.Context,
	id uint64,
	input models.DepartmentUpdate,
) error {
	sets := []string{}
	args := []interface{}{}

	if input.OrganizationID != nil {
		sets = append(sets, "organization_id = ?")
		args = append(args, *input.OrganizationID)
	}

	if input.Code != nil {
		sets = append(sets, "code = ?")
		args = append(args, *input.Code)
	}

	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}

	if input.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *input.Status)
	}

	if len(sets) == 0 {
		return fmt.Errorf("no fields to update")
	}

	args = append(args, id)

	query := `
		UPDATE departments
		SET ` + strings.Join(sets, ", ") + `
		WHERE id = ?
	`

	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

// ============================================================
// ESTATE BLOCK
// ============================================================

func (r *OrganizationRepository) ListEstateBlocks(
	ctx context.Context,
	q models.PaginationQuery,
	companyID uint64,
	estateID uint64,
) ([]models.EstateBlock, int64, error) {
	limit, offset := normalizePagination(q.Page, q.Limit)

	where := []string{"1 = 1"}
	args := []interface{}{}

	if companyID > 0 {
		where = append(where, "company_id = ?")
		args = append(args, companyID)
	}

	if estateID > 0 {
		where = append(where, "estate_id = ?")
		args = append(args, estateID)
	}

	if q.Search != "" {
		where = append(where, "(code LIKE ? OR name LIKE ? OR afdeling LIKE ?)")
		search := "%" + strings.TrimSpace(q.Search) + "%"
		args = append(args, search, search, search)
	}

	if q.Status != "" {
		if err := validateStatusFilter(q.Status, "active", "inactive", "replanting"); err != nil {
			return nil, 0, err
		}

		where = append(where, "status = ?")
		args = append(args, q.Status)
	}

	condition := strings.Join(where, " AND ")

	var total int64

	if err := r.DB.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM estate_blocks
		WHERE `+condition, args...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id,
			company_id,
			estate_id,
			code,
			name,
			afdeling,
			area_hectare,
			planting_year,
			plant_type,
			status,
			created_at,
			updated_at
		FROM estate_blocks
		WHERE ` + condition + `
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`

	queryArgs := append(append([]interface{}{}, args...), limit, offset)

	var data []models.EstateBlock

	if err := r.DB.SelectContext(ctx, &data, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (r *OrganizationRepository) GetEstateBlock(
	ctx context.Context,
	id uint64,
) (*models.EstateBlock, error) {
	var data models.EstateBlock

	err := r.DB.GetContext(ctx, &data, `
		SELECT
			id,
			company_id,
			estate_id,
			code,
			name,
			afdeling,
			area_hectare,
			planting_year,
			plant_type,
			status,
			created_at,
			updated_at
		FROM estate_blocks
		WHERE id = ?
	`, id)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *OrganizationRepository) CreateEstateBlock(
	ctx context.Context,
	input models.EstateBlockCreate,
) (uint64, error) {
	result, err := r.DB.ExecContext(ctx, `
		INSERT INTO estate_blocks (
			company_id,
			estate_id,
			code,
			name,
			afdeling,
			area_hectare,
			planting_year,
			plant_type,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		input.CompanyID,
		input.EstateID,
		input.Code,
		input.Name,
		input.Afdeling,
		input.AreaHectare,
		input.PlantingYear,
		input.PlantType,
		input.Status,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (r *OrganizationRepository) UpdateEstateBlock(
	ctx context.Context,
	id uint64,
	input models.EstateBlockUpdate,
) error {
	sets := []string{}
	args := []interface{}{}

	if input.Code != nil {
		sets = append(sets, "code = ?")
		args = append(args, *input.Code)
	}

	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}

	if input.Afdeling != nil {
		sets = append(sets, "afdeling = ?")
		args = append(args, *input.Afdeling)
	}

	if input.AreaHectare != nil {
		sets = append(sets, "area_hectare = ?")
		args = append(args, *input.AreaHectare)
	}

	if input.PlantingYear != nil {
		sets = append(sets, "planting_year = ?")
		args = append(args, *input.PlantingYear)
	}

	if input.PlantType != nil {
		sets = append(sets, "plant_type = ?")
		args = append(args, *input.PlantType)
	}

	if input.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *input.Status)
	}

	if len(sets) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := `
		UPDATE estate_blocks
		SET ` + strings.Join(sets, ", ") + `
		WHERE id = ?
	`

	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

// ============================================================
// TOTAL PAGES HELPER
// ============================================================

func GetTotalPages(total int64, limit int) int64 {
	return buildTotalPages(total, limit)
}
