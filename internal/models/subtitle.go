package models

import (
	"time"

	"gorm.io/gorm"
)

// Subtitle represents subtitle metadata in database with file path
type Subtitle struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SubtitleID string         `gorm:"uniqueIndex;size:32;not null" json:"subtitle_id"`
	URL        string         `gorm:"type:text;not null" json:"url"`
	TargetLang string         `gorm:"size:10;not null;index" json:"target_lang"`
	SourceLang string         `gorm:"size:10;not null" json:"source_lang"`
	Format     string         `gorm:"size:10;not null" json:"format"`
	FilePath   string         `gorm:"type:varchar(500);not null" json:"file_path"` // Path to VTT file
	FileSize   int64          `gorm:"not null" json:"file_size"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (Subtitle) TableName() string {
	return "subtitles"
}

// SubtitleWithContent represents subtitle with loaded content
type SubtitleWithContent struct {
	ID         uint      `json:"id"`
	SubtitleID string    `json:"subtitle_id"`
	URL        string    `json:"url"`
	TargetLang string    `json:"target_lang"`
	SourceLang string    `json:"source_lang"`
	Format     string    `json:"format"`
	FilePath   string    `json:"file_path"`
	Content    string    `json:"content"` // Loaded from file
	FileSize   int64     `json:"file_size"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
