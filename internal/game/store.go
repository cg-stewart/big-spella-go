package game

import (
	"context"

	"github.com/google/uuid"
)

// GameStore defines the interface for game persistence operations
type GameStore interface {
	CreateGame(ctx context.Context, game *Game) error
	GetGame(ctx context.Context, id uuid.UUID) (*Game, error)
	UpdateGame(ctx context.Context, game *Game) error
	DeleteGame(ctx context.Context, id uuid.UUID) error
	ListGames(ctx context.Context, filter GameFilter) ([]*Game, error)
	
	// Player operations
	AddPlayer(ctx context.Context, gameID uuid.UUID, player *Player) error
	RemovePlayer(ctx context.Context, gameID uuid.UUID, playerID uuid.UUID) error
	UpdatePlayerScore(ctx context.Context, gameID uuid.UUID, playerID uuid.UUID, score int) error
	
	// Word operations
	SetCurrentWord(ctx context.Context, gameID uuid.UUID, word *Word) error
	GetCurrentWord(ctx context.Context, gameID uuid.UUID) (*Word, error)
	
	// Attempt operations
	RecordAttempt(ctx context.Context, attempt *SpellingAttempt) error
	GetAttempts(ctx context.Context, gameID uuid.UUID) ([]*SpellingAttempt, error)
}

// GameFilter defines the criteria for filtering games
type GameFilter struct {
	Status    *GameStatus
	Type      *GameType
	HostID    *uuid.UUID
	PlayerID  *uuid.UUID
	Limit     int
	Offset    int
}

// NewGameFilter creates a new GameFilter with default values
func NewGameFilter() GameFilter {
	return GameFilter{
		Limit:  50,
		Offset: 0,
	}
}
