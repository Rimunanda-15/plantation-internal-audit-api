package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/services"
)

type PermissionMiddleware struct {
	Service *services.AuthService
}

func NewPermissionMiddleware(service *services.AuthService) *PermissionMiddleware {
	return &PermissionMiddleware{
		Service: service,
	}
}

func (m *PermissionMiddleware) Require(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uint64)

		if !ok || userID == 0 {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "ERROR",
				"message": "Unauthorized",
			})
		}

		allowed, err := m.Service.HasPermission(
			c.Context(),
			userID,
			permission,
		)

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "ERROR",
				"message": "Failed to validate permission",
			})
		}

		if !allowed {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"status":   "ERROR",
				"message":  "You do not have permission to access this resource",
				"required": permission,
			})
		}

		return c.Next()
	}
}
