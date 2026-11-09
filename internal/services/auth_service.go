package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"

	"aacs-api/internal/config"
	"aacs-api/internal/models"
	"aacs-api/internal/repositories"
)

type AuthService struct {
	Repository     *repositories.AuthRepository
	UserRepository *repositories.UserRepository
	Config         *config.Config
}

type TokenClaims struct {
	jwt.RegisteredClaims
}

func NewAuthService(
	repo *repositories.AuthRepository,
	userRepo *repositories.UserRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		Repository:     repo,
		UserRepository: userRepo,
		Config:         cfg,
	}
}

func (s *AuthService) Bootstrap(
	ctx context.Context,
	input models.BootstrapRequest,
	ip string,
	userAgent string,
) (*models.AuthResponse, error) {
	count, err := s.Repository.CountUsers(ctx)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, errors.New("bootstrap is already completed")
	}

	createInput := models.UserCreate{
		Username: input.Username,
		Email:    input.Email,
		FullName: input.FullName,
		Phone:    input.Phone,
		Password: input.Password,
		RoleCode: "SUPER_ADMIN",
	}

	return s.createAndLogin(ctx, createInput, nil, true, ip, userAgent)
}

func (s *AuthService) Register(
	ctx context.Context,
	input models.UserCreate,
	createdBy uint64,
	ip string,
	userAgent string,
) (*models.User, error) {
	if err := validateRegistration(
		input.Username,
		input.FullName,
		input.Password,
	); err != nil {
		return nil, err
	}

	if input.RoleCode == "" || input.RoleCode == "SUPER_ADMIN" {
		return nil, errors.New("invalid role assignment")
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	id, err := s.Repository.CreateUser(
		ctx,
		input,
		hash,
		&createdBy,
		false,
	)
	if err != nil {
		return nil, err
	}

	return s.UserRepository.GetByID(ctx, id)
}

func (s *AuthService) createAndLogin(
	ctx context.Context,
	input models.UserCreate,
	createdBy *uint64,
	isBootstrap bool,
	ip string,
	userAgent string,
) (*models.AuthResponse, error) {
	if err := validateRegistration(
		input.Username,
		input.FullName,
		input.Password,
	); err != nil {
		return nil, err
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	id, err := s.Repository.CreateUser(
		ctx,
		input,
		hash,
		createdBy,
		isBootstrap,
	)
	if err != nil {
		return nil, err
	}

	user, err := s.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.generateToken(id)
	if err != nil {
		return nil, err
	}

	if err := s.Repository.SaveSession(
		ctx,
		id,
		hashToken(token),
		ip,
		userAgent,
		expiresAt,
	); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		User:        *user,
	}, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	input models.LoginRequest,
	ip string,
	userAgent string,
) (*models.AuthResponse, error) {
	user, err := s.Repository.FindByUsername(
		ctx,
		strings.TrimSpace(input.Username),
	)

	if err != nil {
		_ = s.Repository.RecordLogin(
			ctx,
			nil,
			input.Username,
			ip,
			"failed",
			"Invalid credentials",
		)

		return nil, errors.New("invalid username or password")
	}

	if user.Status == "inactive" {
		return nil, errors.New("account is inactive")
	}

	if user.Status == "locked" {
		if user.LockedUntil == nil || time.Now().Before(*user.LockedUntil) {
			_ = s.Repository.RecordLogin(
				ctx,
				&user.ID,
				user.Username,
				ip,
				"locked",
				"Account locked",
			)

			return nil, errors.New("account is locked")
		}
	}

	if !verifyPassword(input.Password, user.PasswordHash) {
		_ = s.Repository.RecordFailedLogin(ctx, user.ID)

		_ = s.Repository.RecordLogin(
			ctx,
			&user.ID,
			user.Username,
			ip,
			"failed",
			"Invalid password",
		)

		return nil, errors.New("invalid username or password")
	}

	if err := s.Repository.ResetFailedLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	token, expiresAt, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	if err := s.Repository.SaveSession(
		ctx,
		user.ID,
		hashToken(token),
		ip,
		userAgent,
		expiresAt,
	); err != nil {
		return nil, err
	}

	_ = s.Repository.RecordLogin(
		ctx,
		&user.ID,
		user.Username,
		ip,
		"success",
		"Login successful",
	)

	currentUser, err := s.UserRepository.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		User:        *currentUser,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.Repository.RevokeSession(ctx, hashToken(token))
}

func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	if len(s.Config.JWTSecret) < 32 {
		return nil, errors.New("JWT secret is not configured correctly")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(s.Config.JWTSecret), nil
		},
		jwt.WithIssuer("aacs-plantation"),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) IsSessionActive(
	ctx context.Context,
	token string,
) (bool, error) {
	return s.Repository.IsSessionActive(ctx, hashToken(token))
}

func (s *AuthService) HasPermission(
	ctx context.Context,
	userID uint64,
	permission string,
) (bool, error) {
	roles, err := s.Repository.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	if len(roles) == 0 {
		return false, nil
	}

	for _, role := range roles {
		if role == "SUPER_ADMIN" {
			return true, nil
		}

		switch role {
		case "IT_AUDIT_MANAGER", "INTERNAL_AUDITOR", "MANAGEMENT":
		default:
			return false, nil
		}
	}

	permissions, err := s.Repository.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, item := range permissions {
		if item == permission {
			return true, nil
		}
	}

	return false, nil
}
func (s *AuthService) generateToken(
	userID uint64,
) (string, time.Time, error) {
	if len(s.Config.JWTSecret) < 32 {
		return "", time.Time{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}

	now := time.Now()
	expiresAt := now.Add(
		time.Duration(s.Config.JWTExpiresHours) * time.Hour,
	)

	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		return "", time.Time{}, err
	}

	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(userID),
			ID:        hex.EncodeToString(randomID),
			Issuer:    "aacs-plantation",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.Config.JWTSecret))

	return signed, expiresAt, err
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func validateRegistration(
	username string,
	fullName string,
	password string,
) error {
	username = strings.TrimSpace(username)
	fullName = strings.TrimSpace(fullName)

	if len(username) < 3 || len(username) > 60 {
		return errors.New("username must be between 3 and 60 characters")
	}

	if fullName == "" || len(fullName) > 150 {
		return errors.New("full_name is required and must not exceed 150 characters")
	}

	if len(password) < 12 {
		return errors.New("password must contain at least 12 characters")
	}

	return nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	const (
		memory      = 64 * 1024
		iterations  = 3
		parallelism = 2
		keyLength   = 32
	)

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		b64Salt,
		b64Hash,
	), nil
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")

	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8

	_, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&memory,
		&iterations,
		&parallelism,
	)

	if err != nil ||
		memory < 8*1024 ||
		memory > 256*1024 ||
		iterations == 0 ||
		iterations > 10 ||
		parallelism == 0 ||
		parallelism > 8 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 {
		return false
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) < 16 || len(expected) > 64 {
		return false
	}

	actual := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expected)),
	)

	return subtle.ConstantTimeCompare(actual, expected) == 1
}
