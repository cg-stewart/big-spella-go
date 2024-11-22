package game

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateGame(t *testing.T) {
	mockDB := &MockDB{}
	mockWordService := new(MockWordService)
	mockDictService := new(MockDictionaryService)
	service := NewGameService(mockDB, mockWordService, mockDictService)

	ctx := context.Background()
	gameID := uuid.New().String()
	settings := GameSettings{
		MinPlayers: 2,
		MaxPlayers: 4,
		TimeLimit:  300,
	}

	mockWordService.On("GetRandomWord", ctx, mock.Anything, mock.Anything).Return(&Word{
		Word:       "TESTING",
		Definition: "A test word",
	}, nil)
	mockDB.On("CreateGame", ctx, mock.AnythingOfType("*game.Game")).Return(nil)

	game, err := service.CreateGame(ctx, gameID, GameTypeSolo, settings)
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, gameID, game.ID)
	assert.Equal(t, GameTypeSolo, game.Type)
	assert.Equal(t, GameStatusInitializing, game.Status)

	mockDB.AssertExpectations(t)
	mockWordService.AssertExpectations(t)
}

func TestJoinGame(t *testing.T) {
	mockDB := &MockDB{}
	mockWordService := new(MockWordService)
	mockDictService := new(MockDictionaryService)
	service := NewGameService(mockDB, mockWordService, mockDictService)

	ctx := context.Background()
	gameID := uuid.New().String()
	playerID := uuid.New().String()
	existingGame := &Game{
		ID:      gameID,
		Type:    GameTypeSolo,
		Status:  GameStatusWaiting,
		Players: []*Player{},
		Settings: GameSettings{
			MinPlayers: 2,
			MaxPlayers: 4,
		},
	}

	mockDB.On("GetGame", ctx, gameID).Return(existingGame, nil)
	mockDB.On("UpdateGame", ctx, mock.AnythingOfType("*game.Game")).Return(nil)

	game, err := service.JoinGame(ctx, gameID, playerID)
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Contains(t, game.Players, &Player{ID: playerID})

	mockDB.AssertExpectations(t)
}

func TestStartGame(t *testing.T) {
	mockDB := &MockDB{}
	mockWordService := new(MockWordService)
	mockDictService := new(MockDictionaryService)
	service := NewGameService(mockDB, mockWordService, mockDictService)

	ctx := context.Background()
	gameID := uuid.New().String()
	player1ID := uuid.New().String()
	player2ID := uuid.New().String()

	existingGame := &Game{
		ID:     gameID,
		Type:   GameTypeSolo,
		Status: GameStatusWaiting,
		Players: []*Player{
			{ID: player1ID},
			{ID: player2ID},
		},
		Settings: GameSettings{
			MinPlayers: 2,
			MaxPlayers: 4,
		},
	}

	mockDB.On("GetGame", ctx, gameID).Return(existingGame, nil)
	mockDB.On("UpdateGame", ctx, mock.AnythingOfType("*game.Game")).Return(nil)

	game, err := service.StartGame(ctx, gameID, player1ID)
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, GameStatusPlaying, game.Status)

	mockDB.AssertExpectations(t)
}
