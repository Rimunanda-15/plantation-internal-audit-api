package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"aacs-api/internal/models"
)

type SOPRepository struct {
	DB *sqlx.DB
}

func NewSOPRepository(db *sqlx.DB) *SOPRepository {
	return &SOPRepository{
		DB: db,
	}
}

// ============================================================
// HELPERS
// ============================================================

func normalizeSOPPagination(q models.SOPPaginationQuery) models.SOPPaginationQuery {
	if q.Page < 1 {
		q.Page = 1
	}

	if q.Limit < 1 {
		q.Limit = 10
	}

	if q.Limit > 100 {
		q.Limit = 100
	}

	q.Search = strings.TrimSpace(q.Search)
	q.Status = strings.TrimSpace(q.Status)

	return q
}

func buildSOPPagination(total int64, page, limit int) int {
	if total == 0 || limit < 1 {
		return 0
	}

	pages := int(total) / limit

	if int(total)%limit > 0 {
		pages++
	}

	return pages
}

func scanSOPLastInsertID(result sql.Result) (uint64, error) {
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if id < 1 {
		return 0, errors.New("invalid inserted ID")
	}

	return uint64(id), nil
}

// ============================================================
// CATEGORY
// ============================================================

func (r *SOPRepository) IsCategoryCodeExists(
	ctx context.Context,
	companyID *uint64,
	code string,
	excludeID uint64,
) (bool, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM sop_categories
		WHERE code = ?
		  AND company_id <=> ?
		  AND id <> ?
	`

	err := r.DB.GetContext(
		ctx,
		&count,
		query,
		code,
		companyID,
		excludeID,
	)

	return count > 0, err
}

func (r *SOPRepository) GetCategory(
	ctx context.Context,
	id uint64,
) (*models.SOPCategory, error) {
	var item models.SOPCategory

	query := `
		SELECT
			id,
			company_id,
			code,
			name,
			description,
			status
		FROM sop_categories
		WHERE id = ?
		LIMIT 1
	`

	if err := r.DB.GetContext(ctx, &item, query, id); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *SOPRepository) ListCategories(
	ctx context.Context,
	q models.SOPPaginationQuery,
) ([]models.SOPCategory, int64, error) {
	q = normalizeSOPPagination(q)

	var items []models.SOPCategory
	var total int64

	var where []string
	var args []interface{}

	if q.CompanyID > 0 {
		where = append(where, "(company_id = ? OR company_id IS NULL)")
		args = append(args, q.CompanyID)
	}

	if q.Search != "" {
		where = append(where, "(code LIKE ? OR name LIKE ? OR description LIKE ?)")
		search := "%" + q.Search + "%"
		args = append(args, search, search, search)
	}

	if q.Status != "" {
		where = append(where, "status = ?")
		args = append(args, q.Status)
	}

	whereSQL := ""

	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := `
		SELECT COUNT(*)
		FROM sop_categories
	` + whereSQL

	if err := r.DB.GetContext(
		ctx,
		&total,
		countQuery,
		args...,
	); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit

	listQuery := `
		SELECT
			id,
			company_id,
			code,
			name,
			description,
			status
		FROM sop_categories
	` + whereSQL + `
		ORDER BY name ASC, id DESC
		LIMIT ? OFFSET ?
	`

	listArgs := append(append([]interface{}{}, args...), q.Limit, offset)

	if err := r.DB.SelectContext(
		ctx,
		&items,
		listQuery,
		listArgs...,
	); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *SOPRepository) CreateCategory(
	ctx context.Context,
	input models.SOPCategoryCreate,
) (uint64, error) {
	query := `
		INSERT INTO sop_categories (
			company_id,
			code,
			name,
			description,
			status
		) VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		input.CompanyID,
		input.Code,
		input.Name,
		input.Description,
		input.Status,
	)

	if err != nil {
		return 0, err
	}

	return scanSOPLastInsertID(result)
}

func (r *SOPRepository) UpdateCategory(
	ctx context.Context,
	id uint64,
	input models.SOPCategoryUpdate,
) error {
	var fields []string
	var args []interface{}

	if input.Code != nil {
		fields = append(fields, "code = ?")
		args = append(args, *input.Code)
	}

	if input.Name != nil {
		fields = append(fields, "name = ?")
		args = append(args, *input.Name)
	}

	if input.Description != nil {
		fields = append(fields, "description = ?")
		args = append(args, *input.Description)
	}

	if input.Status != nil {
		fields = append(fields, "status = ?")
		args = append(args, *input.Status)
	}

	if len(fields) == 0 {
		return errors.New("no fields to update")
	}

	query := `
		UPDATE sop_categories
		SET ` + strings.Join(fields, ", ") + `
		WHERE id = ?
	`

	args = append(args, id)

	result, err := r.DB.ExecContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		var count int

		if err := r.DB.GetContext(
			ctx,
			&count,
			"SELECT COUNT(*) FROM sop_categories WHERE id = ?",
			id,
		); err != nil {
			return err
		}

		if count == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}

