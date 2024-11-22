# Big Spella Technical Specification

## Project Overview

Big Spella is a real-time multiplayer spelling bee platform built using Go, with support for:
- Solo practice with AI bots
- Group matches (2-32 players)
- Tournament play (up to 128 players)
- Voice/Text input using OpenAI Whisper
- Video/Audio rooms via AWS Chime SDK
- Real-time feeds and chat via GetStream
- ELO-based matchmaking
- Premium subscriptions

## Project Structure

```
github.com/newrgm/bigspella/
├── cmd/
│   └── server/
│       └── main.go            # Application entry point
├── internal/
│   ├── auth/                 # Authentication
│   │   ├── handler.go
│   │   ├── middleware.go
│   │   └── service.go
│   ├── game/                 # Core game logic
│   │   ├── engine.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── validators.go
│   ├── tournament/           # Tournament system
│   │   ├── bracket.go
│   │   ├── models.go
│   │   └── service.go
│   ├── bot/                  # AI bot system 
│   │   ├── actions.go
│   │   ├── models.go
│   │   └── service.go
│   ├── matchmaking/          # Matchmaking system
│   │   ├── elo.go
│   │   ├── queue.go
│   │   └── service.go
│   └── notification/         # Notifications
│       ├── email.go
│       ├── push.go
│       └── service.go
├── pkg/
│   ├── chime/               # AWS Chime SDK
│   ├── stream/              # GetStream
│   ├── openai/              # OpenAI API
│   ├── payment/             # Stripe integration
│   └── db/                  # Database
├── api/
│   ├── http/                # HTTP handlers
│   ├── ws/                  # WebSocket handlers
│   └── proto/               # Protocol buffers
├── migrations/              # SQL migrations
├── config/                 # Configuration
└── scripts/                # Build/deployment
```

## Core Domain Models

```go
// internal/game/models.go

package game

import (
    "time"
    "github.com/google/uuid"
)

type GameType string

const (
    GameTypeSolo       GameType = "solo"
    GameTypeGroup      GameType = "group"
    GameTypeTournament GameType = "tournament"
    GameTypeRanked     GameType = "ranked"
)

type GameStatus string

const (
    GameStatusInitializing GameStatus = "initializing"
    GameStatusWaiting      GameStatus = "waiting"
    GameStatusActive       GameStatus = "active"
    GameStatusCompleted    GameStatus = "completed"
)

type Game struct {
    ID          uuid.UUID   `json:"id" db:"id"`
    Type        GameType    `json:"type" db:"type"`
    Status      GameStatus  `json:"status" db:"status"`
    Settings    GameSettings `json:"settings" db:"settings"`
    CurrentWord *Word       `json:"current_word,omitempty" db:"current_word"`
    CurrentTurn *uuid.UUID  `json:"current_turn,omitempty" db:"current_turn"`
    Players     []Player    `json:"players" db:"players"`
    MeetingID   *string     `json:"meeting_id,omitempty" db:"meeting_id"`
    Round       int         `json:"round" db:"round"`
    CreatedAt   time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type GameSettings struct {
    MinPlayers  int           `json:"min_players"`
    MaxPlayers  int           `json:"max_players"`
    TimeLimit   time.Duration `json:"time_limit"`
    Category    *string       `json:"category,omitempty"`
    IsRanked    bool         `json:"is_ranked"`
    Elimination bool         `json:"elimination"`
    WordLevel   int          `json:"word_level"`
}

type Player struct {
    ID       uuid.UUID `json:"id" db:"id"`
    Username string    `json:"username" db:"username"`
    ELO      int       `json:"elo" db:"elo"`
    Score    int       `json:"score" db:"score"`
    Status   string    `json:"status" db:"status"`
    IsBot    bool      `json:"is_bot" db:"is_bot"`
    Attempts int       `json:"attempts" db:"attempts"`
    Correct  int       `json:"correct" db:"correct"`
}

type Word struct {
    ID            uuid.UUID `json:"id" db:"id"`
    Word          string    `json:"word" db:"word"`
    Pronunciation string    `json:"pronunciation" db:"pronunciation"`
    Definition    string    `json:"definition" db:"definition"`
    Category      string    `json:"category" db:"category"`
    Level         int       `json:"level" db:"level"`
    AudioURL      string    `json:"audio_url" db:"audio_url"`
}

type SpellingAttempt struct {
    GameID    uuid.UUID   `json:"game_id"`
    PlayerID  uuid.UUID   `json:"player_id"`
    Word      string      `json:"word"`
    Type      AttemptType `json:"type"`
    VoiceData []byte      `json:"voice_data,omitempty"`
    Text      string      `json:"text,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

type AttemptType string

const (
    AttemptTypeText  AttemptType = "text"
    AttemptTypeVoice AttemptType = "voice"
)
```

## Database Schema

```sql
-- migrations/001_initial_schema.sql

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    elo INTEGER NOT NULL DEFAULT 1200,
    is_premium BOOLEAN NOT NULL DEFAULT FALSE,
    premium_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Games table
CREATE TABLE games (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    settings JSONB NOT NULL,
    current_word_id UUID REFERENCES words(id),
    current_turn UUID REFERENCES users(id),
    meeting_id TEXT,
    round INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Game participants
CREATE TABLE game_players (
    game_id UUID REFERENCES games(id),
    player_id UUID REFERENCES users(id),
    score INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    correct INTEGER NOT NULL DEFAULT 0,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (game_id, player_id)
);

-- Words table
CREATE TABLE words (
    id UUID PRIMARY KEY,
    word TEXT NOT NULL,
    pronunciation TEXT NOT NULL,
    definition TEXT NOT NULL,
    category TEXT NOT NULL,
    level INTEGER NOT NULL,
    audio_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Tournament table
CREATE TABLE tournaments (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    settings JSONB NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Tournament rooms
CREATE TABLE tournament_rooms (
    tournament_id UUID REFERENCES tournaments(id),
    game_id UUID REFERENCES games(id),
    round INTEGER NOT NULL,
    PRIMARY KEY (tournament_id, game_id)
);

-- Subscription plans
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    price INTEGER NOT NULL,
    interval TEXT NOT NULL,
    features JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User subscriptions
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    plan_id UUID REFERENCES subscription_plans(id),
    status TEXT NOT NULL,
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_games_type ON games(type);
CREATE INDEX idx_words_category ON words(category);
CREATE INDEX idx_words_level ON words(level);
CREATE INDEX idx_tournaments_status ON tournaments(status);
```

Would you like me to continue with:
1. Core game engine implementation
2. Real-time services (AWS Chime, GetStream)
3. AI bot implementation
4. Authentication and payment systems
5. API specifications