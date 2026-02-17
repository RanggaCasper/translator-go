package translator

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const googleTranslateURL = "https://translate.googleapis.com/translate_a/single"

var (
	// Reuse HTTP client with connection pooling
	httpClient     *http.Client
	httpClientOnce sync.Once
)

func getHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		httpClient = &http.Client{
			Timeout: 10 * time.Second, // Faster timeout
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	})
	return httpClient
}

// GoogleTranslate translates text using Google Translate free API
func GoogleTranslate(text, targetLang, sourceLang string) (string, error) {
	if text == "" {
		return text, nil
	}

	client := getHTTPClient()

	params := url.Values{}
	params.Set("client", "gtx")
	params.Set("sl", sourceLang)
	params.Set("tl", targetLang)
	params.Set("dt", "t")
	params.Set("q", text)

	reqURL := fmt.Sprintf("%s?%s", googleTranslateURL, params.Encode())

	resp, err := client.Get(reqURL)
	if err != nil {
		return text, fmt.Errorf("google translate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return text, fmt.Errorf("google translate returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, fmt.Errorf("failed to read response: %w", err)
	}

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return text, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result) == 0 {
		return text, fmt.Errorf("empty translation result")
	}

	translations, ok := result[0].([]interface{})
	if !ok || len(translations) == 0 {
		return text, fmt.Errorf("invalid translation format")
	}

	var translated string
	for _, trans := range translations {
		if transArray, ok := trans.([]interface{}); ok && len(transArray) > 0 {
			if transText, ok := transArray[0].(string); ok {
				translated += transText
			}
		}
	}

	// Apply informal style for Indonesian
	if targetLang == "id" {
		translated = FormalizeToInformal(translated)
	}

	return translated, nil
}
