package game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrGameNotFound      = errors.New("game not found")
	ErrGameFull         = errors.New("game is full")
	ErrInvalidGameState = errors.New("invalid game state")
	ErrNotPlayerTurn    = errors.New("not player's turn")
	ErrPlayerNotFound   = errors.New("player not found")
)

type GameService interface {
	CreateGame(ctx context.Context, hostID string, gameType GameType, settings GameSettings) (*Game, error)
	JoinGame(ctx context.Context, gameID string, playerID string) (*Game, error)
	StartGame(ctx context.Context, gameID string, userID string) (*Game, error)
	MakeAttempt(ctx context.Context, gameID string, playerID string, attempt *SpellingAttempt) error
	GetGame(ctx context.Context, gameID string) (*Game, error)
	GetHint(ctx context.Context, gameID string, playerID string) (*Hint, error)
	Events() <-chan GameEvent
}

type gameService struct {
	db           *sqlx.DB
	wordService  WordService
	dictService  DictionaryService
	eventChan    chan GameEvent
	activeGames  map[string]*GameEngine
}

type WordService interface {
	GetRandomWord(ctx context.Context, level int, category *string) (*Word, error)
	ValidateSpelling(ctx context.Context, word, attempt string) bool
	TranscribeVoice(ctx context.Context, voiceData []byte) (string, error)
}

func NewGameService(db *sqlx.DB, wordService WordService, dictService DictionaryService) GameService {
	return &gameService{
		db:          db,
		wordService: wordService,
		dictService: dictService,
		eventChan:   make(chan GameEvent, 100),
		activeGames: make(map[string]*GameEngine),
	}
}

