package game

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequestHint(t *testing.T) {
	mockDictService := new(MockDictionaryService)
	engine := NewGameEngine("test-game", mockDictService)

	// Set up the current word
	now := time.Now()
	testWord := &Word{
		Word:       "TESTING",
		Definition: "A test word",
	}
	engine.CurrentWord = testWord
	engine.TurnStartedAt = &now

	mockDictService.On("GetHint", mock.Anything, testWord, HintTypeDefinition).Return("A test word", nil)

	hint, err := engine.GetHint(context.Background(), HintTypeDefinition)
	assert.NoError(t, err)
	assert.Equal(t, "A test word", hint)
}

func TestUnmaskWord(t *testing.T) {
	mockDictService := new(MockDictionaryService)
	engine := NewGameEngine("test-game", mockDictService)

	// Set up the current word
	now := time.Now()
	engine.CurrentWord = &Word{
		Word:       "TESTING",
		Definition: "A test word",
	}
	engine.TurnStartedAt = &now
	engine.WordMasked = true

	err := engine.RevealWord()
	assert.NoError(t, err)
	assert.False(t, engine.WordMasked)
}
