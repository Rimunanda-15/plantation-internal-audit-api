package models

import "time"

type Company struct {
	ID        uint64     `db:"id" json:"id"`
	Code      string     `db:"code" json:"code"`
	Name      string     `db:"name" json:"name"`
	Address   *string    `db:"address" json:"address"`
	Phone     *string    `db:"phone" json:"phone"`
	Email     *string    `db:"email" json:"email"`
	Status    string     `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}

type CompanyCreate struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
	Email   *string `json:"email"`
	Status  string  `json:"status"`
}

type CompanyUpdate struct {
	Code    *string `json:"code"`
	Name    *string `json:"name"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
	Email   *string `json:"email"`
	Status  *string `json:"status"`
}

type Organization struct {
	ID        uint64     `db:"id" json:"id"`
	CompanyID uint64     `db:"company_id" json:"company_id"`
	ParentID  *uint64    `db:"parent_id" json:"parent_id"`
	Code      string     `db:"code" json:"code"`
	Name      string     `db:"name" json:"name"`
	Type      string     `db:"type" json:"type"`
	Status    string     `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}

type OrganizationCreate struct {
	CompanyID uint64  `json:"company_id"`
	ParentID  *uint64 `json:"parent_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Status    string  `json:"status"`
}

type OrganizationUpdate struct {
	ParentID *uint64 `json:"parent_id"`
	Code     *string `json:"code"`
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Status   *string `json:"status"`
}

type Department struct {
	ID             uint64    `db:"id" json:"id"`
	CompanyID      uint64    `db:"company_id" json:"company_id"`
	OrganizationID *uint64   `db:"organization_id" json:"organization_id"`
	Code           string    `db:"code" json:"code"`
	Name           string    `db:"name" json:"name"`
	Status         string    `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type DepartmentCreate struct {
	CompanyID      uint64  `json:"company_id"`
	OrganizationID *uint64 `json:"organization_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
}

type DepartmentUpdate struct {
	OrganizationID *uint64 `json:"organization_id"`
	Code           *string `json:"code"`
	Name           *string `json:"name"`
	Status         *string `json:"status"`
}

type EstateBlock struct {
	ID           uint64     `db:"id" json:"id"`
	CompanyID    uint64     `db:"company_id" json:"company_id"`
	EstateID     uint64     `db:"estate_id" json:"estate_id"`
	Code         string     `db:"code" json:"code"`
	Name         string     `db:"name" json:"name"`
	Afdeling     *string    `db:"afdeling" json:"afdeling"`
	AreaHectare  *float64   `db:"area_hectare" json:"area_hectare"`
	PlantingYear *uint16    `db:"planting_year" json:"planting_year"`
	PlantType    *string    `db:"plant_type" json:"plant_type"`
	Status       string     `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at"`
}

type EstateBlockCreate struct {
	CompanyID    uint64   `json:"company_id"`
	EstateID     uint64   `json:"estate_id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Afdeling     *string  `json:"afdeling"`
	AreaHectare  *float64 `json:"area_hectare"`
	PlantingYear *uint16  `json:"planting_year"`
	PlantType    *string  `json:"plant_type"`
	Status       string   `json:"status"`
}

type EstateBlockUpdate struct {
	Code         *string  `json:"code"`
	Name         *string  `json:"name"`
	Afdeling     *string  `json:"afdeling"`
	AreaHectare  *float64 `json:"area_hectare"`
	PlantingYear *uint16  `json:"planting_year"`
	PlantType    *string  `json:"plant_type"`
	Status       *string  `json:"status"`
}

type PaginationQuery struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
	Status string `json:"status"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int64       `json:"total_pages"`
}
