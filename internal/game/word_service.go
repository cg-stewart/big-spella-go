package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type wordService struct {
	db        *sqlx.DB
	apiKey    string
	apiClient *http.Client
}

func NewWordService(db *sqlx.DB, apiKey string) WordService {
	return &wordService{
		db:        db,
		apiKey:    apiKey,
		apiClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *wordService) GetRandomWord(ctx context.Context, level int, category *string) (*Word, error) {
	query := `
		SELECT * FROM words
		WHERE level = $1`
	args := []interface{}{level}

	if category != nil {
		query += " AND category = $2"
		args = append(args, *category)
	}

	query += `
		ORDER BY RANDOM()
		LIMIT 1`

	word := &Word{}
	err := s.db.GetContext(ctx, word, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get random word: %w", err)
	}

	return word, nil
}

func (s *wordService) ValidateSpelling(ctx context.Context, word, attempt string) bool {
	return strings.EqualFold(strings.TrimSpace(word), strings.TrimSpace(attempt))
}

type TranscriptionRequest struct {
	File      []byte `json:"file"`
	Model     string `json:"model"`
	Language  string `json:"language"`
	Prompt    string `json:"prompt"`
	Response  string `json:"response_format"`
	Temperature float32 `json:"temperature"`
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

func (s *wordService) TranscribeVoice(ctx context.Context, voiceData []byte) (string, error) {
	url := "https://api.openai.com/v1/audio/transcriptions"

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the audio file
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, bytes.NewReader(voiceData))
	if err != nil {
		return "", fmt.Errorf("failed to copy voice data: %w", err)
	}

	// Add other fields
	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "en")
	writer.WriteField("prompt", "This is a spelling bee game. The audio will contain a single word spelled out.")
	writer.WriteField("response_format", "json")
	writer.WriteField("temperature", "0.2")

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := s.apiClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Clean up the transcribed text
	text := strings.TrimSpace(result.Text)
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "!", "")
	text = strings.ReplaceAll(text, "?", "")

	return text, nil
}
