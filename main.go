package main

import (
	"log"
	"os"

	"hyw-webpics/config"
	"hyw-webpics/database"
	"hyw-webpics/handlers"
	"hyw-webpics/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load configuration
	config.Load()

	// Connect to database
	database.Connect()
	defer database.Close()

	// Create uploads directory
	if err := os.MkdirAll(config.AppConfig.UploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Admin-Token",
	}))

	// API routes
	api := app.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Get("/me", middleware.UserAuth(), handlers.GetMe)

	// Image routes
	images := api.Group("/images")
	images.Post("/upload", middleware.UserAuth(), handlers.UploadImage)
	images.Get("/", handlers.GetApprovedImages)
	images.Get("/random", handlers.GetRandomImage)

	// Category routes (Public List)
	api.Get("/categories", handlers.GetCategories)

	// Admin functionality
	admin := api.Group("/admin")

	// Public Routes
	admin.Post("/login", handlers.AdminLogin)
	admin.Post("/logout", handlers.AdminLogout)

	// Protected Routes
	admin.Use(middleware.AdminAuth()) // Ensure all admin routes BELOW are protected

	// Admin Category Management
	admin.Post("/categories", handlers.CreateCategory)
	admin.Put("/categories/:id", handlers.UpdateCategory)
	admin.Delete("/categories/:id", handlers.DeleteCategory)

	// Admin Stats & Ops
	admin.Get("/stats", handlers.GetAdminStats)
	admin.Get("/pending", handlers.GetPendingImages) // Keep for convenience or legacy
	admin.Get("/images", handlers.GetAdminImages)    // New generic list
	admin.Post("/approve/:id", handlers.ApproveImage)
	admin.Post("/bulk-approve", handlers.BulkApproveImages)
	admin.Post("/bulk-delete", handlers.BulkDeleteImages)
	admin.Delete("/images/:id", handlers.RejectImage) // Renamed usage, handlers.RejectImage now does generic delete
	admin.Post("/reject/:id", handlers.RejectImage)   // Keep alias for compatibility

	// Serve uploaded images
	app.Static("/uploads", config.AppConfig.UploadDir)

	// Serve Vue frontend (production build)
	app.Static("/", "./web/dist")

	// SPA fallback - serve index.html for all non-API routes
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./web/dist/index.html")
	})

	log.Printf("Server starting on port %s", config.AppConfig.Port)
	log.Fatal(app.Listen(":" + config.AppConfig.Port))
}
