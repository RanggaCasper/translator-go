package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"subtitle-translator/internal/models"
)

type fakeSubtitleRepository struct {
	subtitleByID       *models.Subtitle
	subtitleByPrimary  *models.Subtitle
	updatedContent     string
	updateContentCalls int
}

func (f *fakeSubtitleRepository) Create(subtitle *models.Subtitle, content string) error {
	return errors.New("not implemented in test")
}

func (f *fakeSubtitleRepository) GetBySubtitleID(subtitleID string) (*models.Subtitle, error) {
	if f.subtitleByID == nil {
		return nil, errors.New("not found")
	}
	return f.subtitleByID, nil
}

func (f *fakeSubtitleRepository) GetAll(page, limit int, targetLang string) ([]models.Subtitle, int64, error) {
	return nil, 0, nil
}

func (f *fakeSubtitleRepository) GetByID(id uint) (*models.Subtitle, error) {
	if f.subtitleByPrimary == nil {
		return nil, errors.New("not found")
	}
	return f.subtitleByPrimary, nil
}

func (f *fakeSubtitleRepository) Update(subtitle *models.Subtitle) error {
	f.subtitleByID = subtitle
	f.subtitleByPrimary = subtitle
	return nil
}

func (f *fakeSubtitleRepository) UpdateContent(id uint, content string) error {
	f.updatedContent = content
	f.updateContentCalls++
	return os.WriteFile(f.subtitleByID.FilePath, []byte(content), 0644)
}

func (f *fakeSubtitleRepository) Delete(id uint) error {
	return nil
}

func TestTranslateSubtitle_ExistingCachedContentIsNormalizedWhenRefreshFalse(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "cached.vtt")

	cached := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nI-Artinya, kesan seseorang terhadap sesuatu!\n\n00:00:01.000 --> 00:00:02.000\nL-Ayo mulai kelasnya!\n"
	if err := os.WriteFile(filePath, []byte(cached), 0644); err != nil {
		t.Fatalf("failed to prepare cached subtitle file: %v", err)
	}

	sub := &models.Subtitle{
		ID:         7,
		SubtitleID: "504b378cd0791e1ba3bd45bfe84b9176",
		URL:        "https://mgstatics.xyz/subtitle/047f9c3c943db7c39d2f9c51097921a2/047f9c3c943db7c39d2f9c51097921a2.vtt",
		TargetLang: "id",
		SourceLang: "auto",
		Format:     "vtt",
		FilePath:   filePath,
		FileSize:   int64(len(cached)),
		IsLock:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	repo := &fakeSubtitleRepository{subtitleByID: sub, subtitleByPrimary: sub}
	svc := NewSubtitleService(repo)

	result, err := svc.TranslateSubtitle(sub.URL, sub.Format, sub.TargetLang, sub.SourceLang, "https://example.com", false, false)
	if err != nil {
		t.Fatalf("TranslateSubtitle returned error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected non-nil result")
	}

	if !strings.Contains(result.Content, "A-Artinya, kesan seseorang terhadap sesuatu!") {
		t.Fatalf("expected I-Artinya to normalize in response, got: %q", result.Content)
	}

	if !strings.Contains(result.Content, "A-Ayo mulai kelasnya!") {
		t.Fatalf("expected L-Ayo to normalize in response, got: %q", result.Content)
	}

	if repo.updateContentCalls == 0 {
		t.Fatalf("expected normalized cached content to be persisted")
	}
}