// ============================================================
// DOCUMENT
// ============================================================

func (r *SOPRepository) IsDocumentCodeExists(
	ctx context.Context,
	companyID uint64,
	code string,
	excludeID uint64,
) (bool, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM sop_documents
		WHERE company_id = ?
		  AND document_code = ?
		  AND deleted_at IS NULL
		  AND id <> ?
	`

	err := r.DB.GetContext(
		ctx,
		&count,
		query,
		companyID,
		code,
		excludeID,
	)

	return count > 0, err
}

func (r *SOPRepository) GetDocument(
	ctx context.Context,
	id uint64,
) (*models.SOPDocument, error) {
	var item models.SOPDocument

	query := `
		SELECT
			id,
			company_id,
			category_id,
			organization_id,
			department_id,
			document_code,
			title,
			purpose,
			scope,
			owner_user_id,
			current_version_id,
			status,
			effective_date,
			review_date,
			created_at,
			created_by,
			updated_at,
			updated_by,
			deleted_at
		FROM sop_documents
		WHERE id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`

	if err := r.DB.GetContext(
		ctx,
		&item,
		query,
		id,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *SOPRepository) ListDocuments(
	ctx context.Context,
	q models.SOPPaginationQuery,
) ([]models.SOPDocument, int64, error) {
	q = normalizeSOPPagination(q)

	var items []models.SOPDocument
	var total int64

	var where []string
	var args []interface{}

	where = append(where, "deleted_at IS NULL")

	if q.CompanyID > 0 {
		where = append(where, "company_id = ?")
		args = append(args, q.CompanyID)
	}

	if q.CategoryID > 0 {
		where = append(where, "category_id = ?")
		args = append(args, q.CategoryID)
	}

	if q.OrganizationID > 0 {
		where = append(where, "organization_id = ?")
		args = append(args, q.OrganizationID)
	}

	if q.DepartmentID > 0 {
		where = append(where, "department_id = ?")
		args = append(args, q.DepartmentID)
	}

	if q.Search != "" {
		where = append(
			where,
			"(document_code LIKE ? OR title LIKE ? OR purpose LIKE ?)",
		)

		search := "%" + q.Search + "%"
		args = append(args, search, search, search)
	}

	if q.Status != "" {
		where = append(where, "status = ?")
		args = append(args, q.Status)
	}

	whereSQL := " WHERE " + strings.Join(where, " AND ")

	countQuery := `
		SELECT COUNT(*)
		FROM sop_documents
	` + whereSQL

	if err := r.DB.GetContext(
		ctx,
		&total,
		countQuery,
		args...,
	); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit

	listQuery := `
		SELECT
			id,
			company_id,
			category_id,
			organization_id,
			department_id,
			document_code,
			title,
			purpose,
			scope,
			owner_user_id,
			current_version_id,
			status,
			effective_date,
			review_date,
			created_at,
			created_by,
			updated_at,
			updated_by,
			deleted_at
		FROM sop_documents
	` + whereSQL + `
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`

	listArgs := append(append([]interface{}{}, args...), q.Limit, offset)

	if err := r.DB.SelectContext(
		ctx,
		&items,
		listQuery,
		listArgs...,
	); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *SOPRepository) CreateDocument(
	ctx context.Context,
	input models.SOPDocumentCreate,
	createdBy uint64,
) (uint64, error) {
	query := `
		INSERT INTO sop_documents (
			company_id,
			category_id,
			organization_id,
			department_id,
			document_code,
			title,
			purpose,
			scope,
			owner_user_id,
			status,
			effective_date,
			review_date,
			created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'draft', ?, ?, ?)
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		input.CompanyID,
		input.CategoryID,
		input.OrganizationID,
		input.DepartmentID,
		input.DocumentCode,
		input.Title,
		input.Purpose,
		input.Scope,
		input.OwnerUserID,
		input.EffectiveDate,
		input.ReviewDate,
		createdBy,
	)

	if err != nil {
		return 0, err
	}

	return scanSOPLastInsertID(result)
}

