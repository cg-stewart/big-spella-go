package game

import (
	"encoding/json"
	"time"
)

// EventType represents different types of game events
type EventType string

const (
	EventTypeGameCreated     EventType = "game_created"
	EventTypeGameStarted     EventType = "game_started"
	EventTypeGameEnded      EventType = "game_ended"
	EventTypeAttemptSucceeded EventType = "attempt_succeeded"
	EventTypeAttemptFailed   EventType = "attempt_failed"
	EventTypePlayerJoined    EventType = "player_joined"
	EventTypePlayerLeft      EventType = "player_left"
	EventTypeRoundStarted    EventType = "round_started"
	EventTypeRoundEnded      EventType = "round_ended"
	EventTypeHintRequested   EventType = "hint_requested"
)

// HintType represents different types of hints
type HintType string

const (
	HintTypeDefinition      HintType = "definition"
	HintTypeExampleSentence HintType = "example_sentence"
	HintTypeEtymology       HintType = "etymology"
	HintTypeSentence        HintType = "sentence"
	HintTypePartOfSpeech    HintType = "part_of_speech"
	HintTypePronunciation   HintType = "pronunciation"
	HintTypePhonetic        HintType = "phonetic"
	HintTypeSynonym         HintType = "synonym"
)

// GameType represents different types of games
type GameType string

const (
	GameTypeSolo     GameType = "solo"
	GameTypeMulti    GameType = "multi"
	GameTypePractice GameType = "practice"
)

// GameState represents the current state of a game
type GameState string

const (
	GameStateWaiting  GameState = "waiting"
	GameStatePlaying  GameState = "playing"
	GameStateFinished GameState = "finished"
)

// GameStatus represents the status of a game
type GameStatus string

const (
	GameStatusCreated      GameStatus = "created"
	GameStatusInitializing GameStatus = "initializing"
	GameStatusWaiting      GameStatus = "waiting"
	GameStatusPlaying      GameStatus = "playing"
	GameStatusActive       GameStatus = "active"
	GameStatusFinished     GameStatus = "finished"
	GameStatusCancelled    GameStatus = "cancelled"
)

// Word represents a word and its associated information
type Word struct {
	ID              string    `json:"id" db:"id"`
	Word            string    `json:"word" db:"word"`
	Definition      string    `json:"definition" db:"definition"`
	ExampleSentence string    `json:"example_sentence" db:"example_sentence"`
	Etymology       string    `json:"etymology" db:"etymology"`
	PartOfSpeech    string    `json:"part_of_speech" db:"part_of_speech"`
	Pronunciation   string    `json:"pronunciation" db:"pronunciation"`
	AudioURL        string    `json:"audio_url" db:"audio_url"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Game represents an active game session
type Game struct {
	ID            string          `json:"id" db:"id"`
	Type          GameType        `json:"type" db:"type"`
	Status        GameStatus      `json:"status" db:"status"`
	Mode          string          `json:"mode" db:"mode"`
	Settings      GameSettings    `json:"settings" db:"settings"`
	CurrentWord   *Word           `json:"current_word,omitempty" db:"current_word_id"`
	CurrentTurn   *string         `json:"current_turn,omitempty" db:"current_turn"`
	MeetingID     *string         `json:"meeting_id,omitempty" db:"meeting_id"`
	Round         int             `json:"round" db:"round"`
	MaxRounds     *int            `json:"max_rounds,omitempty" db:"max_rounds"`
	TimeLimit     *time.Duration  `json:"time_limit,omitempty" db:"time_limit"`
	EnableVideo   bool            `json:"enable_video" db:"enable_video"`
	EnableVoice   bool            `json:"enable_voice" db:"enable_voice"`
	RecordGame    bool            `json:"record_game" db:"record_game"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	TurnStartedAt *time.Time      `json:"turn_started_at,omitempty" db:"turn_started_at"`
	HintsUsed     map[string][]string `json:"hints_used,omitempty" db:"hints_used"`
	WordMasked    bool            `json:"word_masked" db:"word_masked"`
	HostID        string          `json:"host_id" db:"host_id"`
	LastActivity  time.Time       `json:"last_activity" db:"last_activity"`
	CurrentPlayer string          `json:"current_player" db:"current_player"`
	Players       []*Player       `json:"players" db:"players"`
}

