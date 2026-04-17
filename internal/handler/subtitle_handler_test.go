package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"subtitle-translator/internal/models"

	"github.com/gofiber/fiber/v2"
)

type fakeSubtitleService struct {
	called       bool
	url          string
	format       string
	targetLang   string
	sourceLang   string
	referer      string
	isRefresh    bool
	isLock       bool
	result       *models.SubtitleWithContent
	translateErr error
}

func (f *fakeSubtitleService) TranslateSubtitle(url, format, targetLang, sourceLang, referer string, isRefresh, isLock bool) (*models.SubtitleWithContent, error) {
	f.called = true
	f.url = url
	f.format = format
	f.targetLang = targetLang
	f.sourceLang = sourceLang
	f.referer = referer
	f.isRefresh = isRefresh
	f.isLock = isLock

	if f.translateErr != nil {
		return nil, f.translateErr
	}
	return f.result, nil
}

func (f *fakeSubtitleService) TranslateTexts(texts []string, targetLang, sourceLang string) ([]string, error) {
	return texts, nil
}

func (f *fakeSubtitleService) GetAllSubtitles(page, limit int, targetLang string) ([]models.Subtitle, int64, int, error) {
	return nil, 0, 0, nil
}

func (f *fakeSubtitleService) GetSubtitleByID(id uint) (*models.SubtitleWithContent, error) {
	return nil, nil
}

func (f *fakeSubtitleService) UpdateSubtitle(id uint, content string) (*models.SubtitleWithContent, error) {
	return nil, nil
}

func (f *fakeSubtitleService) DeleteSubtitle(id uint) error {
	return nil
}

func TestTranslateSubtitle_RequestWithRefreshFalse(t *testing.T) {
	app := fiber.New()

	stub := &fakeSubtitleService{
		result: &models.SubtitleWithContent{
			ID:         1,
			SubtitleID: "abc123",
			URL:        "https://mgstatics.xyz/subtitle/047f9c3c943db7c39d2f9c51097921a2/047f9c3c943db7c39d2f9c51097921a2.vtt",
			TargetLang: "id",
			SourceLang: "auto",
			Format:     "vtt",
			FilePath:   "storage/subtitles/abc123.vtt",
			Content:    "WEBVTT\n\n1\n00:00:00.000 --> 00:00:01.000\nA-Artinya, kesan seseorang terhadap sesuatu!\n",
			FileSize:   90,
			IsLock:     false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	h := NewSubtitleHandler(stub)
	app.Post("/api/v1/subtitles/translate", h.TranslateSubtitle)

	payload := map[string]interface{}{
		"url":         "https://mgstatics.xyz/subtitle/047f9c3c943db7c39d2f9c51097921a2/047f9c3c943db7c39d2f9c51097921a2.vtt",
		"format":      "vtt",
		"target_lang": "id",
		"source_lang": "auto",
		"referer":     "https://example.com",
		"is_refresh":  false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/subtitles/translate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusOK)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(bodyResp, &decoded); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if status, ok := decoded["status"].(bool); !ok || !status {
		t.Fatalf("unexpected response status field: %v", decoded["status"])
	}

	if !stub.called {
		t.Fatalf("service TranslateSubtitle was not called")
	}
	if stub.isRefresh {
		t.Fatalf("is_refresh should stay false when payload sets false")
	}
	if stub.isLock {
		t.Fatalf("is_lock should default to false when omitted")
	}
}

func TestTranslateSubtitle_RequestWithRefreshTrue(t *testing.T) {
	app := fiber.New()

	stub := &fakeSubtitleService{
		result: &models.SubtitleWithContent{
			ID:         1,
			SubtitleID: "abc123",
			URL:        "https://mgstatics.xyz/subtitle/047f9c3c943db7c39d2f9c51097921a2/047f9c3c943db7c39d2f9c51097921a2.vtt",
			TargetLang: "id",
			SourceLang: "auto",
			Format:     "vtt",
			FilePath:   "storage/subtitles/abc123.vtt",
			Content:    "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nA-Artinya, kesan seseorang terhadap sesuatu!\n",
			FileSize:   90,
			IsLock:     false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	h := NewSubtitleHandler(stub)
	app.Post("/api/v1/subtitles/translate", h.TranslateSubtitle)

	payload := map[string]interface{}{
		"url":         "https://mgstatics.xyz/subtitle/047f9c3c943db7c39d2f9c51097921a2/047f9c3c943db7c39d2f9c51097921a2.vtt",
		"format":      "vtt",
		"target_lang": "id",
		"source_lang": "auto",
		"referer":     "https://example.com",
		"is_refresh":  true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/subtitles/translate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusOK)
	}

	if !stub.called {
		t.Fatalf("service TranslateSubtitle was not called")
	}
	if !stub.isRefresh {
		t.Fatalf("is_refresh should be true when payload sets true")
	}
}
