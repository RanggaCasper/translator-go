package translator

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var englishSignalRe = regexp.MustCompile(`(?i)\b(i|you|he|she|we|they|it|if|and|or|but|so|need|needed|willing|hand|help|know)\b|[a-z]+'[a-z]+`)
var englishWordTokenRe = regexp.MustCompile(`(?i)[a-z]+(?:'[a-z]+)?`)
var formattingTagRe = regexp.MustCompile(`(?i)<\s*(/?)\s*(i|b|u|em|strong)\s*>`)
var brokenFormattingTagRe = regexp.MustCompile(`(?i)(^|[\s"'“”‘’(\[])(/?)(i|b|u|em|strong)>`)
var spaceBeforeCloseFormattingTagRe = regexp.MustCompile(`(?i)\s+(</\s*(i|b|u|em|strong)\s*>)`)
var spaceAfterOpenFormattingTagRe = regexp.MustCompile(`(?i)(<\s*(i|b|u|em|strong)\s*>)\s+`)

var englishStopwords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "this": {}, "that": {}, "these": {}, "those": {},
	"to": {}, "of": {}, "for": {}, "from": {}, "with": {}, "in": {}, "on": {}, "at": {}, "by": {}, "over": {}, "under": {},
	"is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "been": {}, "being": {},
	"do": {}, "does": {}, "did": {}, "may": {}, "can": {}, "could": {}, "will": {}, "would": {}, "should": {}, "shall": {},
	"it": {}, "you": {}, "we": {}, "they": {}, "he": {}, "she": {}, "i": {},
	"and": {}, "or": {}, "but": {}, "if": {}, "then": {}, "than": {}, "as": {},
}

var indonesianStopwords = map[string]struct{}{
	"yang": {}, "dan": {}, "atau": {}, "tapi": {}, "tetapi": {}, "namun": {},
	"ini": {}, "itu": {}, "di": {}, "ke": {}, "dari": {}, "untuk": {}, "dengan": {},
	"karena": {}, "agar": {}, "supaya": {}, "sehingga": {}, "jika": {}, "kalau": {},
	"adalah": {}, "akan": {}, "tidak": {}, "bukan": {}, "sudah": {}, "belum": {},
	"aku": {}, "saya": {}, "kamu": {}, "dia": {}, "mereka": {}, "kami": {}, "kita": {},
	"tak": {},
}

// PostProcessSubtitleContent normalizes stored subtitle content before returning it to clients.
func PostProcessSubtitleContent(content, targetLang string) string {
	cleaned := RemoveFontTags(content)
	if strings.ToLower(strings.TrimSpace(targetLang)) != "id" {
		return cleaned
	}

	lines := strings.Split(cleaned, "\n")
	processed := make([]string, 0, len(lines))
	block := make([]string, 0, 8)

	flushBlock := func() {
		if len(block) == 0 {
			return
		}

		processedBlock := processSubtitleBlock(block)
		if len(processedBlock) > 0 {
			if len(processed) > 0 {
				processed = append(processed, "")
			}
			processed = append(processed, processedBlock...)
		}
		block = block[:0]
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flushBlock()
			continue
		}

		block = append(block, line)
	}

	flushBlock()

	return strings.Join(processed, "\n")
}

func isSubtitleMetadataLine(line string) bool {
	return strings.HasPrefix(line, "WEBVTT") ||
		strings.HasPrefix(line, "NOTE") ||
		strings.HasPrefix(line, "STYLE") ||
		vttTimestampRe.MatchString(line) ||
		isDigitOnly(line)
}

func processSubtitleBlock(block []string) []string {
	if len(block) == 0 {
		return nil
	}

	if shouldDropSubtitleBlock(block) {
		return nil
	}

	processed := make([]string, 0, len(block))
	for _, line := range block {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if isSubtitleMetadataLine(trimmed) {
			processed = append(processed, trimmed)
			continue
		}

		normalized := SingleLine(trimmed)
		normalized = ensureIndonesianLine(normalized)
		fixed := EnhanceIndonesianSubtitle(normalized)
		if fixed == "" {
			continue
		}
		fixed = normalizeFormattingTags(fixed)

		if isStandalonePunctuationLine(fixed) {
			if len(processed) == 0 {
				continue
			}

			last := len(processed) - 1
			if processed[last] == "" || isSubtitleMetadataLine(processed[last]) {
				continue
			}

			if shouldAppendStandalonePunctuation(fixed) {
				processed[last] = strings.TrimRight(processed[last], " ") + fixed
			}
			continue
		}

		fixed = stripLeadingPunctuationPrefix(fixed)
		processed = append(processed, fixed)
	}

	return processed
}

func shouldDropSubtitleBlock(block []string) bool {
	return shouldDropVTTBlock(block)
}

func isStandalonePunctuationLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return false
		}
	}

	return true
}

func shouldAppendStandalonePunctuation(line string) bool {
	return line == "!" || line == "?" || line == "?!" || line == "!?"
}

func stripLeadingPunctuationPrefix(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "<") {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "- [") || strings.HasPrefix(trimmed, "-[") {
		return trimmed
	}
	trimmed = strings.TrimLeft(trimmed, ".,!?/+-=(){}|\\'\";:~*_`")
	trimmed = strings.TrimLeft(trimmed, " ")
	return strings.TrimSpace(trimmed)
}

func ensureIndonesianLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !looksUntranslatedEnglish(trimmed) {
		return trimmed
	}

	masked, tags := maskFormattingTags(trimmed)

	translated, err := GoogleTranslate(masked, "id", "auto")
	if err != nil {
		return trimmed
	}

	translated = strings.TrimSpace(translated)
	if translated == "" {
		return trimmed
	}

	translated = unmaskFormattingTags(translated, tags)
	translated = normalizeFormattingTags(translated)

	return translated
}

func looksUntranslatedEnglish(line string) bool {
	matches := englishSignalRe.FindAllString(line, -1)
	if len(matches) >= 2 {
		return true
	}

	tokens := englishWordTokenRe.FindAllString(strings.ToLower(line), -1)
	if len(tokens) < 3 {
		return false
	}

	englishHits := 0
	indonesianHits := 0
	for _, token := range tokens {
		if _, ok := englishStopwords[token]; ok {
			englishHits++
		}
		if _, ok := indonesianStopwords[token]; ok {
			indonesianHits++
		}
	}

	if englishHits >= 3 && englishHits > indonesianHits {
		return true
	}

	return false
}

func normalizeFormattingTags(line string) string {
	fixed := brokenFormattingTagRe.ReplaceAllString(line, `${1}<${2}${3}>`)
	fixed = formattingTagRe.ReplaceAllStringFunc(fixed, func(tag string) string {
		parts := formattingTagRe.FindStringSubmatch(tag)
		if len(parts) != 3 {
			return tag
		}
		return "<" + parts[1] + strings.ToLower(parts[2]) + ">"
	})
	fixed = spaceBeforeCloseFormattingTagRe.ReplaceAllString(fixed, `$1`)
	fixed = spaceAfterOpenFormattingTagRe.ReplaceAllString(fixed, `$1`)
	return fixed
}

func maskFormattingTags(line string) (string, map[string]string) {
	tags := map[string]string{}
	idx := 0
	masked := formattingTagRe.ReplaceAllStringFunc(line, func(tag string) string {
		key := "__RANIME_TAG_" + strconv.Itoa(idx) + "__"
		tags[key] = tag
		idx++
		return key
	})
	return masked, tags
}

func unmaskFormattingTags(line string, tags map[string]string) string {
	out := line
	for key, tag := range tags {
		out = strings.ReplaceAll(out, key, tag)
	}
	return out
}
