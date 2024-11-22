package game

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type DictionaryEntry struct {
	Meta struct {
		ID        string   `json:"id"`
		UUID      string   `json:"uuid"`
		Offensive bool     `json:"offensive"`
		Stems     []string `json:"stems"`
	} `json:"meta"`
	HWI struct {
		Pronunciation struct {
			IPA  string `json:"ipa"`
			WAV  string `json:"wav"`
			MWOD []struct {
				Subdirectory string `json:"subdirectory"`
				FileName     string `json:"file"`
			} `json:"mwod"`
		} `json:"prs"`
	} `json:"hwi"`
	FL   string `json:"fl"` // Part of speech
	Def  []Definition `json:"def"`
	Et   []string `json:"et"` // Etymology
	Date string   `json:"date"`
}

type Definition struct {
	SseqList [][]struct {
		Sense struct {
			DT [][]interface{} `json:"dt"`
			VIS []struct {
				T string `json:"t"` // Example sentence
			} `json:"vis,omitempty"`
		} `json:"sense,omitempty"`
	} `json:"sseq"`
}

type DictionaryService interface {
	GetWordInfo(ctx context.Context, word string) (*Word, error)
	GenerateAudio(ctx context.Context, text string) ([]byte, error)
	GetHint(ctx context.Context, word *Word, hintType HintType) (string, error)
}

type dictionaryService struct {
	dictionaryAPIKey string
	thesaurusAPIKey  string
	openAIKey        string
	httpClient       *http.Client
}

func NewDictionaryService(dictionaryAPIKey, thesaurusAPIKey, openAIKey string) DictionaryService {
	return &dictionaryService{
		dictionaryAPIKey: dictionaryAPIKey,
		thesaurusAPIKey:  thesaurusAPIKey,
		openAIKey:       openAIKey,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (s *dictionaryService) GetWordInfo(ctx context.Context, word string) (*Word, error) {
	url := fmt.Sprintf("https://www.dictionaryapi.com/api/v3/references/collegiate/json/%s?key=%s",
		word, s.dictionaryAPIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get word info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var entries []DictionaryEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("word not found: %s", word)
	}

	entry := entries[0]
	wordInfo := &Word{
		Word:          word,
		PartOfSpeech: entry.FL,
	}

	// Get pronunciation
	if len(entry.HWI.Pronunciation.MWOD) > 0 {
		pron := entry.HWI.Pronunciation.MWOD[0]
		wordInfo.AudioURL = fmt.Sprintf(
			"https://media.merriam-webster.com/audio/prons/en/us/mp3/%s/%s.mp3",
			pron.Subdirectory, pron.FileName)
		wordInfo.Pronunciation = entry.HWI.Pronunciation.IPA
	}

	// Get definition and example
	if len(entry.Def) > 0 && len(entry.Def[0].SseqList) > 0 {
		for _, sseq := range entry.Def[0].SseqList {
			if len(sseq) > 0 {
				sense := sseq[0].Sense
				if len(sense.DT) > 0 && len(sense.DT[0]) > 1 {
					if def, ok := sense.DT[0][1].(string); ok {
						wordInfo.Definition = strings.TrimSpace(def)
						break
					}
				}
				if len(sense.VIS) > 0 {
					wordInfo.ExampleSentence = strings.TrimSpace(sense.VIS[0].T)
				}
			}
		}
	}

	// Get etymology
	if len(entry.Et) > 0 {
		wordInfo.Etymology = strings.Join(entry.Et, " ")
	}

	return wordInfo, nil
}

func (s *dictionaryService) GenerateAudio(ctx context.Context, text string) ([]byte, error) {
	url := "https://api.openai.com/v1/audio/speech"
	reqBody := map[string]interface{}{
		"model": "tts-1",
		"input": text,
		"voice": "onyx",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.openAIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audioData, nil
}

func (s *dictionaryService) GetHint(ctx context.Context, word *Word, hintType HintType) (string, error) {
	switch hintType {
	case HintTypeDefinition:
		return word.Definition, nil
	case HintTypeSentence:
		return word.ExampleSentence, nil
	case HintTypeEtymology:
		return word.Etymology, nil
	case HintTypePartOfSpeech:
		return word.PartOfSpeech, nil
	case HintTypePronunciation:
		return word.Pronunciation, nil
	default:
		return "", fmt.Errorf("invalid hint type: %s", hintType)
	}
}
