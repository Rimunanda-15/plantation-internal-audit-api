package models

import "time"

// ============================================================
// SOP CATEGORY
// ============================================================

type SOPCategory struct {
	ID          uint64  `db:"id" json:"id"`
	CompanyID   *uint64 `db:"company_id" json:"company_id"`
	Code        string  `db:"code" json:"code"`
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description"`
	Status      string  `db:"status" json:"status"`
}

type SOPCategoryCreate struct {
	CompanyID   *uint64 `json:"company_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
}

type SOPCategoryUpdate struct {
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

// ============================================================
// SOP DOCUMENT
// ============================================================

type SOPDocument struct {
	ID               uint64     `db:"id" json:"id"`
	CompanyID        uint64     `db:"company_id" json:"company_id"`
	CategoryID       uint64     `db:"category_id" json:"category_id"`
	OrganizationID   *uint64    `db:"organization_id" json:"organization_id"`
	DepartmentID     *uint64    `db:"department_id" json:"department_id"`
	DocumentCode     string     `db:"document_code" json:"document_code"`
	Title            string     `db:"title" json:"title"`
	Purpose          *string    `db:"purpose" json:"purpose"`
	Scope            *string    `db:"scope" json:"scope"`
	OwnerUserID      uint64     `db:"owner_user_id" json:"owner_user_id"`
	CurrentVersionID *uint64    `db:"current_version_id" json:"current_version_id"`
	Status           string     `db:"status" json:"status"`
	EffectiveDate    *time.Time `db:"effective_date" json:"effective_date"`
	ReviewDate       *time.Time `db:"review_date" json:"review_date"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	CreatedBy        uint64     `db:"created_by" json:"created_by"`
	UpdatedAt        *time.Time `db:"updated_at" json:"updated_at"`
	UpdatedBy        *uint64    `db:"updated_by" json:"updated_by"`
	DeletedAt        *time.Time `db:"deleted_at" json:"deleted_at"`
}

type SOPDocumentCreate struct {
	CompanyID      uint64     `json:"company_id"`
	CategoryID     uint64     `json:"category_id"`
	OrganizationID *uint64    `json:"organization_id"`
	DepartmentID   *uint64    `json:"department_id"`
	DocumentCode   string     `json:"document_code"`
	Title          string     `json:"title"`
	Purpose        *string    `json:"purpose"`
	Scope          *string    `json:"scope"`
	OwnerUserID    uint64     `json:"owner_user_id"`
	EffectiveDate  *time.Time `json:"effective_date"`
	ReviewDate     *time.Time `json:"review_date"`
}

type SOPDocumentUpdate struct {
	CategoryID     *uint64    `json:"category_id"`
	OrganizationID *uint64    `json:"organization_id"`
	DepartmentID   *uint64    `json:"department_id"`
	Title          *string    `json:"title"`
	Purpose        *string    `json:"purpose"`
	Scope          *string    `json:"scope"`
	OwnerUserID    *uint64    `json:"owner_user_id"`
	EffectiveDate  *time.Time `json:"effective_date"`
	ReviewDate     *time.Time `json:"review_date"`
}

// ============================================================
// SOP VERSION
// ============================================================

type SOPVersion struct {
	ID             uint64     `db:"id" json:"id"`
	SOPDocumentID  uint64     `db:"sop_document_id" json:"sop_document_id"`
	VersionNumber  uint32     `db:"version_number" json:"version_number"`
	ChangeSummary  *string    `db:"change_summary" json:"change_summary"`
	Content        *string    `db:"content" json:"content"`
	FilePath       *string    `db:"file_path" json:"file_path"`
	FileName       *string    `db:"file_name" json:"file_name"`
	FileSize       *uint64    `db:"file_size" json:"file_size"`
	ChecksumSHA256 *string    `db:"checksum_sha256" json:"checksum_sha256"`
	Status         string     `db:"status" json:"status"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	CreatedBy      uint64     `db:"created_by" json:"created_by"`
	PublishedAt    *time.Time `db:"published_at" json:"published_at"`
	PublishedBy    *uint64    `db:"published_by" json:"published_by"`
}

type SOPVersionCreate struct {
	ChangeSummary *string `json:"change_summary"`
	Content       *string `json:"content"`
}

type SOPVersionUpdate struct {
	ChangeSummary *string `json:"change_summary"`
	Content       *string `json:"content"`
}

// ============================================================
// SOP APPROVAL
// ============================================================

type SOPApproval struct {
	ID             uint64     `db:"id" json:"id"`
	SOPVersionID   uint64     `db:"sop_version_id" json:"sop_version_id"`
	SequenceNumber uint16     `db:"sequence_number" json:"sequence_number"`
	ApproverUserID uint64     `db:"approver_user_id" json:"approver_user_id"`
	Status         string     `db:"status" json:"status"`
	Comments       *string    `db:"comments" json:"comments"`
	ActionAt       *time.Time `db:"action_at" json:"action_at"`
}

type SOPApprovalCreate struct {
	ApproverUserID uint64 `json:"approver_user_id"`
	SequenceNumber uint16 `json:"sequence_number"`
}

type SOPApprovalAction struct {
	Comments *string `json:"comments"`
}

// ============================================================
// SOP ACKNOWLEDGEMENT
// ============================================================

type SOPAcknowledgement struct {
	ID             uint64    `db:"id" json:"id"`
	SOPVersionID   uint64    `db:"sop_version_id" json:"sop_version_id"`
	UserID         uint64    `db:"user_id" json:"user_id"`
	AcknowledgedAt time.Time `db:"acknowledged_at" json:"acknowledged_at"`
}

// ============================================================
// PAGINATION
// ============================================================

type SOPPaginationQuery struct {
	Page           int    `json:"page"`
	Limit          int    `json:"limit"`
	Search         string `json:"search"`
	Status         string `json:"status"`
	CompanyID      uint64 `json:"company_id"`
	CategoryID     uint64 `json:"category_id"`
	OrganizationID uint64 `json:"organization_id"`
	DepartmentID   uint64 `json:"department_id"`
}

type SOPPaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
