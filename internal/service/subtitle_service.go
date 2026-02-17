package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"subtitle-translator/internal/models"
	"subtitle-translator/internal/repository"
	"subtitle-translator/pkg/translator"
	"time"

	"gorm.io/gorm"
)

type SubtitleService interface {
	TranslateSubtitle(url, format, targetLang, sourceLang, referer string) (*models.SubtitleWithContent, error)
	TranslateTexts(texts []string, targetLang, sourceLang string) ([]string, error)
	GetAllSubtitles(page, limit int, targetLang string) ([]models.Subtitle, int64, int, error)
	GetSubtitleByID(id uint) (*models.SubtitleWithContent, error)
	UpdateSubtitle(id uint, content string) (*models.SubtitleWithContent, error)
	DeleteSubtitle(id uint) error
}

type subtitleService struct {
	repo repository.SubtitleRepository
}

func NewSubtitleService(repo repository.SubtitleRepository) SubtitleService {
	return &subtitleService{
		repo: repo,
	}
}

func (s *subtitleService) TranslateSubtitle(url, format, targetLang, sourceLang, referer string) (*models.SubtitleWithContent, error) {
	// Generate subtitle ID
	subtitleID := s.generateSubtitleID(url, targetLang, format)
	filePath := repository.GenerateFilePath(subtitleID)

	// Check if already exists in database
	existing, err := s.repo.GetBySubtitleID(subtitleID)
	if err == nil {
		log.Printf("Subtitle already exists in DB with ID: %s, loading from file", subtitleID[:8])

		// Keep stored path URL-safe across platforms
		normalizedPath := filepath.ToSlash(existing.FilePath)
		if existing.FilePath != normalizedPath {
			existing.FilePath = normalizedPath
			if updateErr := s.repo.Update(existing); updateErr != nil {
				log.Printf("Failed to normalize file path for subtitle ID %s: %v", subtitleID[:8], updateErr)
			}
		}

		// Load content from file
		content, err := repository.LoadContent(existing.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load content: %w", err)
		}
		cleanedContent := translator.RemoveFontTags(content)
		if cleanedContent != content {
			content = cleanedContent
			if updateErr := s.repo.UpdateContent(existing.ID, content); updateErr != nil {
				log.Printf("Failed to persist cleaned content for subtitle ID %s: %v", subtitleID[:8], updateErr)
			}
		}

		return &models.SubtitleWithContent{
			ID:         existing.ID,
			SubtitleID: existing.SubtitleID,
			URL:        existing.URL,
			TargetLang: existing.TargetLang,
			SourceLang: existing.SourceLang,
			Format:     existing.Format,
			FilePath:   existing.FilePath,
			Content:    content,
			FileSize:   existing.FileSize,
			CreatedAt:  existing.CreatedAt,
			UpdatedAt:  existing.UpdatedAt,
		}, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	log.Printf("Subtitle not found with ID: %s, fetching and translating", subtitleID[:8])

	// Fetch and translate
	content, err := translator.FetchAndTranslate(url, format, targetLang, sourceLang, referer)
	if err != nil {
		return nil, err
	}
	content = translator.RemoveFontTags(content)

	// Create subtitle record
	subtitle := &models.Subtitle{
		SubtitleID: subtitleID,
		URL:        url,
		TargetLang: targetLang,
		SourceLang: sourceLang,
		Format:     format,
		FilePath:   filePath,
		FileSize:   int64(len(content)),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to database and file
	if err := s.repo.Create(subtitle, content); err != nil {
		log.Printf("Failed to save subtitle: %v", err)
		return nil, fmt.Errorf("failed to save subtitle: %w", err)
	}

	return &models.SubtitleWithContent{
		ID:         subtitle.ID,
		SubtitleID: subtitle.SubtitleID,
		URL:        subtitle.URL,
		TargetLang: subtitle.TargetLang,
		SourceLang: subtitle.SourceLang,
		Format:     subtitle.Format,
		FilePath:   subtitle.FilePath,
		Content:    content,
		FileSize:   subtitle.FileSize,
		CreatedAt:  subtitle.CreatedAt,
		UpdatedAt:  subtitle.UpdatedAt,
	}, nil
}

func (s *subtitleService) TranslateTexts(texts []string, targetLang, sourceLang string) ([]string, error) {
	if targetLang == "" {
		targetLang = "id"
	}
	if sourceLang == "" {
		sourceLang = "auto"
	}

	translated, err := translator.BatchTranslate(texts, targetLang, sourceLang)
	if err != nil {
		return nil, err
	}

	return translated, nil
}

func (s *subtitleService) GetAllSubtitles(page, limit int, targetLang string) ([]models.Subtitle, int64, int, error) {
	subtitles, total, err := s.repo.GetAll(page, limit, targetLang)
	if err != nil {
		return nil, 0, 0, err
	}

	for i := range subtitles {
		subtitles[i].FilePath = filepath.ToSlash(subtitles[i].FilePath)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return subtitles, total, totalPages, nil
}

func (s *subtitleService) GetSubtitleByID(id uint) (*models.SubtitleWithContent, error) {
	subtitle, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	subtitle.FilePath = filepath.ToSlash(subtitle.FilePath)

	// Load content from file
	content, err := repository.LoadContent(subtitle.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load content: %w", err)
	}
	cleanedContent := translator.RemoveFontTags(content)
	if cleanedContent != content {
		content = cleanedContent
		if updateErr := s.repo.UpdateContent(subtitle.ID, content); updateErr != nil {
			log.Printf("Failed to persist cleaned content for subtitle ID %s: %v", subtitle.SubtitleID[:8], updateErr)
		}
	}

	return &models.SubtitleWithContent{
		ID:         subtitle.ID,
		SubtitleID: subtitle.SubtitleID,
		URL:        subtitle.URL,
		TargetLang: subtitle.TargetLang,
		SourceLang: subtitle.SourceLang,
		Format:     subtitle.Format,
		FilePath:   subtitle.FilePath,
		Content:    content,
		FileSize:   subtitle.FileSize,
		CreatedAt:  subtitle.CreatedAt,
		UpdatedAt:  subtitle.UpdatedAt,
	}, nil
}

func (s *subtitleService) UpdateSubtitle(id uint, content string) (*models.SubtitleWithContent, error) {
	// Update content file and database
	if err := s.repo.UpdateContent(id, content); err != nil {
		return nil, err
	}

	// Return updated subtitle with content
	return s.GetSubtitleByID(id)
}

func (s *subtitleService) DeleteSubtitle(id uint) error {
	return s.repo.Delete(id)
}

func (s *subtitleService) generateSubtitleID(url, targetLang, format string) string {
	key := fmt.Sprintf("%s|%s|%s", url, targetLang, format)
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}
