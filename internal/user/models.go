package user

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents a user's profile in the system
type Profile struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Username      string    `json:"username" db:"username"`
	Email         string    `json:"email" db:"email"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	FirstName     string    `json:"first_name" db:"first_name"`
	LastName      string    `json:"last_name" db:"last_name"`
	Age           int       `json:"age" db:"age"`
	PhoneNumber   string    `json:"phone_number" db:"phone_number"`
	ProfilePicURL string    `json:"profile_pic_url" db:"profile_pic_url"`
	
	// Game statistics
	TotalGames     int     `json:"total_games" db:"total_games"`
	GamesWon       int     `json:"games_won" db:"games_won"`
	WinRate        float64 `json:"win_rate" db:"win_rate"`
	AverageScore   float64 `json:"average_score" db:"average_score"`
	HighestScore   int     `json:"highest_score" db:"highest_score"`
	CurrentStreak  int     `json:"current_streak" db:"current_streak"`
	LongestStreak  int     `json:"longest_streak" db:"longest_streak"`
	RankingPoints  int     `json:"ranking_points" db:"ranking_points"`
	CurrentRank    string  `json:"current_rank" db:"current_rank"`

	// Social features
	Followers    int      `json:"followers" db:"followers"`
	Following    int      `json:"following" db:"following"`
	
	// Activity tracking
	LastActive   time.Time `json:"last_active" db:"last_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserActivity tracks user actions in the system
type UserActivity struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Details   string    `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GameHistory represents a user's game history
type GameHistory struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	GameID      uuid.UUID `json:"game_id" db:"game_id"`
	GameType    string    `json:"game_type" db:"game_type"`
	Score       int       `json:"score" db:"score"`
	Position    int       `json:"position" db:"position"`
	WordsSpelled []string  `json:"words_spelled" db:"words_spelled"`
	Duration    int       `json:"duration" db:"duration"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserAchievement tracks user achievements
type UserAchievement struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`
	UnlockedAt   time.Time `json:"unlocked_at" db:"unlocked_at"`
}

// UserPreferences stores user settings
type UserPreferences struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	NotificationsOn bool      `json:"notifications_on" db:"notifications_on"`
	Theme           string    `json:"theme" db:"theme"`
	Language        string    `json:"language" db:"language"`
	SoundEffects    bool      `json:"sound_effects" db:"sound_effects"`
	Music          bool      `json:"music" db:"music"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
