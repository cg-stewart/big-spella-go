package profile

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Profile represents a user's extended profile information
type Profile struct {
	UserID               uuid.UUID       `json:"user_id" db:"user_id"`
	Bio                  string         `json:"bio" db:"bio"`
	ProfileImageURL      string         `json:"profile_image_url" db:"profile_image_url"`
	SocialLinks         json.RawMessage `json:"social_links" db:"social_links"`
	NotificationPrefs   json.RawMessage `json:"notification_preferences" db:"notification_preferences"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// WordHistory represents a user's history with a specific word
type WordHistory struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	WordID           uuid.UUID `json:"word_id" db:"word_id"`
	Status           string    `json:"status" db:"status"`
	CorrectAttempts  int       `json:"correct_attempts" db:"correct_attempts"`
	IncorrectAttempts int      `json:"incorrect_attempts" db:"incorrect_attempts"`
	LastAttemptAt    time.Time `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
	NextReviewAt     time.Time `json:"next_review_at,omitempty" db:"next_review_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Post represents a user's social media post
type Post struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	Type          string         `json:"type" db:"type"`
	Content       json.RawMessage `json:"content" db:"content"`
	GameID        *uuid.UUID     `json:"game_id,omitempty" db:"game_id"`
	MediaURLs     json.RawMessage `json:"media_urls" db:"media_urls"`
	LikesCount    int            `json:"likes_count" db:"likes_count"`
	CommentsCount int            `json:"comments_count" db:"comments_count"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// PostInteraction represents a user's interaction with a post
type PostInteraction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PostID    uuid.UUID `json:"post_id" db:"post_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Content   string    `json:"content,omitempty" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserFollow represents a follow relationship between users
type UserFollow struct {
	FollowerID  uuid.UUID `json:"follower_id" db:"follower_id"`
	FollowingID uuid.UUID `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
