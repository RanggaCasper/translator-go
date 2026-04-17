package translator

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	idSpaceBeforePunctRe = regexp.MustCompile(`\s+([,.;:!?])`)
	idSpaceBeforeQuoteRe = regexp.MustCompile(`\s+(["”’])`)
	idQuotePeriodEndRe   = regexp.MustCompile(`\s+(["”’])\.$`)
	idManySpacesRe       = regexp.MustCompile(`[ \t]{2,}`)
	idEllipsisVariantRe  = regexp.MustCompile(`\.\s*\.\s*\.?`)
	idJointPunctRe       = regexp.MustCompile(`([!?])\s+([!?])`)
	idAfterPunctSpaceRe  = regexp.MustCompile(`([,.;:!?])([^\s,.;:!?])`)
	idBrokenSuffixRe     = regexp.MustCompile(`(?i)\b([a-z]+)\s*-\s*(ku|mu|nya|lah|kah|pun)\b`)
	idStutterPrefixRe    = regexp.MustCompile(`^([A-Za-z]{1,3})\s*\.\.\.\s*(.+)$`)
	idSpeakerPrefixRe    = regexp.MustCompile(`^([A-Za-z]{1,3})-\s*([A-Za-z].*)$`)
	idSpeakerWordRe      = regexp.MustCompile(`^[A-Za-z]{1,3}-[A-Za-z].*$`)
	idNoiseOnlyRe        = regexp.MustCompile(`(?i)^(aturan|komposisi grup|keterangan|catatan|instruksi)\b.*$`)
	idSentenceSplitRe    = regexp.MustCompile(`([.!?])\s+`)
)

// EnhanceIndonesianSubtitle polishes machine translation output to fit natural Indonesian subtitle style.
func EnhanceIndonesianSubtitle(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}

	t := strings.TrimSpace(text)
	t = strings.ReplaceAll(t, "…", "...")
	t = idEllipsisVariantRe.ReplaceAllString(t, "...")
	t = idBrokenSuffixRe.ReplaceAllString(t, `$1-$2`)
	t = idSpaceBeforePunctRe.ReplaceAllString(t, `$1`)
	t = idSpaceBeforeQuoteRe.ReplaceAllString(t, `$1`)
	t = idJointPunctRe.ReplaceAllString(t, `$1$2`)
	t = idAfterPunctSpaceRe.ReplaceAllString(t, `$1 $2`)
	t = idManySpacesRe.ReplaceAllString(t, " ")

	t = fixStutteringPrefix(t)
	t = fixSpeakerPrefix(t)
	t = fixSpeakerPrefixTokens(t)
	t = upgradeWordChoiceID(t)
	t = normalizeSentenceCapitalization(t)
	t = idQuotePeriodEndRe.ReplaceAllString(t, `.$1`)
	t = idSpaceBeforeQuoteRe.ReplaceAllString(t, `$1`)

	if idNoiseOnlyRe.MatchString(strings.ToLower(strings.TrimSpace(t))) {
		return ""
	}

	return strings.TrimSpace(t)
}

func fixStutteringPrefix(s string) string {
	m := idStutterPrefixRe.FindStringSubmatch(s)
	if len(m) != 3 {
		return s
	}

	firstWord := firstWordOnly(m[2])
	if firstWord == "" {
		return s
	}

	return stutterPrefixFromWord(firstWord, len([]rune(m[1]))) + "... " + capitalizeFirstLetter(strings.TrimSpace(m[2]))
}

func fixSpeakerPrefix(s string) string {
	m := idSpeakerPrefixRe.FindStringSubmatch(strings.TrimSpace(s))
	if len(m) != 3 {
		return s
	}

	word := firstWordOnly(m[2])
	if word == "" {
		return s
	}

	return stutterPrefixFromWord(word, len([]rune(m[1]))) + "-" + strings.TrimLeft(m[2], " ")
}

func fixSpeakerPrefixTokens(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	for i, w := range words {
		trimmed := strings.Trim(w, "\"'“”‘’.,!?;:()[]{}")
		if !idSpeakerWordRe.MatchString(trimmed) {
			continue
		}

		parts := strings.SplitN(trimmed, "-", 2)
		if len(parts) != 2 {
			continue
		}

		prefixLen := len([]rune(parts[0]))
		nextWord := firstWordOnly(parts[1])
		if nextWord == "" {
			continue
		}

		replacement := stutterPrefixFromWord(nextWord, prefixLen) + "-" + parts[1]
		words[i] = strings.Replace(words[i], trimmed, replacement, 1)
	}

	return strings.Join(words, " ")
}

func firstWordOnly(s string) string {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) == 0 {
		return ""
	}
	return strings.Trim(parts[0], "\"'“”‘’.,!?;:()[]{}")
}

func upgradeWordChoiceID(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	for i, w := range words {
		trimmed := strings.Trim(w, "\"'“”‘’.,!?;:()[]{}")
		if strings.EqualFold(trimmed, "buat") {
			if i+1 < len(words) {
				next := strings.Trim(strings.ToLower(words[i+1]), "\"'“”‘’.,!?;:()[]{}")
				if next == "apa" {
					continue
				}
			}

			replacement := "untuk"
			if len(trimmed) > 0 {
				runes := []rune(trimmed)
				if unicode.IsUpper(runes[0]) {
					replacement = "Untuk"
				}
			}
			words[i] = strings.Replace(words[i], trimmed, replacement, 1)
		}
	}

	return strings.Join(words, " ")
}

func normalizeSentenceCapitalization(s string) string {
	if s == "" {
		return s
	}

	parts := idSentenceSplitRe.Split(s, -1)
	delims := idSentenceSplitRe.FindAllStringSubmatch(s, -1)

	for i := range parts {
		parts[i] = capitalizeFirstLetter(parts[i])
	}

	if len(delims) == 0 {
		return parts[0]
	}

	var b strings.Builder
	for i := 0; i < len(delims) && i < len(parts); i++ {
		b.WriteString(parts[i])
		b.WriteString(strings.TrimSpace(delims[i][1]))
		b.WriteString(" ")
	}
	if len(parts) > len(delims) {
		b.WriteString(parts[len(parts)-1])
	}

	return strings.TrimSpace(idManySpacesRe.ReplaceAllString(b.String(), " "))
}

func capitalizeFirstLetter(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsLetter(r) {
			runes[i] = unicode.ToUpper(r)
			break
		}
	}
	return string(runes)
}

func stutterPrefixFromWord(word string, prefixLen int) string {
	if prefixLen <= 0 {
		return word
	}

	runes := []rune(word)
	if len(runes) == 0 {
		return word
	}

	if prefixLen > len(runes) {
		prefixLen = len(runes)
	}

	selected := make([]rune, prefixLen)
	for i := 0; i < prefixLen; i++ {
		if i == 0 {
			selected[i] = unicode.ToUpper(runes[i])
			continue
		}
		selected[i] = unicode.ToLower(runes[i])
	}

	return string(selected)
}
