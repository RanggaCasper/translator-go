package translator

import "regexp"

var (
	fontTagRe      = regexp.MustCompile(`(?i)</?font\b[^>]*>`)
	emptyFontTagRe = regexp.MustCompile(`(?i)<font\b[^>]*>\s*</font>`)
)

// RemoveFontTags removes all <font ...> and </font> tags from subtitle content.
func RemoveFontTags(content string) string {
	// Run empty-tag cleanup first to avoid leaving extra noise around.
	cleaned := emptyFontTagRe.ReplaceAllString(content, "")
	return fontTagRe.ReplaceAllString(cleaned, "")
}

// RemoveEmptyFontTags removes empty <font ...></font> tags from subtitle content.
func RemoveEmptyFontTags(content string) string {
	return RemoveFontTags(content)
}
