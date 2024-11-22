package modes

import (
	"fmt"
	"time"
)

// GameMode represents different game modes
type GameMode string

const (
	ModeRoundRobin GameMode = "round_robin" // Tournament default
	ModeRapidFire  GameMode = "rapid_fire"  // 1v1 speed spelling
	ModeTotalGame  GameMode = "total_game"  // Time/round-based points
)

// GameSettings represents the configuration for a game
type GameSettings struct {
	Mode              GameMode        `json:"mode"`
	MaxPlayers        int            `json:"max_players"`
	MaxRounds         int            `json:"max_rounds,omitempty"`
	TimeLimit         time.Duration  `json:"time_limit,omitempty"`
	WordLevel         int            `json:"word_level"`
	Category         string         `json:"category"`
	IsTournament     bool           `json:"is_tournament"`
	IsPrivate        bool           `json:"is_private"`
	EnableVideo      bool           `json:"enable_video"`
	EnableVoice      bool           `json:"enable_voice"`
	RecordGame       bool           `json:"record_game"`
}

// DefaultSettings returns default settings for each game mode
func DefaultSettings(mode GameMode) GameSettings {
	base := GameSettings{
		Mode:          mode,
		WordLevel:     1,
		EnableVoice:   true,
		RecordGame:    false,
	}

	switch mode {
	case ModeRoundRobin:
		base.MaxPlayers = 32
		base.MaxRounds = 10
		base.EnableVideo = true
		
	case ModeRapidFire:
		base.MaxPlayers = 2
		base.TimeLimit = 10 * time.Minute
		base.EnableVideo = true
		
	case ModeTotalGame:
		base.MaxPlayers = 8
		base.MaxRounds = 20
		base.TimeLimit = 30 * time.Minute
		base.EnableVideo = true
	}

	return base
}

// ValidateSettings validates game settings
func ValidateSettings(settings GameSettings) error {
	switch settings.Mode {
	case ModeRoundRobin:
		if settings.MaxPlayers < 2 || settings.MaxPlayers > 32 {
			return fmt.Errorf("round robin requires 2-32 players")
		}
		if settings.MaxRounds < 1 {
			return fmt.Errorf("round robin requires at least 1 round")
		}

	case ModeRapidFire:
		if settings.MaxPlayers != 2 {
			return fmt.Errorf("rapid fire is strictly 1v1")
		}
		if settings.TimeLimit < time.Minute || settings.TimeLimit > 30*time.Minute {
			return fmt.Errorf("rapid fire time limit must be between 1-30 minutes")
		}

	case ModeTotalGame:
		if settings.MaxPlayers < 2 || settings.MaxPlayers > 8 {
			return fmt.Errorf("total game requires 2-8 players")
		}
		if settings.TimeLimit < 5*time.Minute || settings.TimeLimit > time.Hour {
			return fmt.Errorf("total game time limit must be between 5-60 minutes")
		}
	}

	if settings.WordLevel < 1 || settings.WordLevel > 10 {
		return fmt.Errorf("word level must be between 1-10")
	}

	return nil
}

// CalculateScore calculates the score based on game mode and performance
func CalculateScore(mode GameMode, correctAttempts, totalAttempts int, averageTime float64) int {
	baseScore := correctAttempts * 100

	switch mode {
	case ModeRapidFire:
		if averageTime < 5.0 {
			return int(float64(baseScore) * 1.5) // Speed bonus
		}
	case ModeTotalGame:
		accuracy := float64(correctAttempts) / float64(totalAttempts)
		if accuracy >= 0.9 {
			return int(float64(baseScore) * 1.3) // Accuracy bonus
		}
	}

	return baseScore
}

// IsCompetitive returns whether a game mode affects ranking
func IsCompetitive(settings GameSettings) bool {
	return !settings.IsPrivate && (settings.IsTournament || settings.Mode != ModeRapidFire)
}

// RequiresRecording returns whether a game should be recorded
func RequiresRecording(settings GameSettings) bool {
	return settings.RecordGame && !settings.IsPrivate && settings.IsTournament
}