func (r *SOPRepository) UpdateDocument(
	ctx context.Context,
	id uint64,
	input models.SOPDocumentUpdate,
	updatedBy uint64,
) error {
	var fields []string
	var args []interface{}

	if input.CategoryID != nil {
		fields = append(fields, "category_id = ?")
		args = append(args, *input.CategoryID)
	}

	if input.OrganizationID != nil {
		fields = append(fields, "organization_id = ?")
		args = append(args, *input.OrganizationID)
	}

	if input.DepartmentID != nil {
		fields = append(fields, "department_id = ?")
		args = append(args, *input.DepartmentID)
	}

	if input.Title != nil {
		fields = append(fields, "title = ?")
		args = append(args, *input.Title)
	}

	if input.Purpose != nil {
		fields = append(fields, "purpose = ?")
		args = append(args, *input.Purpose)
	}

	if input.Scope != nil {
		fields = append(fields, "scope = ?")
		args = append(args, *input.Scope)
	}

	if input.OwnerUserID != nil {
		fields = append(fields, "owner_user_id = ?")
		args = append(args, *input.OwnerUserID)
	}

	if input.EffectiveDate != nil {
		fields = append(fields, "effective_date = ?")
		args = append(args, *input.EffectiveDate)
	}

	if input.ReviewDate != nil {
		fields = append(fields, "review_date = ?")
		args = append(args, *input.ReviewDate)
	}

	if len(fields) == 0 {
		return errors.New("no fields to update")
	}

	fields = append(fields, "updated_at = NOW()", "updated_by = ?")
	args = append(args, updatedBy, id)

	query := `
		UPDATE sop_documents
		SET ` + strings.Join(fields, ", ") + `
		WHERE id = ?
		  AND deleted_at IS NULL
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		var count int

		if err := r.DB.GetContext(
			ctx,
			&count,
			"SELECT COUNT(*) FROM sop_documents WHERE id = ? AND deleted_at IS NULL",
			id,
		); err != nil {
			return err
		}

		if count == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}

func (r *SOPRepository) SoftDeleteDocument(
	ctx context.Context,
	id uint64,
	updatedBy uint64,
) error {
	query := `
		UPDATE sop_documents
		SET
			deleted_at = NOW(),
			updated_at = NOW(),
			updated_by = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		updatedBy,
		id,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================================
// DOCUMENT VERSIONS
// ============================================================

func (r *SOPRepository) ListVersions(
	ctx context.Context,
	documentID uint64,
) ([]models.SOPVersion, error) {
	var items []models.SOPVersion

	query := `
		SELECT
			id,
			sop_document_id,
			version_number,
			change_summary,
			content,
			file_path,
			file_name,
			file_size,
			checksum_sha256,
			status,
			created_at,
			created_by,
			published_at,
			published_by
		FROM sop_versions
		WHERE sop_document_id = ?
		ORDER BY version_number DESC
	`

	if err := r.DB.SelectContext(
		ctx,
		&items,
		query,
		documentID,
	); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SOPRepository) GetVersion(
	ctx context.Context,
	versionID uint64,
) (*models.SOPVersion, error) {
	var item models.SOPVersion

	query := `
		SELECT
			id,
			sop_document_id,
			version_number,
			change_summary,
			content,
			file_path,
			file_name,
			file_size,
			checksum_sha256,
			status,
			created_at,
			created_by,
			published_at,
			published_by
		FROM sop_versions
		WHERE id = ?
		LIMIT 1
	`

	if err := r.DB.GetContext(
		ctx,
		&item,
		query,
		versionID,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *SOPRepository) CreateVersion(
	ctx context.Context,
	documentID uint64,
	input models.SOPVersionCreate,
	createdBy uint64,
) (uint64, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	var currentVersion uint32

	err = tx.GetContext(
		ctx,
		&currentVersion,
		`
			SELECT COALESCE(MAX(version_number), 0)
			FROM sop_versions
			WHERE sop_document_id = ?
			FOR UPDATE
		`,
		documentID,
	)

	if err != nil {
		return 0, err
	}

	versionNumber := currentVersion + 1

	result, err := tx.ExecContext(
		ctx,
		`
			INSERT INTO sop_versions (
				sop_document_id,
				version_number,
				change_summary,
				content,
				status,
				created_by
			) VALUES (?, ?, ?, ?, 'draft', ?)
		`,
		documentID,
		versionNumber,
		input.ChangeSummary,
		input.Content,
		createdBy,
	)

	if err != nil {
		return 0, err
	}

	versionID, err := scanSOPLastInsertID(result)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(
		ctx,
		`
			UPDATE sop_documents
			SET
				status = 'draft',
				updated_at = NOW(),
				updated_by = ?
			WHERE id = ?
			  AND deleted_at IS NULL
		`,
		createdBy,
		documentID,
	)

	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return versionID, nil
}

// ============================================================
// APPROVALS
// ============================================================

func (r *SOPRepository) ListApprovals(
	ctx context.Context,
	versionID uint64,
) ([]models.SOPApproval, error) {
	var items []models.SOPApproval

	query := `
		SELECT
			id,
			sop_version_id,
			sequence_number,
			approver_user_id,
			status,
			comments,
			action_at
		FROM sop_approvals
		WHERE sop_version_id = ?
		ORDER BY sequence_number ASC
	`

	if err := r.DB.SelectContext(
		ctx,
		&items,
		query,
		versionID,
	); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SOPRepository) CreateApproval(
	ctx context.Context,
	versionID uint64,
	input models.SOPApprovalCreate,
) (uint64, error) {
	query := `
		INSERT INTO sop_approvals (
			sop_version_id,
			sequence_number,
			approver_user_id,
			status
		) VALUES (?, ?, ?, 'pending')
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		versionID,
		input.SequenceNumber,
		input.ApproverUserID,
	)

	if err != nil {
		return 0, err
	}

	return scanSOPLastInsertID(result)
}

func (r *SOPRepository) UpdateApprovalStatus(
	ctx context.Context,
	approvalID uint64,
	status string,
	comments *string,
) error {
	query := `
		UPDATE sop_approvals
		SET
			status = ?,
			comments = ?,
			action_at = NOW()
		WHERE id = ?
		  AND status = 'pending'
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		status,
		comments,
		approvalID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================================
// ACKNOWLEDGEMENTS
// ============================================================

func (r *SOPRepository) ListAcknowledgements(
	ctx context.Context,
	versionID uint64,
) ([]models.SOPAcknowledgement, error) {
	var items []models.SOPAcknowledgement

	query := `
		SELECT
			id,
			sop_version_id,
			user_id,
			acknowledged_at
		FROM sop_acknowledgements
		WHERE sop_version_id = ?
		ORDER BY acknowledged_at DESC
	`

	if err := r.DB.SelectContext(
		ctx,
		&items,
		query,
		versionID,
	); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SOPRepository) CreateAcknowledgement(
	ctx context.Context,
	versionID uint64,
	userID uint64,
) (uint64, error) {
	query := `
		INSERT INTO sop_acknowledgements (
			sop_version_id,
			user_id
		) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			acknowledged_at = acknowledged_at
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		versionID,
		userID,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if id == 0 {
		var existingID uint64

		err := r.DB.GetContext(
			ctx,
			&existingID,
			`
				SELECT id
				FROM sop_acknowledgements
				WHERE sop_version_id = ?
				  AND user_id = ?
			`,
			versionID,
			userID,
		)

		if err != nil {
			return 0, err
		}

		return existingID, nil
	}

	return uint64(id), nil
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

func (r *SOPRepository) CategoryBelongsToCompany(
	ctx context.Context,
	categoryID uint64,
	companyID uint64,
) (bool, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM sop_categories
		WHERE id = ?
		  AND (company_id = ? OR company_id IS NULL)
		  AND status = 'active'
	`

	err := r.DB.GetContext(
		ctx,
		&count,
		query,
		categoryID,
		companyID,
	)

	return count > 0, err
}

func (r *SOPRepository) DocumentExists(
	ctx context.Context,
	documentID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(
		ctx,
		&count,
		`
			SELECT COUNT(*)
			FROM sop_documents
			WHERE id = ?
			  AND deleted_at IS NULL
		`,
		documentID,
	)

	return count > 0, err
}

func (r *SOPRepository) VersionBelongsToDocument(
	ctx context.Context,
	versionID uint64,
	documentID uint64,
) (bool, error) {
	var count int

	err := r.DB.GetContext(
		ctx,
		&count,
		`
			SELECT COUNT(*)
			FROM sop_versions
			WHERE id = ?
			  AND sop_document_id = ?
		`,
		versionID,
		documentID,
	)

	return count > 0, err
}

func (r *SOPRepository) GetNextVersionNumber(
	ctx context.Context,
	documentID uint64,
) (uint32, error) {
	var version uint32

	err := r.DB.GetContext(
		ctx,
		&version,
		`
			SELECT COALESCE(MAX(version_number), 0) + 1
			FROM sop_versions
			WHERE sop_document_id = ?
		`,
		documentID,
	)

	return version, err
}

// ============================================================
// PAGINATION RESPONSE HELPER
// ============================================================

func BuildSOPPaginatedResponse(
	data interface{},
	total int64,
	page int,
	limit int,
) models.SOPPaginatedResponse {
	return models.SOPPaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: buildSOPPagination(total, page, limit),
	}
}

// ============================================================
// ERROR HELPERS
// ============================================================

func IsSOPNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func SOPRepositoryError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("SOP repository %s: %w", operation, err)
}
