package routes

import (
	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/controllers"
	"aacs-api/internal/middleware"
)

func RegisterAuthRoutes(
	api fiber.Router,
	controller *controllers.AuthController,
	auth *middleware.AuthMiddleware,
) {
	group := api.Group("/auth")

	group.Post("/bootstrap", controller.Bootstrap)
	group.Post("/login", controller.Login)

	group.Post("/logout", auth.Handle, controller.Logout)
	group.Get("/me", auth.Handle, controller.Me)
}
