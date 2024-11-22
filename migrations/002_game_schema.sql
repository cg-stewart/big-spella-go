-- Create words table
CREATE TABLE IF NOT EXISTS words (
    id UUID PRIMARY KEY,
    word TEXT NOT NULL,
    pronunciation TEXT NOT NULL,
    definition TEXT NOT NULL,
    category TEXT NOT NULL,
    level INTEGER NOT NULL,
    audio_url TEXT,
    example_sentence TEXT,
    etymology TEXT,
    part_of_speech TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(word)
);

-- Create games table
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    settings JSONB NOT NULL,
    current_word_id UUID REFERENCES words(id),
    current_turn UUID,
    meeting_id TEXT,
    round INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    turn_started_at TIMESTAMP WITH TIME ZONE,
    hints_used JSONB,
    word_masked BOOLEAN NOT NULL DEFAULT true
);

-- Create players table
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id),
    player_id UUID NOT NULL REFERENCES users(id),
    score INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    attempts INTEGER NOT NULL DEFAULT 0,
    correct INTEGER NOT NULL DEFAULT 0,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(game_id, player_id)
);

-- Create spelling_attempts table
CREATE TABLE IF NOT EXISTS spelling_attempts (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id),
    player_id UUID NOT NULL REFERENCES players(id),
    word TEXT NOT NULL,
    type TEXT NOT NULL,
    voice_data BYTEA,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_words_level_category ON words(level, category);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_players_game_id ON players(game_id);
CREATE INDEX IF NOT EXISTS idx_players_joined_at ON players(joined_at);
CREATE INDEX IF NOT EXISTS idx_spelling_attempts_game_id ON spelling_attempts(game_id);
CREATE INDEX IF NOT EXISTS idx_spelling_attempts_player_id ON spelling_attempts(player_id);
