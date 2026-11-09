package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"aacs-api/internal/models"
)

type UserRepository struct {
	DB *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	var user models.User

	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.full_name,
			u.phone,
			u.status,
			u.last_login_at,
			u.created_at,
			u.created_by,
			u.updated_at,
			u.updated_by,
			COALESCE(r.code, '') AS role_code,
			COALESCE(r.name, '') AS role_name,
			ur.company_id
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.id = ?
		  AND u.deleted_at IS NULL
		ORDER BY ur.id
		LIMIT 1
	`

	err := r.DB.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	users := make([]models.User, 0)

	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.full_name,
			u.phone,
			u.status,
			u.last_login_at,
			u.created_at,
			u.created_by,
			u.updated_at,
			u.updated_by,
			COALESCE(r.code, '') AS role_code,
			COALESCE(r.name, '') AS role_name,
			ur.company_id
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.deleted_at IS NULL
		ORDER BY u.id DESC
		LIMIT ? OFFSET ?
	`

	err := r.DB.SelectContext(ctx, &users, query, limit, offset)
	return users, err
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var total int64

	err := r.DB.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM users
		WHERE deleted_at IS NULL
	`)

	return total, err
}

func (r *UserRepository) Update(
	ctx context.Context,
	id uint64,
	input models.UserUpdate,
	updatedBy uint64,
) error {
	if input.Status != nil &&
		*input.Status != "active" &&
		*input.Status != "inactive" &&
		*input.Status != "locked" {
		return fmt.Errorf("invalid user status")
	}

	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.GetContext(ctx, &exists, `
		SELECT COUNT(*)
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}

	if exists == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE users
		SET
			full_name = COALESCE(?, full_name),
			email = COALESCE(?, email),
			phone = COALESCE(?, phone),
			status = COALESCE(?, status),
			updated_at = NOW(),
			updated_by = ?
		WHERE id = ?
	`, input.FullName, input.Email, input.Phone, input.Status, updatedBy, id)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO activity_logs
			(user_id, module, action, entity_type, entity_id, ip_address)
		VALUES (?, 'users', 'update', 'user', ?, NULL)
	`, updatedBy, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
