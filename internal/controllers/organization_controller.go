package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/models"
	"aacs-api/internal/services"
)

type OrganizationController struct {
	Service *services.OrganizationService
}

func NewOrganizationController(
	service *services.OrganizationService,
) *OrganizationController {
	return &OrganizationController{
		Service: service,
	}
}

func successResponse(c fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": message,
		"data":    data,
	})
}

func errorResponse(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "ERROR",
		"message": message,
	})
}

func parseID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid ID")
	}

	return id, nil
}

func parseOptionalID(value string) (uint64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}

	return parseID(value)
}

func getPaginationQuery(c fiber.Ctx) models.PaginationQuery {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	return models.PaginationQuery{
		Page:   page,
		Limit:  limit,
		Search: strings.TrimSpace(c.Query("search")),
		Status: strings.TrimSpace(c.Query("status")),
	}
}

func paginationResponse(
	data interface{},
	total int64,
	page int,
	limit int,
) models.PaginatedResponse {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	totalPages := int64(0)

	if total > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	return models.PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

func handleOrganizationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errorResponse(c, http.StatusNotFound, "Data tidak ditemukan")

	case errors.Is(err, services.ErrCompanyNotFound):
		return errorResponse(c, http.StatusNotFound, "Company tidak ditemukan")

	case errors.Is(err, services.ErrOrganizationNotFound):
		return errorResponse(c, http.StatusNotFound, "Organization tidak ditemukan")

	case errors.Is(err, services.ErrDepartmentNotFound):
		return errorResponse(c, http.StatusNotFound, "Department tidak ditemukan")

	case errors.Is(err, services.ErrEstateBlockNotFound):
		return errorResponse(c, http.StatusNotFound, "Estate block tidak ditemukan")

	case errors.Is(err, services.ErrDuplicateCode):
		return errorResponse(c, http.StatusConflict, "Kode sudah digunakan")

	case errors.Is(err, services.ErrInvalidRelation):
		return errorResponse(c, http.StatusBadRequest, "Relasi data tidak valid")

	case errors.Is(err, services.ErrInvalidInput):
		return errorResponse(c, http.StatusBadRequest, err.Error())

	case errors.Is(err, services.ErrNoFieldsToUpdate):
		return errorResponse(c, http.StatusBadRequest, "Tidak ada data yang diperbarui")

	default:
		return errorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
}

// ============================================================
// COMPANY
// ============================================================

func (ctrl *OrganizationController) ListCompanies(c fiber.Ctx) error {
	query := getPaginationQuery(c)

	data, total, err := ctrl.Service.ListCompanies(c.Context(), query)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	response := paginationResponse(
		data,
		total,
		query.Page,
		query.Limit,
	)

	return successResponse(
		c,
		http.StatusOK,
		"Daftar company berhasil diambil",
		response,
	)
}

func (ctrl *OrganizationController) DetailCompany(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Company ID tidak valid")
	}

	data, err := ctrl.Service.GetCompany(c.Context(), id)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Detail company berhasil diambil",
		data,
	)
}

func (ctrl *OrganizationController) CreateCompany(c fiber.Ctx) error {
	var input models.CompanyCreate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	id, err := ctrl.Service.CreateCompany(c.Context(), input)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusCreated,
		"Company berhasil dibuat",
		fiber.Map{
			"id": id,
		},
	)
}

func (ctrl *OrganizationController) UpdateCompany(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Company ID tidak valid")
	}

	var input models.CompanyUpdate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := ctrl.Service.UpdateCompany(c.Context(), id, input); err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Company berhasil diperbarui",
		nil,
	)
}

// ============================================================
// ORGANIZATION
// ============================================================

func (ctrl *OrganizationController) ListOrganizations(c fiber.Ctx) error {
	query := getPaginationQuery(c)

	companyID, err := parseOptionalID(c.Query("company_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Company ID tidak valid")
	}

	data, total, err := ctrl.Service.ListOrganizations(
		c.Context(),
		query,
		companyID,
	)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	response := paginationResponse(
		data,
		total,
		query.Page,
		query.Limit,
	)

	return successResponse(
		c,
		http.StatusOK,
		"Daftar organization berhasil diambil",
		response,
	)
}

