package models

import "time"

type User struct {
	ID          uint64     `db:"id" json:"id"`
	Username    string     `db:"username" json:"username"`
	Email       *string    `db:"email" json:"email"`
	FullName    string     `db:"full_name" json:"full_name"`
	Phone       *string    `db:"phone" json:"phone"`
	Status      string     `db:"status" json:"status"`
	LastLoginAt *time.Time `db:"last_login_at" json:"last_login_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CreatedBy   *uint64    `db:"created_by" json:"created_by"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`
	UpdatedBy   *uint64    `db:"updated_by" json:"updated_by"`
	RoleCode    string     `db:"role_code" json:"role_code"`
	RoleName    string     `db:"role_name" json:"role_name"`
	CompanyID   *uint64    `db:"company_id" json:"company_id"`
}

type UserCreate struct {
	Username  string  `json:"username"`
	Email     *string `json:"email"`
	FullName  string  `json:"full_name"`
	Phone     *string `json:"phone"`
	Password  string  `json:"password"`
	RoleCode  string  `json:"role_code"`
	CompanyID *uint64 `json:"company_id"`
}

type UserUpdate struct {
	FullName *string `json:"full_name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Status   *string `json:"status"`
}
