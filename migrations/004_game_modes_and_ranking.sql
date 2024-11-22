-- Add game mode support
ALTER TABLE games
    ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'round_robin',
    ADD COLUMN IF NOT EXISTS time_limit INTERVAL,
    ADD COLUMN IF NOT EXISTS max_rounds INTEGER,
    ADD COLUMN IF NOT EXISTS enable_video BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS enable_voice BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS record_game BOOLEAN NOT NULL DEFAULT false;

-- Add ranking support
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS rank_points INTEGER NOT NULL DEFAULT 1200,
    ADD COLUMN IF NOT EXISTS rank_color TEXT NOT NULL DEFAULT 'Gray',
    ADD COLUMN IF NOT EXISTS games_won INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS games_played INTEGER NOT NULL DEFAULT 0;

-- Add game results tracking
CREATE TABLE IF NOT EXISTS game_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID REFERENCES games(id) ON DELETE CASCADE,
    player_id UUID REFERENCES users(id) ON DELETE CASCADE,
    placement INTEGER NOT NULL,
    points_earned INTEGER NOT NULL,
    previous_rank_points INTEGER NOT NULL,
    new_rank_points INTEGER NOT NULL,
    previous_rank_color TEXT NOT NULL,
    new_rank_color TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add game recording metadata
CREATE TABLE IF NOT EXISTS game_recordings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID REFERENCES games(id) ON DELETE CASCADE,
    s3_key TEXT NOT NULL,
    duration INTERVAL NOT NULL,
    size_bytes BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_game_results_game_id ON game_results(game_id);
CREATE INDEX IF NOT EXISTS idx_game_results_player_id ON game_results(player_id);
CREATE INDEX IF NOT EXISTS idx_game_recordings_game_id ON game_recordings(game_id);
CREATE INDEX IF NOT EXISTS idx_users_rank_points ON users(rank_points DESC);

-- Add trigger to update game_recordings.updated_at
CREATE TRIGGER update_game_recordings_updated_at
    BEFORE UPDATE ON game_recordings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