func (ctrl *OrganizationController) DetailOrganization(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Organization ID tidak valid")
	}

	data, err := ctrl.Service.GetOrganization(c.Context(), id)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Detail organization berhasil diambil",
		data,
	)
}

func (ctrl *OrganizationController) CreateOrganization(c fiber.Ctx) error {
	var input models.OrganizationCreate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	id, err := ctrl.Service.CreateOrganization(c.Context(), input)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusCreated,
		"Organization berhasil dibuat",
		fiber.Map{
			"id": id,
		},
	)
}

func (ctrl *OrganizationController) UpdateOrganization(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Organization ID tidak valid")
	}

	var input models.OrganizationUpdate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := ctrl.Service.UpdateOrganization(c.Context(), id, input); err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Organization berhasil diperbarui",
		nil,
	)
}

// ============================================================
// DEPARTMENT
// ============================================================

func (ctrl *OrganizationController) ListDepartments(c fiber.Ctx) error {
	query := getPaginationQuery(c)

	companyID, err := parseOptionalID(c.Query("company_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Company ID tidak valid")
	}

	data, total, err := ctrl.Service.ListDepartments(
		c.Context(),
		query,
		companyID,
	)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	response := paginationResponse(
		data,
		total,
		query.Page,
		query.Limit,
	)

	return successResponse(
		c,
		http.StatusOK,
		"Daftar department berhasil diambil",
		response,
	)
}

func (ctrl *OrganizationController) DetailDepartment(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Department ID tidak valid")
	}

	data, err := ctrl.Service.GetDepartment(c.Context(), id)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Detail department berhasil diambil",
		data,
	)
}

func (ctrl *OrganizationController) CreateDepartment(c fiber.Ctx) error {
	var input models.DepartmentCreate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	id, err := ctrl.Service.CreateDepartment(c.Context(), input)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusCreated,
		"Department berhasil dibuat",
		fiber.Map{
			"id": id,
		},
	)
}

func (ctrl *OrganizationController) UpdateDepartment(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Department ID tidak valid")
	}

	var input models.DepartmentUpdate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := ctrl.Service.UpdateDepartment(c.Context(), id, input); err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Department berhasil diperbarui",
		nil,
	)
}

// ============================================================
// ESTATE BLOCK
// ============================================================

func (ctrl *OrganizationController) ListEstateBlocks(c fiber.Ctx) error {
	query := getPaginationQuery(c)

	companyID, err := parseOptionalID(c.Query("company_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Company ID tidak valid")
	}

	estateID, err := parseOptionalID(c.Query("estate_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Estate ID tidak valid")
	}

	data, total, err := ctrl.Service.ListEstateBlocks(
		c.Context(),
		query,
		companyID,
		estateID,
	)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	response := paginationResponse(
		data,
		total,
		query.Page,
		query.Limit,
	)

	return successResponse(
		c,
		http.StatusOK,
		"Daftar estate block berhasil diambil",
		response,
	)
}

func (ctrl *OrganizationController) DetailEstateBlock(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Estate block ID tidak valid")
	}

	data, err := ctrl.Service.GetEstateBlock(c.Context(), id)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Detail estate block berhasil diambil",
		data,
	)
}

func (ctrl *OrganizationController) CreateEstateBlock(c fiber.Ctx) error {
	var input models.EstateBlockCreate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	id, err := ctrl.Service.CreateEstateBlock(c.Context(), input)
	if err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusCreated,
		"Estate block berhasil dibuat",
		fiber.Map{
			"id": id,
		},
	)
}

func (ctrl *OrganizationController) UpdateEstateBlock(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Estate block ID tidak valid")
	}

	var input models.EstateBlockUpdate

	if err := c.Bind().Body(&input); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := ctrl.Service.UpdateEstateBlock(c.Context(), id, input); err != nil {
		return handleOrganizationError(c, err)
	}

	return successResponse(
		c,
		http.StatusOK,
		"Estate block berhasil diperbarui",
		nil,
	)
}
