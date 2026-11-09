package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"aacs-api/internal/models"
)

type AuthRepository struct {
	DB *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{
		DB: db,
	}
}

type AuthUser struct {
	models.User
	PasswordHash        string     `db:"password_hash"`
	FailedLoginAttempts uint8      `db:"failed_login_attempts"`
	LockedUntil         *time.Time `db:"locked_until"`
}

func (r *AuthRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*AuthUser, error) {
	var user AuthUser

	err := r.DB.GetContext(ctx, &user, `
		SELECT
			id,
			username,
			email,
			password_hash,
			full_name,
			phone,
			status,
			last_login_at,
			created_at,
			created_by,
			updated_at,
			updated_by,
			failed_login_attempts,
			locked_until
		FROM users
		WHERE username = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, username)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) CountUsers(
	ctx context.Context,
) (int64, error) {
	var count int64

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM users
		WHERE deleted_at IS NULL
	`)

	return count, err
}

func (r *AuthRepository) GetRoleID(
	ctx context.Context,
	roleCode string,
) (uint64, error) {
	var roleID uint64

	err := r.DB.GetContext(ctx, &roleID, `
		SELECT id
		FROM roles
		WHERE code = ?
		LIMIT 1
	`, roleCode)

	return roleID, err
}

func (r *AuthRepository) CreateUser(
	ctx context.Context,
	input models.UserCreate,
	passwordHash string,
	createdBy *uint64,
	isBootstrap bool,
) (uint64, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	roleCode := input.RoleCode

	if isBootstrap {
		roleCode = "SUPER_ADMIN"
	} else {
		if !isAllowedRole(roleCode) || roleCode == "SUPER_ADMIN" {
			return 0, errors.New("invalid role assignment")
		}
	}

	var roleID uint64

	err = tx.GetContext(ctx, &roleID, `
		SELECT id
		FROM roles
		WHERE code = ?
		LIMIT 1
	`, roleCode)

	if err != nil {
		return 0, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (
			username,
			email,
			password_hash,
			full_name,
			phone,
			status,
			created_by
		) VALUES (?, ?, ?, ?, ?, 'active', ?)
	`,
		input.Username,
		input.Email,
		passwordHash,
		input.FullName,
		input.Phone,
		createdBy,
	)

	if err != nil {
		return 0, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	userID := uint64(lastInsertID)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_roles (
			user_id,
			role_id,
			company_id
		) VALUES (?, ?, ?)
	`, userID, roleID, input.CompanyID)

	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO activity_logs (
			user_id,
			module,
			action,
			entity_type,
			entity_id
		) VALUES (?, 'users', 'create', 'user', ?)
	`, createdBy, userID)

	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

func isAllowedRole(roleCode string) bool {
	switch roleCode {
	case "SUPER_ADMIN",
		"IT_AUDIT_MANAGER",
		"INTERNAL_AUDITOR",
		"MANAGEMENT":
		return true
	default:
		return false
	}
}

func (r *AuthRepository) SaveSession(
	ctx context.Context,
	userID uint64,
	tokenHash string,
	ip string,
	userAgent string,
	expiresAt time.Time,
) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO user_sessions (
			user_id,
			token_hash,
			ip_address,
			user_agent,
			expires_at
		) VALUES (?, ?, ?, ?, ?)
	`,
		userID,
		tokenHash,
		ip,
		userAgent,
		expiresAt,
	)

	return err
}

func (r *AuthRepository) IsSessionActive(
	ctx context.Context,
	tokenHash string,
) (bool, error) {
	var count int64

	err := r.DB.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM user_sessions s
		INNER JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ?
		  AND s.revoked_at IS NULL
		  AND s.expires_at > NOW()
		  AND u.status = 'active'
		  AND u.deleted_at IS NULL
	`, tokenHash)

	return count > 0, err
}

func (r *AuthRepository) RevokeSession(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE token_hash = ?
		  AND revoked_at IS NULL
	`, tokenHash)

	return err
}

func (r *AuthRepository) RecordLogin(
	ctx context.Context,
	userID *uint64,
	username string,
	ip string,
	status string,
	detail string,
) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO login_logs (
			user_id,
			username,
			ip_address,
			status,
			detail
		) VALUES (?, ?, ?, ?, ?)
	`,
		userID,
		username,
		ip,
		status,
		detail,
	)

	return err
}

func (r *AuthRepository) RecordFailedLogin(
	ctx context.Context,
	userID uint64,
) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE users
		SET
			failed_login_attempts = LEAST(failed_login_attempts + 1, 255),
			locked_until = CASE
				WHEN failed_login_attempts + 1 >= 5
				THEN DATE_ADD(NOW(), INTERVAL 15 MINUTE)
				ELSE locked_until
			END,
			status = CASE
				WHEN failed_login_attempts + 1 >= 5
				THEN 'locked'
				ELSE status
			END
		WHERE id = ?
		  AND status = 'active'
	`, userID)

	return err
}

func (r *AuthRepository) ResetFailedLogin(
	ctx context.Context,
	userID uint64,
) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE users
		SET
			failed_login_attempts = 0,
			locked_until = NULL,
			status = 'active',
			last_login_at = NOW()
		WHERE id = ?
		  AND deleted_at IS NULL
	`, userID)

	return err
}

func (r *AuthRepository) GetUserPermissions(
	ctx context.Context,
	userID uint64,
) ([]string, error) {
	permissions := make([]string, 0)

	err := r.DB.SelectContext(ctx, &permissions, `
		SELECT DISTINCT p.code
		FROM user_roles ur
		INNER JOIN role_permissions rp
			ON rp.role_id = ur.role_id
		INNER JOIN permissions p
			ON p.id = rp.permission_id
		WHERE ur.user_id = ?
	`, userID)

	return permissions, err
}

func (r *AuthRepository) GetUserRole(
	ctx context.Context,
	userID uint64,
) (string, error) {
	var role string

	err := r.DB.GetContext(ctx, &role, `
		SELECT r.code
		FROM user_roles ur
		INNER JOIN roles r
			ON r.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY
			CASE r.code
				WHEN 'SUPER_ADMIN' THEN 1
				WHEN 'IT_AUDIT_MANAGER' THEN 2
				WHEN 'INTERNAL_AUDITOR' THEN 3
				WHEN 'MANAGEMENT' THEN 4
				ELSE 5
			END,
			ur.id ASC
		LIMIT 1
	`, userID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return role, err
}

func (r *AuthRepository) GetUserRoles(
	ctx context.Context,
	userID uint64,
) ([]string, error) {
	roles := make([]string, 0)

	err := r.DB.SelectContext(ctx, &roles, `
		SELECT DISTINCT r.code
		FROM user_roles ur
		INNER JOIN roles r
			ON r.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY r.code ASC
	`, userID)

	return roles, err
}
