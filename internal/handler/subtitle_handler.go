package handler

import (
	"strconv"
	"strings"
	"subtitle-translator/internal/service"
	"subtitle-translator/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type SubtitleHandler struct {
	service service.SubtitleService
}

func NewSubtitleHandler(service service.SubtitleService) *SubtitleHandler {
	return &SubtitleHandler{service: service}
}

type TranslateRequest struct {
	URL        string `json:"url" validate:"required"`
	Format     string `json:"format" validate:"required,oneof=vtt ass"`
	TargetLang string `json:"target_lang"`
	SourceLang string `json:"source_lang"`
	Referer    string `json:"referer"`
}

type UpdateSubtitleRequest struct {
	Content string `json:"content" validate:"required"`
}

type TranslateBatchContentItem struct {
	Type string      `json:"type"`
	Text string      `json:"text"`
	Src  interface{} `json:"src"`
}

type TranslateBatchContentData struct {
	Title   string                      `json:"title"`
	Content []TranslateBatchContentItem `json:"content"`
}

type TranslateBatchContentRequest struct {
	TargetLang string                    `json:"target_lang"`
	SourceLang string                    `json:"source_lang"`
	Data       TranslateBatchContentData `json:"data"`
}

type TranslateTextRequest struct {
	Text       string `json:"text"`
	TargetLang string `json:"target_lang"`
	SourceLang string `json:"source_lang"`
}

// TranslateSubtitle handles subtitle translation requests
func (h *SubtitleHandler) TranslateSubtitle(c *fiber.Ctx) error {
	var req TranslateRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	// Set defaults
	if req.TargetLang == "" {
		req.TargetLang = c.Query("target_lang", "id")
	}
	if req.SourceLang == "" {
		req.SourceLang = c.Query("source_lang", "auto")
	}

	// Validate
	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "URL is required",
			Message: "Please provide a subtitle URL",
		})
	}

	if req.Format != "vtt" && req.Format != "ass" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid format",
			Message: "Format must be 'vtt' or 'ass'",
		})
	}

	// Translate or get existing
	subtitle, err := h.service.TranslateSubtitle(
		req.URL,
		req.Format,
		req.TargetLang,
		req.SourceLang,
		req.Referer,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Translation failed",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data:   subtitle,
	})
}

// TranslateText handles translating a single text/sentence.
func (h *SubtitleHandler) TranslateText(c *fiber.Ctx) error {
	var req TranslateTextRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if req.TargetLang == "" {
		req.TargetLang = c.Query("target_lang", "id")
	}
	if req.SourceLang == "" {
		req.SourceLang = c.Query("source_lang", "auto")
	}

	if strings.TrimSpace(req.Text) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Text is required",
			Message: "Please provide text to translate",
		})
	}

	translated, err := h.service.TranslateTexts([]string{req.Text}, req.TargetLang, req.SourceLang)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Translation failed",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data: fiber.Map{
			"text":            req.Text,
			"translated_text": translated[0],
			"target_lang":     req.TargetLang,
			"source_lang":     req.SourceLang,
		},
	})
}

// TranslateBatchContent handles translating many text blocks in one request.
func (h *SubtitleHandler) TranslateBatchContent(c *fiber.Ctx) error {
	var req TranslateBatchContentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if req.TargetLang == "" {
		req.TargetLang = c.Query("target_lang", "id")
	}
	if req.SourceLang == "" {
		req.SourceLang = c.Query("source_lang", "auto")
	}

	hasTitle := strings.TrimSpace(req.Data.Title) != ""
	hasContent := len(req.Data.Content) > 0
	if !hasTitle && !hasContent {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "No translatable content",
			Message: "Provide data.title or data.content[].text",
		})
	}

	var texts []string
	titleIncluded := false
	contentTextIndexes := make([]int, 0, len(req.Data.Content))

	if hasTitle {
		texts = append(texts, req.Data.Title)
		titleIncluded = true
	}

	for i := range req.Data.Content {
		text := strings.TrimSpace(req.Data.Content[i].Text)
		if text == "" {
			continue
		}
		texts = append(texts, req.Data.Content[i].Text)
		contentTextIndexes = append(contentTextIndexes, i)
	}

	if len(texts) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "No translatable content",
			Message: "All provided text values are empty",
		})
	}

	translated, err := h.service.TranslateTexts(texts, req.TargetLang, req.SourceLang)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Batch translation failed",
			Message: err.Error(),
		})
	}

	cursor := 0
	if titleIncluded {
		req.Data.Title = translated[cursor]
		cursor++
	}
	for _, contentIdx := range contentTextIndexes {
		req.Data.Content[contentIdx].Text = translated[cursor]
		cursor++
	}

	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data: fiber.Map{
			"target_lang":      req.TargetLang,
			"source_lang":      req.SourceLang,
			"translated_count": len(texts),
			"data":             req.Data,
		},
	})
}

// GetAllSubtitles handles listing all subtitles with pagination
func (h *SubtitleHandler) GetAllSubtitles(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	targetLang := c.Query("target_lang", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	subtitles, _, totalPages, err := h.service.GetAllSubtitles(page, limit, targetLang)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Failed to fetch subtitles",
			Message: err.Error(),
		})
	}

	var nextPage, prevPage *int
	if page < totalPages {
		next := page + 1
		nextPage = &next
	}
	if page > 1 {
		prev := page - 1
		prevPage = &prev
	}

	return c.JSON(utils.PaginatedResponse{
		Status: true,
		Data:   subtitles,
		Meta: utils.Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			NextPage:    nextPage,
			PrevPage:    prevPage,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	})
}

// GetSubtitleByID handles fetching a single subtitle by ID
func (h *SubtitleHandler) GetSubtitleByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
		})
	}

	subtitle, err := h.service.GetSubtitleByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Subtitle not found",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data:   subtitle,
	})
}

// UpdateSubtitle handles updating subtitle content
func (h *SubtitleHandler) UpdateSubtitle(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
		})
	}

	var req UpdateSubtitleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Content is required",
			Message: "Please provide subtitle content",
		})
	}

	subtitle, err := h.service.UpdateSubtitle(uint(id), req.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Update failed",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data:   subtitle,
	})
}

// DeleteSubtitle handles deleting a subtitle
func (h *SubtitleHandler) DeleteSubtitle(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
		})
	}

	if err := h.service.DeleteSubtitle(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Status:  false,
			Error:   "Delete failed",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data: fiber.Map{
			"message": "Subtitle deleted successfully",
		},
	})
}

// HealthCheck handles health check requests
func (h *SubtitleHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(utils.SuccessResponse{
		Status: true,
		Data:   "OK",
	})
}
