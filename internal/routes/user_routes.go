package routes

import (
	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/controllers"
	"aacs-api/internal/middleware"
)

func RegisterUserRoutes(
	api fiber.Router,
	controller *controllers.UserController,
	auth *middleware.AuthMiddleware,
	permission *middleware.PermissionMiddleware,
) {
	group := api.Group(
		"/users",
		auth.Handle,
		permission.Require("users.manage"),
	)

	group.Post("/register", controller.Register)
	group.Get("/", controller.List)
	group.Get("/:id", controller.Detail)
	group.Put("/:id", controller.Update)
}