func (s *gameService) CreateGame(ctx context.Context, hostID string, gameType GameType, settings GameSettings) (*Game, error) {
	id := uuid.New().String()
	game := &Game{
		ID:        id,
		HostID:    hostID,
		Type:      gameType,
		Status:    GameStatusInitializing,
		Settings:  settings,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO games (id, host_id, type, status, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	if err := s.db.QueryRowContext(ctx, query,
		game.ID, game.HostID, game.Type, game.Status, game.Settings,
		game.CreatedAt, game.UpdatedAt).Scan(&game.ID); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Create game engine
	s.activeGames[game.ID] = NewGameEngine(game.ID, s.dictService)

	s.emitEvent(EventTypeGameCreated, game.ID, nil, map[string]any{
		"game": game,
	})

	return game, nil
}

func (s *gameService) JoinGame(ctx context.Context, gameID string, playerID string) (*Game, error) {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != GameStatusWaiting {
		return nil, ErrInvalidGameState
	}

	// Check if player count is within limits
	var playerCount int
	if err := s.db.GetContext(ctx, &playerCount,
		"SELECT COUNT(*) FROM players WHERE game_id = $1", gameID); err != nil {
		return nil, fmt.Errorf("failed to count players: %w", err)
	}

	if playerCount >= game.Settings.MaxPlayers {
		return nil, ErrGameFull
	}

	// Add player
	player := &Player{
		ID:       uuid.New().String(),
		GameID:   gameID,
		UserID:   playerID,
		Status:   "active",
		JoinedAt: time.Now(),
	}

	query := `
		INSERT INTO players (id, game_id, user_id, status, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	if err := s.db.QueryRowContext(ctx, query,
		player.ID, player.GameID, player.UserID, player.Status,
		player.JoinedAt).Scan(&player.ID); err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}

	s.emitEvent(EventTypePlayerJoined, gameID, &playerID, map[string]any{
		"player": player,
	})

	return game, nil
}

func (s *gameService) StartGame(ctx context.Context, gameID string, userID string) (*Game, error) {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != GameStatusWaiting {
		return nil, ErrInvalidGameState
	}

	// Get first word
	word, err := s.wordService.GetRandomWord(ctx, game.Settings.WordLevel, game.Settings.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	// Start game engine
	engine := s.activeGames[gameID]
	if engine == nil {
		engine = NewGameEngine(gameID, s.dictService)
		s.activeGames[gameID] = engine
	}

	if err := engine.StartTurn(ctx, word.Word); err != nil {
		return nil, fmt.Errorf("failed to start turn: %w", err)
	}

	// Update game status
	query := `
		UPDATE games
		SET status = $1, current_word_id = $2, updated_at = $3,
			turn_started_at = $4, word_masked = $5
		WHERE id = $6
		RETURNING *`

	now := time.Now()
	if err := s.db.GetContext(ctx, game, query,
		GameStatusActive, word.ID, now,
		now, true, gameID); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	s.emitEvent(EventTypeGameStarted, gameID, nil, map[string]any{
		"game": game,
		"word": word,
	})

	return game, nil
}

func (s *gameService) MakeAttempt(ctx context.Context, gameID string, playerID string, attempt *SpellingAttempt) error {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.Status != GameStatusActive {
		return ErrInvalidGameState
	}

	engine := s.activeGames[gameID]
	if engine == nil {
		return ErrGameNotFound
	}

	// Validate attempt
	isCorrect, err := engine.ValidateAttempt(attempt.Text)
	if err != nil {
		return fmt.Errorf("failed to validate attempt: %w", err)
	}

	// Update game state based on result
	now := time.Now()
	var query string
	var args []interface{}

	if isCorrect {
		// Player succeeded - update score and move to next word
		query = `
			UPDATE games
			SET current_word_id = NULL,
				updated_at = $1,
				turn_started_at = NULL,
				word_masked = false,
				scores = jsonb_set(
					scores,
					array[$2],
					(COALESCE((scores->$2)::int, 0) + 1)::text::jsonb
				)
			WHERE id = $4
			RETURNING *`
		args = []interface{}{now, playerID, gameID}
	} else {
		// Player failed - just update timestamp
		query = `
			UPDATE games
			SET updated_at = $1
			WHERE id = $2
			RETURNING *`
		args = []interface{}{now, gameID}
	}

	if err := s.db.GetContext(ctx, game, query, args...); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Emit appropriate event
	eventType := EventTypeAttemptFailed
	if isCorrect {
		eventType = EventTypeAttemptSucceeded
	}

	s.emitEvent(eventType, gameID, &playerID, map[string]any{
		"attempt": attempt,
		"correct": isCorrect,
	})

	return nil
}

func (s *gameService) nextTurn(ctx context.Context, game *Game) error {
	// Get next word
	word, err := s.wordService.GetRandomWord(ctx, game.Settings.WordLevel, game.Settings.Category)
	if err != nil {
		return fmt.Errorf("failed to get next word: %w", err)
	}

	engine := s.activeGames[game.ID]
	if engine == nil {
		return ErrGameNotFound
	}

	if err := engine.StartTurn(ctx, word.Word); err != nil {
		return fmt.Errorf("failed to start turn: %w", err)
	}

	// Update game state
	now := time.Now()
	query := `
		UPDATE games
		SET current_word_id = $1,
			updated_at = $2,
			turn_started_at = $3,
			word_masked = true,
			round = round + 1
		WHERE id = $4
		RETURNING *`

	if err := s.db.GetContext(ctx, game, query, word.ID, now, now, game.ID); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	s.emitEvent(EventTypeRoundStarted, game.ID, nil, map[string]any{
		"game": game,
		"word": word,
	})

	return nil
}

func (s *gameService) GetGame(ctx context.Context, gameID string) (*Game, error) {
	query := `
		SELECT g.*, array_agg(p.*) as players
		FROM games g
		LEFT JOIN players p ON p.game_id = g.id
		WHERE g.id = $1
		GROUP BY g.id`

	var game Game
	if err := s.db.GetContext(ctx, &game, query, gameID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Get active game engine if exists
	if engine, ok := s.activeGames[gameID]; ok {
		game.CurrentWord = engine.CurrentWord
		game.WordMasked = engine.WordMasked
		game.TurnStartedAt = engine.TurnStartedAt
	}

	return &game, nil
}

func (s *gameService) GetHint(ctx context.Context, gameID string, playerID string) (*Hint, error) {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	if game.Status != GameStatusActive {
		return nil, ErrInvalidGameState
	}

	engine := s.activeGames[gameID]
	if engine == nil {
		return nil, ErrGameNotFound
	}

	// Get a random hint type
	hintTypes := []HintType{
		HintTypeDefinition,
		HintTypeExampleSentence,
		HintTypeEtymology,
		HintTypePartOfSpeech,
		HintTypePronunciation,
	}
	hintType := hintTypes[time.Now().UnixNano()%int64(len(hintTypes))]

	hint, err := engine.GetHint(ctx, hintType)
	if err != nil {
		return nil, fmt.Errorf("failed to get hint: %w", err)
	}

	s.emitEvent(EventTypeHintRequested, gameID, &playerID, map[string]any{
		"hint": &Hint{
			Type:    hintType,
			Content: hint,
		},
	})

	return &Hint{
		Type:    hintType,
		Content: hint,
	}, nil
}

func (s *gameService) emitEvent(eventType EventType, gameID string, playerID *string, payload map[string]any) {
	event := GameEvent{
		Type:      eventType,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload:   payload,
	}
	s.eventChan <- event
}

func (s *gameService) Events() <-chan GameEvent {
	return s.eventChan
}
