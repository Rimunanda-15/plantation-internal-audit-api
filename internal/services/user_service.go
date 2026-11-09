package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"aacs-api/internal/models"
	"aacs-api/internal/repositories"
)

type UserService struct {
	Repository *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{
		Repository: repo,
	}
}

func (s *UserService) List(
	ctx context.Context,
	page int,
	limit int,
) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	users, err := s.Repository.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.Repository.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *UserService) Get(
	ctx context.Context,
	id uint64,
) (*models.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.Repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(
	ctx context.Context,
	id uint64,
	input models.UserUpdate,
	updatedBy uint64,
) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	if updatedBy == 0 {
		return errors.New("invalid updater ID")
	}

	if input.FullName == nil &&
		input.Email == nil &&
		input.Phone == nil &&
		input.Status == nil {
		return errors.New("no fields to update")
	}

	if input.FullName != nil {
		fullName := strings.TrimSpace(*input.FullName)

		if fullName == "" {
			return errors.New("full_name cannot be empty")
		}

		if len(fullName) > 150 {
			return errors.New("full_name must not exceed 150 characters")
		}

		input.FullName = &fullName
	}

	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)

		if email != "" {
			if len(email) > 150 {
				return errors.New("email must not exceed 150 characters")
			}

			parsed, err := mail.ParseAddress(email)
			if err != nil || parsed.Address != email {
				return errors.New("invalid email address")
			}
		}

		input.Email = &email
	}

	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)

		if len(phone) > 30 {
			return errors.New("phone must not exceed 30 characters")
		}

		input.Phone = &phone
	}

	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)

		switch status {
		case "active", "inactive":
		default:
			return errors.New("status must be active or inactive")
		}

		input.Status = &status
	}

	return s.Repository.Update(ctx, id, input, updatedBy)
}
