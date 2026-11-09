package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/services"
)

type AuthMiddleware struct {
	Service *services.AuthService
}

func NewAuthMiddleware(service *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		Service: service,
	}
}

func (m *AuthMiddleware) Handle(c fiber.Ctx) error {
	header := strings.TrimSpace(c.Get("Authorization"))

	if header == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Authorization header is required",
		})
	}

	parts := strings.SplitN(header, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Invalid authorization format. Use Bearer token",
		})
	}

	token := strings.TrimSpace(parts[1])

	if token == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Bearer token is empty",
		})
	}

	claims, err := m.Service.ValidateToken(token)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Invalid or expired token",
		})
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || userID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Invalid user identity",
		})
	}

	active, err := m.Service.IsSessionActive(c.Context(), token)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Failed to validate user session",
		})
	}

	if !active {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "ERROR",
			"message": "Session expired or revoked",
		})
	}

	c.Locals("userID", userID)
	c.Locals("token", token)
	c.Locals("claims", claims)

	return c.Next()
}
