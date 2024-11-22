package ranking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePoints(t *testing.T) {
	tests := []struct {
		name         string
		place        int
		players      int
		isTournament bool
		expected     int
	}{
		{
			name:         "First place in tournament",
			place:        1,
			players:      8,
			isTournament: true,
			expected:     72,
		},
		{
			name:         "Third place in large game",
			place:        3,
			players:      12,
			isTournament: false,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := CalculatePoints(tt.place, tt.players, tt.isTournament)
			assert.Equal(t, tt.expected, points)
		})
	}
}

func TestGetRankByPoints(t *testing.T) {
	tests := []struct {
		name          string
		points        int
		expectedColor string
	}{
		{"Starter rank", 0, "Gray"},
		{"Mid violet", 350, "Violet"},
		{"High indigo", 550, "Indigo"},
		{"Low blue", 600, "Blue"},
		{"Mid green", 800, "Green"},
		{"High yellow", 1000, "Yellow"},
		{"Low orange", 1050, "Orange"},
		{"Max red", 1200, "Red"},
		{"Over max", 1300, "Gray"}, // Should default to Gray if out of range
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank := GetRankByPoints(tt.points)
			assert.Equal(t, tt.expectedColor, rank.Color)
		})
	}
}

func TestCalculateNewRating(t *testing.T) {
	tests := []struct {
		name          string
		currentRating int
		pointsEarned  int
		expected      int
	}{
		{"Normal increase", 1000, 30, 1030},
		{"Hit max cap", 1190, 30, 1200},
		{"Normal decrease", 1000, -15, 985},
		{"Hit min cap", 10, -20, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRating := CalculateNewRating(tt.currentRating, tt.pointsEarned)
			assert.Equal(t, tt.expected, newRating)
		})
	}
}
