package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"aacs-api/internal/config"
	"aacs-api/internal/database"
	"aacs-api/internal/routes"
)

func main() {
	cfg := config.Load()

	log.Println("Starting", cfg.AppName)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Printf("WARNING: %v", err)
	} else {
		log.Println("Database connected successfully")
		defer db.Close()
	}

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	routes.Register(app, db, cfg)

	log.Printf("AACS API running on port %s", cfg.AppPort)

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start API: %v", err)
	}
}
