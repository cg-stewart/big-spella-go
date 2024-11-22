package game

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	MaxHints    = 3
	TurnTimeout = 10 * time.Second
)

var (
	ErrNoWordSet     = errors.New("no word is set for the current turn")
	ErrMaxHintsUsed  = errors.New("maximum number of hints already used")
	ErrTurnNotActive = errors.New("no active turn")
)

type GameEngine struct {
	ID            string
	dict          DictionaryService
	CurrentWord   *Word
	WordMasked    bool
	HintsUsed     int
	TurnStartedAt *time.Time
}

func NewGameEngine(id string, dict DictionaryService) *GameEngine {
	return &GameEngine{
		ID:   id,
		dict: dict,
	}
}

func (g *GameEngine) StartNewTurn(ctx context.Context) error {
	word, err := g.dict.GetWordInfo(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get word: %w", err)
	}
	
	now := time.Now()
	g.CurrentWord = word
	g.WordMasked = true
	g.HintsUsed = 0
	g.TurnStartedAt = &now
	
	return nil
}

func (g *GameEngine) StartTurn(ctx context.Context, word string) error {
	wordInfo, err := g.dict.GetWordInfo(ctx, word)
	if err != nil {
		return fmt.Errorf("failed to get word info: %w", err)
	}

	now := time.Now()
	g.CurrentWord = wordInfo
	g.WordMasked = true
	g.HintsUsed = 0
	g.TurnStartedAt = &now

	return nil
}

func (g *GameEngine) ValidateAttempt(attempt string) (bool, error) {
	if g.CurrentWord == nil {
		return false, ErrNoWordSet
	}
	
	if g.TurnStartedAt == nil {
		return false, ErrTurnNotActive
	}
	
	if time.Since(*g.TurnStartedAt) > TurnTimeout {
		return false, errors.New("turn has timed out")
	}
	
	return strings.EqualFold(attempt, g.CurrentWord.Word), nil
}

func (g *GameEngine) GetHint(ctx context.Context, hintType HintType) (string, error) {
	if g.CurrentWord == nil {
		return "", ErrNoWordSet
	}
	
	if g.HintsUsed >= MaxHints {
		return "", ErrMaxHintsUsed
	}
	
	hint, err := g.dict.GetHint(ctx, g.CurrentWord, hintType)
	if err != nil {
		return "", fmt.Errorf("failed to get hint: %w", err)
	}
	
	g.HintsUsed++
	return hint, nil
}

func (g *GameEngine) CheckTimeLimit() bool {
	if g.TurnStartedAt == nil {
		return false
	}
	return time.Since(*g.TurnStartedAt) <= TurnTimeout
}

func (g *GameEngine) RevealWord() error {
	if g.CurrentWord == nil {
		return ErrNoWordSet
	}
	g.WordMasked = false
	return nil
}

func (g *GameEngine) GenerateWordAudio(ctx context.Context) ([]byte, error) {
	if g.CurrentWord == nil {
		return nil, ErrNoWordSet
	}
	return g.dict.GenerateAudio(ctx, g.CurrentWord.Word)
}

func (g *GameEngine) RequestHint(hintType HintType) (*Hint, error) {
	if g.CurrentWord == nil {
		return nil, fmt.Errorf("no word is currently active")
	}

	switch hintType {
	case HintTypeDefinition:
		return &Hint{
			Type:    HintTypeDefinition,
			Content: "Sample definition hint", // TODO: Get from dictionary service
		}, nil
	case HintTypePhonetic:
		return &Hint{
			Type:    HintTypePhonetic,
			Content: "Sample phonetic hint", // TODO: Get from dictionary service
		}, nil
	case HintTypeSynonym:
		return &Hint{
			Type:    HintTypeSynonym,
			Content: "Sample synonym hint", // TODO: Get from dictionary service
		}, nil
	default:
		return nil, fmt.Errorf("unsupported hint type: %v", hintType)
	}
}

func (g *GameEngine) UnmaskWord() string {
	if g.CurrentWord == nil {
		return ""
	}
	g.WordMasked = false
	return g.CurrentWord.Word
}
