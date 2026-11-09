package routes

import (
	"github.com/gofiber/fiber/v3"

	"aacs-api/internal/controllers"
	"aacs-api/internal/middleware"
)

func RegisterOrganizationRoutes(
	api fiber.Router,
	controller *controllers.OrganizationController,
	auth *middleware.AuthMiddleware,
	permission *middleware.PermissionMiddleware,
) {
	companies := api.Group(
		"/companies",
		auth.Handle,
		permission.Require("organization.manage"),
	)

	companies.Get("/", controller.ListCompanies)
	companies.Get("/:id", controller.DetailCompany)
	companies.Post("/", controller.CreateCompany)
	companies.Put("/:id", controller.UpdateCompany)

	organizations := api.Group(
		"/organizations",
		auth.Handle,
		permission.Require("organization.manage"),
	)

	organizations.Get("/", controller.ListOrganizations)
	organizations.Get("/:id", controller.DetailOrganization)
	organizations.Post("/", controller.CreateOrganization)
	organizations.Put("/:id", controller.UpdateOrganization)

	departments := api.Group(
		"/departments",
		auth.Handle,
		permission.Require("organization.manage"),
	)

	departments.Get("/", controller.ListDepartments)
	departments.Get("/:id", controller.DetailDepartment)
	departments.Post("/", controller.CreateDepartment)
	departments.Put("/:id", controller.UpdateDepartment)

	estateBlocks := api.Group(
		"/estate-blocks",
		auth.Handle,
		permission.Require("organization.manage"),
	)

	estateBlocks.Get("/", controller.ListEstateBlocks)
	estateBlocks.Get("/:id", controller.DetailEstateBlock)
	estateBlocks.Post("/", controller.CreateEstateBlock)
	estateBlocks.Put("/:id", controller.UpdateEstateBlock)
}