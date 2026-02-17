package translator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

const (
	chunkSize          = 80
	maxChunkChars      = 1800
	maxSingleTextChars = 1400
)

var horizontalWhitespaceRe = regexp.MustCompile(`[ \t]+`)

type indexedText struct {
	index int
	text  string
}

// BatchTranslate translates multiple texts in batches with concurrent processing
func BatchTranslate(texts []string, targetLang, sourceLang string) ([]string, error) {
	if len(texts) == 0 {
		return texts, nil
	}

	// Filter non-empty texts
	var nonEmpty []indexedText
	for i, t := range texts {
		if strings.TrimSpace(t) != "" {
			nonEmpty = append(nonEmpty, indexedText{index: i, text: t})
		}
	}

	if len(nonEmpty) == 0 {
		return texts, nil
	}

	result := make([]string, len(texts))
	copy(result, texts)

	// Create chunks
	chunks := buildChunks(nonEmpty)

	// Process chunks concurrently
	var wg sync.WaitGroup
	resultMutex := &sync.Mutex{}

	for _, chunk := range chunks {
		wg.Add(1)
		go func(chunk []indexedText) {
			defer wg.Done()
			processChunk(chunk, targetLang, sourceLang, result, resultMutex)
		}(chunk)
	}

	wg.Wait()
	return result, nil
}

func processChunk(chunk []indexedText, targetLang, sourceLang string, result []string, mutex *sync.Mutex) {
	// Try batch translation first
	success := tryBatchTranslate(chunk, targetLang, sourceLang, result, mutex)

	if !success {
		// If batch fails, try smaller batches (divide by 2)
		if len(chunk) > 1 {
			mid := len(chunk) / 2
			processChunk(chunk[:mid], targetLang, sourceLang, result, mutex)
			processChunk(chunk[mid:], targetLang, sourceLang, result, mutex)
		} else {
			// Last fallback for a single line
			translateConcurrent(chunk, targetLang, sourceLang, result, mutex)
		}
	}
}

func tryBatchTranslate(chunk []indexedText, targetLang, sourceLang string, result []string, mutex *sync.Mutex) bool {
	// Generate unique separator
	tokenBytes := make([]byte, 8)
	rand.Read(tokenBytes)
	token := fmt.Sprintf("RANIMESEP%sRANIMESEP", hex.EncodeToString(tokenBytes))
	separator := fmt.Sprintf(" %s ", token)

	// Combine texts
	var combined strings.Builder
	for i, item := range chunk {
		if i > 0 {
			combined.WriteString(separator)
		}
		combined.WriteString(item.text)
	}

	// Translate
	translated, err := GoogleTranslate(combined.String(), targetLang, sourceLang)
	if err != nil {
		log.Printf("Batch translation failed for chunk of %d items: %v", len(chunk), err)
		return false
	}

	// Split results
	parts := strings.Split(translated, token)

	if len(parts) != len(chunk) {
		log.Printf("Batch split mismatch (got %d parts for %d lines)", len(parts), len(chunk))
		return false
	}

	// Store results
	mutex.Lock()
	for i, item := range chunk {
		// Clean up spaces
		cleaned := horizontalWhitespaceRe.ReplaceAllString(parts[i], " ")
		result[item.index] = strings.TrimSpace(cleaned)
	}
	mutex.Unlock()

	return true
}

func translateConcurrent(chunk []indexedText, targetLang, sourceLang string, result []string, mutex *sync.Mutex) {
	var wg sync.WaitGroup

	for _, item := range chunk {
		wg.Add(1)
		go func(item indexedText) {
			defer wg.Done()

			trans, err := translateLongText(item.text, targetLang, sourceLang)
			if err != nil {
				log.Printf("Individual translation failed: %v", err)
				return
			}

			mutex.Lock()
			result[item.index] = strings.TrimSpace(trans)
			mutex.Unlock()
		}(item)
	}

	wg.Wait()
}

func buildChunks(items []indexedText) [][]indexedText {
	var chunks [][]indexedText
	current := make([]indexedText, 0, chunkSize)
	currentChars := 0

	for _, item := range items {
		itemChars := len([]rune(item.text))

		shouldFlush := len(current) > 0 && (len(current) >= chunkSize || currentChars+itemChars > maxChunkChars)
		if shouldFlush {
			chunks = append(chunks, current)
			current = make([]indexedText, 0, chunkSize)
			currentChars = 0
		}

		current = append(current, item)
		currentChars += itemChars
	}

	if len(current) > 0 {
		chunks = append(chunks, current)
	}

	return chunks
}

func translateLongText(text, targetLang, sourceLang string) (string, error) {
	if len([]rune(text)) <= maxSingleTextChars {
		return GoogleTranslate(text, targetLang, sourceLang)
	}

	parts := splitTextByLength(text, maxSingleTextChars)
	translatedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		translated, err := GoogleTranslate(part, targetLang, sourceLang)
		if err != nil {
			return "", err
		}
		translatedParts = append(translatedParts, translated)
	}

	return strings.Join(translatedParts, " "), nil
}

func splitTextByLength(text string, maxLen int) []string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return []string{text}
	}

	var result []string
	start := 0
	for start < len(runes) {
		end := start + maxLen
		if end >= len(runes) {
			part := strings.TrimSpace(string(runes[start:]))
			if part != "" {
				result = append(result, part)
			}
			break
		}

		splitAt := end
		for i := end; i > start+maxLen/2; i-- {
			if unicode.IsSpace(runes[i-1]) {
				splitAt = i
				break
			}
		}

		part := strings.TrimSpace(string(runes[start:splitAt]))
		if part != "" {
			result = append(result, part)
		}
		start = splitAt
	}

	if len(result) == 0 {
		return []string{text}
	}
	return result
}
