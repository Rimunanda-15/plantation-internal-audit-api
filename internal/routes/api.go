package routes

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"

	"aacs-api/internal/config"
	"aacs-api/internal/controllers"
	"aacs-api/internal/middleware"
	"aacs-api/internal/repositories"
	"aacs-api/internal/services"
)

func Register(app *fiber.App, db *sqlx.DB, cfg *config.Config) {
	api := app.Group("/api/v1")

	// ========================================================
	// HEALTH CHECK
	// ========================================================

	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"message": "AACS Plantation API is running",
		})
	})

	api.Get("/health/db", func(c fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "ERROR",
				"database": "DISCONNECTED",
			})
		}

		if err := db.PingContext(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "ERROR",
				"database": "DISCONNECTED",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "OK",
			"database": "CONNECTED",
			"name":     cfg.DBName,
		})
	})

	// ========================================================
	// REPOSITORIES
	// ========================================================

	userRepo := repositories.NewUserRepository(db)
	authRepo := repositories.NewAuthRepository(db)
	organizationRepo := repositories.NewOrganizationRepository(db)

	// ========================================================
	// SERVICES
	// ========================================================

	userService := services.NewUserService(userRepo)

	authService := services.NewAuthService(
		authRepo,
		userRepo,
		cfg,
	)

	organizationService := services.NewOrganizationService(
		organizationRepo,
	)

	// ========================================================
	// CONTROLLERS
	// ========================================================

	authController := controllers.NewAuthController(
		authService,
		userService,
	)

	userController := controllers.NewUserController(
		userService,
		authService,
	)

	organizationController := controllers.NewOrganizationController(
		organizationService,
	)

	// ========================================================
	// MIDDLEWARE
	// ========================================================

	authMiddleware := middleware.NewAuthMiddleware(authService)

	permissionMiddleware := middleware.NewPermissionMiddleware(
		authService,
	)

	// ========================================================
	// AUTH ROUTES
	// ========================================================

	RegisterAuthRoutes(
		api,
		authController,
		authMiddleware,
	)

	// ========================================================
	// USER ROUTES
	// ========================================================

	RegisterUserRoutes(
		api,
		userController,
		authMiddleware,
		permissionMiddleware,
	)

	// ========================================================
	// PLANTATION MASTER ROUTES
	// ========================================================

	RegisterOrganizationRoutes(
		api,
		organizationController,
		authMiddleware,
		permissionMiddleware,
	)

	// ========================================================
	// DEBUG REGISTERED ROUTES
	// ========================================================

	log.Println("================================================")
	log.Println("REGISTERED API ROUTES")
	log.Println("================================================")

	for _, route := range app.GetRoutes() {
		log.Printf(
			"%-8s %s",
			route.Method,
			route.Path,
		)
	}

	log.Println("================================================")
}
