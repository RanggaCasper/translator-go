package translator

import (
	"log"
	"regexp"
	"strings"
)

const (
	maxCueTextLines  = 3
	maxCueWords      = 25
	maxOutputLines   = 2
	softLineChars    = 35
	hardLineChars    = 40
	softLineWords    = 6
	hardLineWords    = 7
	minLeadLineChars = 7
)

var (
	vttTimestampRe = regexp.MustCompile(`^(?P<s>(?:\d{1,2}:)?\d{2}:\d{2}[.,]\d{3})\s*-->\s*(?P<e>(?:\d{1,2}:)?\d{2}:\d{2}[.,]\d{3})(?P<rest>.*)$`)
	vttTagRe       = regexp.MustCompile(`<[^>]+>`)
)

type vttCueBatch struct {
	textLineIndices []int
	originalText    string
}

// TranslateVTT parses VTT subtitle, translates per-timestamp cue text, and returns translated VTT content.
func TranslateVTT(content, targetLang, sourceLang string) (string, error) {
	lines := strings.Split(content, "\n")
	blockedLines := markLongCueBlocks(lines)
	cues := collectVTTCueBatches(lines, blockedLines)
	if len(cues) == 0 {
		return strings.Join(lines, "\n"), nil
	}

	textValues := make([]string, 0, len(cues))
	for _, cue := range cues {
		textValues = append(textValues, cue.originalText)
	}

	log.Printf("Starting translation of %d cue blocks...", len(textValues))

	// Translate all cue text blocks.
	translated, err := BatchTranslate(textValues, targetLang, sourceLang)
	if err != nil {
		return "", err
	}

	log.Printf("Translation completed successfully")

	// Replace translated cue text back.
	for idx, trans := range translated {
		applyTranslatedCue(lines, cues[idx], trans, targetLang)
	}

	return strings.Join(lines, "\n"), nil
}

func collectVTTCueBatches(lines []string, blockedLines map[int]bool) []vttCueBatch {
	for i := range lines {
		if blockedLines[i] {
			lines[i] = ""
			continue
		}

		line := RemoveFontTags(lines[i])
		if vttTimestampRe.MatchString(strings.TrimSpace(line)) {
			line = normalizeTimestampLine(line)
		}
		lines[i] = line
	}

	cues := make([]vttCueBatch, 0, 32)
	start := 0
	for start < len(lines) {
		for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
			start++
		}
		if start >= len(lines) {
			break
		}

		end := start
		for end < len(lines) && strings.TrimSpace(lines[end]) != "" {
			end++
		}

		cue, ok := buildCueBatch(lines, start, end)
		if ok {
			cues = append(cues, cue)
		}

		start = end
	}

	return cues
}

func buildCueBatch(lines []string, start, end int) (vttCueBatch, bool) {
	timestampLine := -1
	for i := start; i < end; i++ {
		if vttTimestampRe.MatchString(strings.TrimSpace(lines[i])) {
			timestampLine = i
			break
		}
	}

	if timestampLine == -1 {
		return vttCueBatch{}, false
	}

	textLineIndices := make([]int, 0, 4)
	textParts := make([]string, 0, 4)
	for i := timestampLine + 1; i < end; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || isDigitOnly(trimmed) {
			continue
		}

		clean := strings.TrimSpace(vttTagRe.ReplaceAllString(lines[i], ""))
		if clean == "" {
			continue
		}

		textLineIndices = append(textLineIndices, i)
		textParts = append(textParts, clean)
	}

	if len(textParts) == 0 {
		return vttCueBatch{}, false
	}

	return vttCueBatch{
		textLineIndices: textLineIndices,
		// Send cue text as a single sentence for better translation quality.
		originalText: strings.Join(textParts, " "),
	}, true
}

func applyTranslatedCue(lines []string, cue vttCueBatch, translated, targetLang string) {
	translatedLines := splitCueTextLines(translated, targetLang)
	translatedLines = capCueOutputLines(translatedLines, maxOutputLines)
	if len(translatedLines) == 0 {
		if strings.ToLower(targetLang) != "id" {
			translatedLines = splitCueTextLines(cue.originalText, targetLang)
			translatedLines = capCueOutputLines(translatedLines, maxOutputLines)
		}
	}

	if len(translatedLines) == 0 {
		for _, lineIdx := range cue.textLineIndices {
			lines[lineIdx] = ""
		}
		return
	}

	lineSlots := len(cue.textLineIndices)
	if lineSlots == 0 {
		return
	}

	if len(translatedLines) <= lineSlots {
		for i, lineIdx := range cue.textLineIndices {
			if i < len(translatedLines) {
				lines[lineIdx] = translatedLines[i]
				continue
			}
			lines[lineIdx] = ""
		}
		return
	}

	for i := 0; i < lineSlots-1; i++ {
		lines[cue.textLineIndices[i]] = translatedLines[i]
	}
	lines[cue.textLineIndices[lineSlots-1]] = strings.Join(translatedLines[lineSlots-1:], "\n")
}

func splitCueTextLines(text string, targetLang string) []string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	rawLines := strings.Split(normalized, "\n")

	result := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		fixed := strings.TrimSpace(RemoveFontTags(SingleLine(line)))
		if fixed == "" {
			continue
		}
		result = append(result, fixed)
	}

	if len(result) == 0 {
		fallback := strings.TrimSpace(RemoveFontTags(SingleLine(normalized)))
		if fallback != "" {
			result = append(result, fallback)
		}
	}

	if strings.EqualFold(strings.TrimSpace(targetLang), "id") {
		joined := strings.TrimSpace(strings.Join(result, " "))
		if joined == "" {
			return result
		}

		wrapped := wrapCueLineByHeuristics(joined)
		wrapped = rebalanceShortLeadLine(wrapped)

		return capCueOutputLines(wrapped, maxOutputLines)
	}

	return result
}

