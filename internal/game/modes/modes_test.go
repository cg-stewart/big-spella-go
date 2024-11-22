package modes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettings(t *testing.T) {
	tests := []struct {
		name string
		mode GameMode
		want GameSettings
	}{
		{
			name: "Round Robin defaults",
			mode: ModeRoundRobin,
			want: GameSettings{
				Mode:       ModeRoundRobin,
				MaxPlayers: 32,
				MaxRounds:  10,
				WordLevel:  1,
				EnableVideo: true,
				EnableVoice: true,
				RecordGame: false,
			},
		},
		{
			name: "Rapid Fire defaults",
			mode: ModeRapidFire,
			want: GameSettings{
				Mode:       ModeRapidFire,
				MaxPlayers: 2,
				TimeLimit:  10 * time.Minute,
				WordLevel:  1,
				EnableVideo: true,
				EnableVoice: true,
				RecordGame: false,
			},
		},
		{
			name: "Total Game defaults",
			mode: ModeTotalGame,
			want: GameSettings{
				Mode:       ModeTotalGame,
				MaxPlayers: 8,
				MaxRounds:  20,
				TimeLimit:  30 * time.Minute,
				WordLevel:  1,
				EnableVideo: true,
				EnableVoice: true,
				RecordGame: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultSettings(tt.mode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings GameSettings
		wantErr  bool
	}{
		{
			name: "Valid Round Robin",
			settings: GameSettings{
				Mode:       ModeRoundRobin,
				MaxPlayers: 16,
				MaxRounds:  5,
				WordLevel:  5,
			},
			wantErr: false,
		},
		{
			name: "Invalid Round Robin players",
			settings: GameSettings{
				Mode:       ModeRoundRobin,
				MaxPlayers: 40,
				MaxRounds:  5,
				WordLevel:  5,
			},
			wantErr: true,
		},
		{
			name: "Valid Rapid Fire",
			settings: GameSettings{
				Mode:       ModeRapidFire,
				MaxPlayers: 2,
				TimeLimit:  15 * time.Minute,
				WordLevel:  3,
			},
			wantErr: false,
		},
		{
			name: "Invalid Rapid Fire players",
			settings: GameSettings{
				Mode:       ModeRapidFire,
				MaxPlayers: 3,
				TimeLimit:  15 * time.Minute,
				WordLevel:  3,
			},
			wantErr: true,
		},
		{
			name: "Valid Total Game",
			settings: GameSettings{
				Mode:       ModeTotalGame,
				MaxPlayers: 6,
				TimeLimit:  30 * time.Minute,
				WordLevel:  7,
			},
			wantErr: false,
		},
		{
			name: "Invalid word level",
			settings: GameSettings{
				Mode:       ModeTotalGame,
				MaxPlayers: 6,
				TimeLimit:  30 * time.Minute,
				WordLevel:  11,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSettings(tt.settings)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name            string
		mode            GameMode
		correctAttempts int
		totalAttempts   int
		averageTime     float64
		expected        int
	}{
		{
			name:            "Round Robin normal score",
			mode:            ModeRoundRobin,
			correctAttempts: 5,
			totalAttempts:   7,
			averageTime:     6.0,
			expected:        500,
		},
		{
			name:            "Rapid Fire fast answers",
			mode:            ModeRapidFire,
			correctAttempts: 5,
			totalAttempts:   6,
			averageTime:     3.0,
			expected:        750, // 500 * 1.5 for speed bonus
		},
		{
			name:            "Total Game high accuracy",
			mode:            ModeTotalGame,
			correctAttempts: 9,
			totalAttempts:   10,
			averageTime:     7.0,
			expected:        1170, // 900 * 1.3 for accuracy bonus
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateScore(tt.mode, tt.correctAttempts, tt.totalAttempts, tt.averageTime)
			assert.Equal(t, tt.expected, score)
		})
	}
}
