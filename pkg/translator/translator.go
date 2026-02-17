package translator

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FetchAndTranslate fetches a subtitle file from URL and translates it
func FetchAndTranslate(url, format, targetLang, sourceLang, referer string) (string, error) {
	// Fetch subtitle content
	content, err := fetchSubtitle(url, referer)
	if err != nil {
		return "", err
	}

	// Translate based on format
	format = strings.ToLower(format)
	if format == "ass" {
		return TranslateASSToVTT(content, targetLang, sourceLang)
	}

	return TranslateVTT(content, targetLang, sourceLang)
}

func fetchSubtitle(url, referer string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:146.0) Gecko/20100101 Firefox/146.0")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subtitle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch subtitle: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}
