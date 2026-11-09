package controllers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/models"
	"aacs-api/internal/services"
)

type AuthController struct {
	Service     *services.AuthService
	UserService *services.UserService
}

func NewAuthController(
	service *services.AuthService,
	userService *services.UserService,
) *AuthController {
	return &AuthController{
		Service:     service,
		UserService: userService,
	}
}

func (h *AuthController) Bootstrap(c fiber.Ctx) error {
	var input models.BootstrapRequest

	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	result, err := h.Service.Bootstrap(
		c.Context(),
		input,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Initial administrator created successfully",
		"data":    result,
	})
}

func (h *AuthController) Login(c fiber.Ctx) error {
	var input models.LoginRequest

	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if input.Username == "" || input.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Username and password are required",
		})
	}

	result, err := h.Service.Login(
		c.Context(),
		input,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"data":    result,
	})
}

func (h *AuthController) Logout(c fiber.Ctx) error {
	token := c.Locals("token")

	tokenString, ok := token.(string)
	if !ok || tokenString == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authentication required",
		})
	}

	if err := h.Service.Logout(c.Context(), tokenString); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to logout",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}

func (h *AuthController) Me(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint64)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authentication required",
		})
	}

	user, err := h.UserService.Get(c.Context(), userID)
	if err != nil {
		if errors.Is(err, errors.New("sql: no rows in result set")) {
			return c.SendStatus(http.StatusNotFound)
		}

		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve user",
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}
