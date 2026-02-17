package translator

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	assDialogueRe = regexp.MustCompile(`^(Dialogue:\s*\d+,\d[^,]*,\d[^,]*,)[^,]*,[^,]*,[^,]*,[^,]*,[^,]*,(.*)$`)
	assBraceRe    = regexp.MustCompile(`\{[^}]*\}`)
)

type assDialogue struct {
	start string
	end   string
	text  string
}

// TranslateASSToVTT parses ASS subtitle, translates dialogue, and outputs as VTT
func TranslateASSToVTT(content, targetLang, sourceLang string) (string, error) {
	lines := strings.Split(content, "\n")
	var dialogues []assDialogue

	for _, line := range lines {
		match := assDialogueRe.FindStringSubmatch(strings.TrimSpace(line))
		if match != nil {
			prefix := match[1]
			text := match[2]

			// Extract timestamps from prefix
			parts := strings.Split(prefix, ",")
			if len(parts) >= 3 {
				start := strings.TrimSpace(parts[1])
				end := strings.TrimSpace(parts[2])

				// Clean ASS formatting
				cleanText := assBraceRe.ReplaceAllString(text, "")
				cleanText = strings.ReplaceAll(cleanText, "\\N", "\n")
				cleanText = strings.ReplaceAll(cleanText, "\\n", "\n")
				cleanText = strings.TrimSpace(cleanText)

				if cleanText != "" {
					dialogues = append(dialogues, assDialogue{
						start: start,
						end:   end,
						text:  cleanText,
					})
				}
			}
		}
	}

	if len(dialogues) == 0 {
		return "WEBVTT\n\n", nil
	}

	log.Printf("Starting translation of %d ASS dialogue lines...", len(dialogues))

	// Translate all dialogue text
	var texts []string
	for _, d := range dialogues {
		texts = append(texts, d.text)
	}

	translated, err := BatchTranslate(texts, targetLang, sourceLang)
	if err != nil {
		return "", err
	}

	log.Printf("ASS translation completed successfully")

	// Build VTT
	var vttLines []string
	vttLines = append(vttLines, "WEBVTT", "")

	for i, d := range dialogues {
		vttStart := assTimeToVTT(d.start)
		vttEnd := assTimeToVTT(d.end)

		vttLines = append(vttLines, strconv.Itoa(i+1))
		vttLines = append(vttLines, fmt.Sprintf("%s --> %s", vttStart, vttEnd))
		vttLines = append(vttLines, translated[i])
		vttLines = append(vttLines, "")
	}

	return strings.Join(vttLines, "\n"), nil
}

// assTimeToVTT converts ASS time (H:MM:SS.cc) to VTT time (HH:MM:SS.mmm)
func assTimeToVTT(t string) string {
	parts := strings.Split(t, ":")
	if len(parts) != 3 {
		return t
	}

	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return t
	}

	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return t
	}

	secParts := strings.Split(parts[2], ".")
	s, err := strconv.Atoi(secParts[0])
	if err != nil {
		return t
	}

	cs := 0
	if len(secParts) > 1 {
		cs, _ = strconv.Atoi(secParts[1])
	}

	ms := cs * 10

	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}
