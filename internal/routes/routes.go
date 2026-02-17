package routes

import (
	"subtitle-translator/config"
	"subtitle-translator/internal/handler"
	"subtitle-translator/internal/repository"
	"subtitle-translator/internal/service"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Serve subtitle files so generated .vtt can be accessed by URL
	app.Static("/storage/subtitles", repository.StorageDir)

	// Initialize dependencies
	subtitleRepo := repository.NewSubtitleRepository(config.DB)
	subtitleService := service.NewSubtitleService(subtitleRepo)
	subtitleHandler := handler.NewSubtitleHandler(subtitleService)

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Subtitle routes
	subtitle := v1.Group("/subtitles")
	subtitle.Post("/translate", subtitleHandler.TranslateSubtitle)
	subtitle.Post("/translate/text", subtitleHandler.TranslateText)
	subtitle.Post("/translate/batch", subtitleHandler.TranslateBatchContent)
	subtitle.Get("/", subtitleHandler.GetAllSubtitles)
	subtitle.Get("/:id", subtitleHandler.GetSubtitleByID)
	subtitle.Put("/:id", subtitleHandler.UpdateSubtitle)
	subtitle.Delete("/:id", subtitleHandler.DeleteSubtitle)

	// Health check endpoint
	app.Get("/health", subtitleHandler.HealthCheck)
}