// GameSettings represents the settings for a game
type GameSettings struct {
	MinPlayers  int           `json:"min_players"`
	MaxPlayers  int           `json:"max_players"`
	TimeLimit   time.Duration `json:"time_limit"`
	Category    *string       `json:"category,omitempty"`
	IsRanked    bool         `json:"is_ranked"`
	Elimination bool         `json:"elimination"`
	WordLevel   int          `json:"word_level"`
	HintsAllowed int         `json:"hints_allowed"`
	SpellStartTimeout time.Duration `json:"spell_start_timeout"`
}

// Player represents a player in a game
type Player struct {
	ID       string    `json:"id" db:"id"`
	GameID   string    `json:"game_id" db:"game_id"`
	UserID   string    `json:"user_id" db:"player_id"`
	Score    int       `json:"score" db:"score"`
	Status   string    `json:"status" db:"status"`
	IsBot    bool      `json:"is_bot" db:"is_bot"`
	Attempts int       `json:"attempts" db:"attempts"`
	Correct  int       `json:"correct" db:"correct"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// Hint represents a hint provided during the game
type Hint struct {
	Type    HintType `json:"type"`
	Content string   `json:"content"`
}

// SpellingAttempt represents a player's attempt to spell a word
type SpellingAttempt struct {
	ID        string      `json:"id" db:"id"`
	GameID    string      `json:"game_id" db:"game_id"`
	PlayerID  string      `json:"player_id" db:"player_id"`
	Word      string      `json:"word" db:"word"`
	Type      AttemptType `json:"type" db:"type"`
	VoiceData []byte      `json:"voice_data,omitempty" db:"voice_data"`
	Text      string      `json:"text,omitempty" db:"text"`
	IsCorrect bool        `json:"is_correct" db:"is_correct"`
	Timestamp time.Time   `json:"timestamp" db:"timestamp"`
}

// AttemptType represents the type of spelling attempt
type AttemptType string

const (
	AttemptTypeText  AttemptType = "text"
	AttemptTypeVoice AttemptType = "voice"
)

// GameEvent represents an event that occurred during a game
type GameEvent struct {
	Type      EventType         `json:"type"`
	GameID    string           `json:"game_id"`
	PlayerID  *string          `json:"player_id,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
	Payload   map[string]any   `json:"payload"`
}

const (
	DefaultHintsAllowed = 3
	DefaultSpellStartTimeout = 10 * time.Second
)

// GameResult represents the outcome of a game for a player
type GameResult struct {
	ID                string    `json:"id" db:"id"`
	GameID            string    `json:"game_id" db:"game_id"`
	PlayerID          string    `json:"player_id" db:"player_id"`
	Placement         int       `json:"placement" db:"placement"`
	PointsEarned      int       `json:"points_earned" db:"points_earned"`
	PreviousRankPoints int      `json:"previous_rank_points" db:"previous_rank_points"`
	NewRankPoints     int       `json:"new_rank_points" db:"new_rank_points"`
	PreviousRankColor string    `json:"previous_rank_color" db:"previous_rank_color"`
	NewRankColor      string    `json:"new_rank_color" db:"new_rank_color"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// GameRecording represents metadata about a recorded game
type GameRecording struct {
	ID        string        `json:"id" db:"id"`
	GameID    string        `json:"game_id" db:"game_id"`
	S3Key     string        `json:"s3_key" db:"s3_key"`
	Duration  time.Duration `json:"duration" db:"duration"`
	SizeBytes int64         `json:"size_bytes" db:"size_bytes"`
	Status    string        `json:"status" db:"status"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at"`
}

// Value implements the driver.Valuer interface for GameSettings
func (g GameSettings) Value() (interface{}, error) {
	return json.Marshal(g)
}