func capCueOutputLines(lines []string, maxLines int) []string {
	if maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}

	trimmed := make([]string, 0, maxLines)
	for i := 0; i < maxLines-1; i++ {
		trimmed = append(trimmed, lines[i])
	}
	trimmed = append(trimmed, strings.Join(lines[maxLines-1:], " "))

	return trimmed
}

func wrapCueLineByHeuristics(text string) []string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return nil
	}

	lines := make([]string, 0, 4)
	current := make([]string, 0, hardLineWords)
	currentLen := 0

	flush := func() {
		if len(current) == 0 {
			return
		}
		lines = append(lines, strings.Join(current, " "))
		current = current[:0]
		currentLen = 0
	}

	for i, word := range words {
		wordLen := len([]rune(word))
		addedLen := wordLen
		if len(current) > 0 {
			addedLen++
		}

		if len(current) > 0 && (currentLen+addedLen > hardLineChars || len(current)+1 > hardLineWords) {
			flush()
			addedLen = wordLen
		}

		current = append(current, word)
		currentLen += addedLen

		wordCount := len(current)
		lastWord := current[wordCount-1]
		remaining := len(words) - (i + 1)

		if strings.HasSuffix(lastWord, ".") || strings.HasSuffix(lastWord, "!") || strings.HasSuffix(lastWord, "?") {
			if currentLen >= minLeadLineChars || remaining == 0 {
				flush()
			}
			continue
		}

		if currentLen >= hardLineChars || wordCount >= hardLineWords {
			flush()
			continue
		}

		if currentLen >= softLineChars && wordCount >= softLineWords {
			flush()
			continue
		}

		if currentLen >= softLineChars && (strings.HasSuffix(lastWord, ",") || strings.HasSuffix(lastWord, ";") || strings.HasSuffix(lastWord, ":")) {
			flush()
			continue
		}

		if currentLen >= 32 && remaining >= 2 && i+1 < len(words) && isNaturalPauseConjunction(words[i+1]) {
			flush()
		}
	}

	flush()
	return lines
}

func rebalanceShortLeadLine(lines []string) []string {
	if len(lines) < 2 {
		return lines
	}

	first := strings.TrimSpace(lines[0])
	second := strings.TrimSpace(lines[1])
	if first == "" || second == "" {
		return lines
	}

	if len([]rune(first)) >= minLeadLineChars {
		return lines
	}

	firstWords := strings.Fields(first)
	secondWords := strings.Fields(second)
	if len(secondWords) == 0 {
		return lines
	}

	for len(secondWords) > 0 {
		candidateWords := append(append([]string{}, firstWords...), secondWords[0])
		candidateLine := strings.Join(candidateWords, " ")

		if len([]rune(candidateLine)) > hardLineChars || len(candidateWords) > hardLineWords {
			break
		}

		firstWords = candidateWords
		secondWords = secondWords[1:]

		if len([]rune(candidateLine)) >= minLeadLineChars {
			break
		}
	}

	balanced := make([]string, 0, 2)
	if len(firstWords) > 0 {
		balanced = append(balanced, strings.Join(firstWords, " "))
	}
	if len(secondWords) > 0 {
		balanced = append(balanced, strings.Join(secondWords, " "))
	}

	if len(balanced) == 0 {
		return lines
	}

	return balanced
}

func isNaturalPauseConjunction(word string) bool {
	w := strings.ToLower(strings.Trim(word, "\"'“”‘’.,!?;:()[]{}"))
	switch w {
	case "dan", "atau", "tapi", "tetapi", "namun", "karena", "agar", "supaya", "sehingga", "lalu", "kemudian", "sedangkan", "sementara":
		return true
	default:
		return false
	}
}

func markLongCueBlocks(lines []string) map[int]bool {
	blocked := make(map[int]bool)
	start := 0

	for start < len(lines) {
		for start < len(lines) && strings.TrimSpace(RemoveFontTags(lines[start])) == "" {
			start++
		}
		if start >= len(lines) {
			break
		}

		end := start
		for end < len(lines) && strings.TrimSpace(RemoveFontTags(lines[end])) != "" {
			end++
		}

		if shouldDropVTTBlock(lines[start:end]) {
			for i := start; i < end; i++ {
				blocked[i] = true
			}
		}

		start = end
	}

	return blocked
}

func shouldDropVTTBlock(block []string) bool {
	timestampIdx := -1
	for i, line := range block {
		if vttTimestampRe.MatchString(strings.TrimSpace(RemoveFontTags(line))) {
			timestampIdx = i
			break
		}
	}

	if timestampIdx == -1 {
		return false
	}

	textLines := 0
	wordCount := 0
	for i := timestampIdx + 1; i < len(block); i++ {
		trimmed := strings.TrimSpace(RemoveFontTags(block[i]))
		if trimmed == "" || isDigitOnly(trimmed) || isStandalonePunctuationLine(trimmed) {
			continue
		}
		textLines++
		wordCount += len(strings.Fields(trimmed))
	}

	return textLines > maxCueTextLines || wordCount > maxCueWords
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
