package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"subtitle-translator/internal/models"

	"gorm.io/gorm"
)

const StorageDir = "storage/subtitles"

type SubtitleRepository interface {
	Create(subtitle *models.Subtitle, content string) error
	GetBySubtitleID(subtitleID string) (*models.Subtitle, error)
	GetAll(page, limit int, targetLang string) ([]models.Subtitle, int64, error)
	GetByID(id uint) (*models.Subtitle, error)
	Update(subtitle *models.Subtitle) error
	UpdateContent(id uint, content string) error
	Delete(id uint) error
}

type subtitleRepository struct {
	db *gorm.DB
}

func NewSubtitleRepository(db *gorm.DB) SubtitleRepository {
	// Create storage directory
	os.MkdirAll(StorageDir, 0755)
	return &subtitleRepository{db: db}
}

func (r *subtitleRepository) Create(subtitle *models.Subtitle, content string) error {
	// Save content to file
	if err := ioutil.WriteFile(subtitle.FilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save content file: %w", err)
	}

	// Save metadata to database
	return r.db.Create(subtitle).Error
}

func (r *subtitleRepository) GetBySubtitleID(subtitleID string) (*models.Subtitle, error) {
	var subtitle models.Subtitle
	err := r.db.Where("subtitle_id = ?", subtitleID).First(&subtitle).Error
	if err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (r *subtitleRepository) GetAll(page, limit int, targetLang string) ([]models.Subtitle, int64, error) {
	var subtitles []models.Subtitle
	var total int64

	query := r.db.Model(&models.Subtitle{})

	if targetLang != "" {
		query = query.Where("target_lang = ?", targetLang)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&subtitles).Error

	if err != nil {
		return nil, 0, err
	}

	return subtitles, total, nil
}

func (r *subtitleRepository) GetByID(id uint) (*models.Subtitle, error) {
	var subtitle models.Subtitle
	err := r.db.First(&subtitle, id).Error
	if err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (r *subtitleRepository) Update(subtitle *models.Subtitle) error {
	return r.db.Save(subtitle).Error
}

func (r *subtitleRepository) UpdateContent(id uint, content string) error {
	// Get subtitle to find file path
	subtitle, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Update file content
	if err := ioutil.WriteFile(subtitle.FilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to update content file: %w", err)
	}

	// Update file size in database
	subtitle.FileSize = int64(len(content))
	return r.db.Save(subtitle).Error
}

func (r *subtitleRepository) Delete(id uint) error {
	// Get subtitle to find file path
	subtitle, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(subtitle.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete content file: %w", err)
	}

	// Delete from database
	return r.db.Delete(&models.Subtitle{}, id).Error
}

// LoadContent loads subtitle content from file
func LoadContent(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read content file: %w", err)
	}
	return string(content), nil
}

// GenerateFilePath generates file path for subtitle
func GenerateFilePath(subtitleID string) string {
	filename := fmt.Sprintf("%s.vtt", subtitleID)
	return filepath.ToSlash(filepath.Join(StorageDir, filename))
}
