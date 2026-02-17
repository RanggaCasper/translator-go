package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"subtitle-translator/config"
	"subtitle-translator/internal/routes"
	"subtitle-translator/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	config.InitDB()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	routes.SetupRoutes(app)

	// Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Database metadata stored in MySQL")
		log.Printf("Subtitle files stored in: storage/subtitles/")
		errCh <- app.Listen(":" + port)
	}()

	// Wait for OS signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal: %s. Shutting down...", sig.String())
	case err := <-errCh:
		// Listen() returns error when server fails to start or stops unexpectedly
		log.Printf("Server error: %v. Shutting down...", err)
	}

	// Graceful shutdown
	if err := app.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped.")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(utils.ErrorResponse{
		Status:  false,
		Error:   err.Error(),
		Message: "An unexpected error occurred",
	})
}
