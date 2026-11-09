package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/models"
	"aacs-api/internal/services"
)

type UserController struct {
	UserService *services.UserService
	AuthService *services.AuthService
}

func NewUserController(
	userService *services.UserService,
	authService *services.AuthService,
) *UserController {
	return &UserController{
		UserService: userService,
		AuthService: authService,
	}
}

func (h *UserController) Register(c fiber.Ctx) error {
	var input models.UserCreate

	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	createdBy, ok := c.Locals("userID").(uint64)
	if !ok {
		return c.SendStatus(http.StatusUnauthorized)
	}

	user, err := h.AuthService.Register(
		c.Context(),
		input,
		createdBy,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"data":    user,
	})
}

func (h *UserController) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	users, total, err := h.UserService.List(c.Context(), page, limit)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve users",
		})
	}

	return c.JSON(fiber.Map{
		"data": users,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *UserController) Detail(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	user, err := h.UserService.Get(c.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve user",
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

func (h *UserController) Update(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	var input models.UserUpdate

	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	updatedBy, ok := c.Locals("userID").(uint64)
	if !ok {
		return c.SendStatus(http.StatusUnauthorized)
	}

	if err := h.UserService.Update(c.Context(), id, input, updatedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}

		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
	})
}