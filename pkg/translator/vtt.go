package translator

import (
	"log"
	"regexp"
	"strings"
)

var (
	vttTimestampRe = regexp.MustCompile(`^(?P<s>(?:\d{1,2}:)?\d{2}:\d{2}[.,]\d{3})\s*-->\s*(?P<e>(?:\d{1,2}:)?\d{2}:\d{2}[.,]\d{3})(?P<rest>.*)$`)
	vttTagRe       = regexp.MustCompile(`<[^>]+>`)
	leadingTagsRe  = regexp.MustCompile(`^(?:<[^>]+>)+`)
	trailingTagsRe = regexp.MustCompile(`(?:<[^>]+>)+$`)
)

// TranslateVTT parses VTT subtitle, translates text lines, and returns translated VTT content
func TranslateVTT(content, targetLang, sourceLang string) (string, error) {
	lines := strings.Split(content, "\n")

	var textIndices []int
	var textValues []string

	for i, line := range lines {
		// Remove font tags that should not appear in output
		line = RemoveFontTags(line)
		lines[i] = line

		// Normalize timestamp lines
		if vttTimestampRe.MatchString(strings.TrimSpace(line)) {
			lines[i] = normalizeTimestampLine(line)
			continue
		}

		// Skip metadata lines
		if strings.HasPrefix(line, "WEBVTT") ||
			strings.HasPrefix(line, "NOTE") ||
			strings.HasPrefix(line, "STYLE") ||
			strings.TrimSpace(line) == "" ||
			isDigitOnly(strings.TrimSpace(line)) {
			continue
		}

		// Extract text without tags for translation
		clean := vttTagRe.ReplaceAllString(line, "")
		clean = strings.TrimSpace(clean)

		if clean != "" {
			textIndices = append(textIndices, i)
			textValues = append(textValues, clean)
		}
	}

	if len(textValues) == 0 {
		return strings.Join(lines, "\n"), nil
	}

	log.Printf("Starting translation of %d text lines...", len(textValues))

	// Translate all text
	translated, err := BatchTranslate(textValues, targetLang, sourceLang)
	if err != nil {
		return "", err
	}

	log.Printf("Translation completed successfully")

	// Replace translated text back
	for idx, trans := range translated {
		lineIdx := textIndices[idx]
		original := lines[lineIdx]

		// Preserve leading/trailing tags
		leadingMatch := leadingTagsRe.FindString(original)
		trailingMatch := trailingTagsRe.FindString(original)

		leadingTags := ""
		if leadingMatch != "" {
			leadingTags = RemoveFontTags(leadingMatch)
		}

		trailingTags := ""
		if trailingMatch != "" {
			trailingTags = RemoveFontTags(trailingMatch)
		}

		// Ensure translated text is single line
		transFixed := SingleLine(trans)
		transFixed = strings.TrimSpace(RemoveFontTags(transFixed))
		if transFixed == "" {
			// Fallback to source text if translation comes back empty
			transFixed = strings.TrimSpace(vttTagRe.ReplaceAllString(original, ""))
		}

		lines[lineIdx] = RemoveFontTags(leadingTags + transFixed + trailingTags)
	}

	return strings.Join(lines, "\n"), nil
}

func normalizeTimestampLine(line string) string {
	match := vttTimestampRe.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return line
	}

	s := toVTTTime(match[1])
	e := toVTTTime(match[2])
	rest := ""
	if len(match) > 3 {
		rest = match[3]
	}

	return s + " --> " + e + rest
}

func toVTTTime(t string) string {
	t = strings.TrimSpace(strings.ReplaceAll(t, ",", "."))

	// MM:SS.mmm -> 00:MM:SS.mmm
	if regexp.MustCompile(`^\d{2}:\d{2}\.\d{3}$`).MatchString(t) {
		t = "00:" + t
	}

	// H:MM:SS.mmm -> HH:MM:SS.mmm
	if regexp.MustCompile(`^\d{1}:\d{2}:\d{2}\.\d{3}$`).MatchString(t) {
		t = "0" + t
	}

	return t
}

func isDigitOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
