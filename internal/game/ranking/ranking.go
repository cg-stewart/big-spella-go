package ranking

import "math"

// Rank represents a player's rank in the game
type Rank struct {
	Color     string
	MinPoints int
	MaxPoints int
}

// Available ranks in ascending order
var Ranks = []Rank{
	{Color: "Gray", MinPoints: 0, MaxPoints: 299},
	{Color: "Violet", MinPoints: 300, MaxPoints: 449},
	{Color: "Indigo", MinPoints: 450, MaxPoints: 599},
	{Color: "Blue", MinPoints: 600, MaxPoints: 749},
	{Color: "Green", MinPoints: 750, MaxPoints: 899},
	{Color: "Yellow", MinPoints: 900, MaxPoints: 1049},
	{Color: "Orange", MinPoints: 1050, MaxPoints: 1149},
	{Color: "Red", MinPoints: 1150, MaxPoints: 1200},
}

// Points awarded for different placements
const (
	GoldPoints   = 30
	SilverPoints = 15
	BronzePoints = 5
)

// CalculatePoints calculates points earned in a game
func CalculatePoints(placement int, playerCount int, isTournament bool) int {
	var basePoints int
	switch placement {
	case 1:
		basePoints = GoldPoints
	case 2:
		basePoints = SilverPoints
	case 3:
		basePoints = BronzePoints
	default:
		return 0
	}

	// Player count multiplier (more players = more points)
	playerMultiplier := 1.0
	if playerCount > 2 {
		playerMultiplier = 1.0 + (float64(playerCount-2) * 0.1) // 10% bonus per additional player
	}

	// Tournament multiplier
	tournamentMultiplier := 1.0
	if isTournament {
		tournamentMultiplier = 1.5 // 50% bonus for tournament games
	}

	return int(math.Round(float64(basePoints) * playerMultiplier * tournamentMultiplier))
}

// GetRankByPoints returns the rank for a given point total
func GetRankByPoints(points int) Rank {
	for _, rank := range Ranks {
		if points >= rank.MinPoints && points <= rank.MaxPoints {
			return rank
		}
	}
	return Ranks[0] // Default to Gray if points are out of range
}

// CalculateNewRating calculates the new rating after a game
func CalculateNewRating(currentRating, pointsEarned int) int {
	newRating := currentRating + pointsEarned
	if newRating > 1200 {
		return 1200
	}
	if newRating < 0 {
		return 0
	}
	return newRating
}
