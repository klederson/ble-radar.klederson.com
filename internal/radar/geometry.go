package radar

import (
	"math"

	"ble-radar.klederson.com/internal/config"
)

// CellDistance computes the distance from a cell to the radar center,
// accounting for terminal aspect ratio.
func CellDistance(col, row, centerX, centerY int) float64 {
	dx := float64(col - centerX)
	dy := float64(row-centerY) / config.AspectRatio
	return math.Sqrt(dx*dx + dy*dy)
}

// CellAngle computes the angle from center to a cell.
// Returns radians in [0, 2π), where 0=north, increasing clockwise.
func CellAngle(col, row, centerX, centerY int) float64 {
	dx := float64(col - centerX)
	dy := float64(row-centerY) / config.AspectRatio
	angle := math.Atan2(dx, -dy) // 0=north, clockwise
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}

// RingChar returns the appropriate character for a ring at the given angle.
func RingChar(angle float64) rune {
	// Normalize angle to [0, 2π)
	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}

	// 8 sectors for character selection
	sector := int(math.Round(angle/(math.Pi/4))) % 8

	switch sector {
	case 0: // North
		return '-'
	case 1: // NE
		return '/'
	case 2: // East
		return '|'
	case 3: // SE
		return '\\'
	case 4: // South
		return '-'
	case 5: // SW
		return '/'
	case 6: // West
		return '|'
	case 7: // NW
		return '\\'
	default:
		return '.'
	}
}

// NormalizeAngle wraps an angle to [0, 2π).
func NormalizeAngle(a float64) float64 {
	for a < 0 {
		a += 2 * math.Pi
	}
	for a >= 2*math.Pi {
		a -= 2 * math.Pi
	}
	return a
}

// AngleDiff returns the shortest angular distance between two angles.
// Result is in [0, π].
func AngleDiff(a, b float64) float64 {
	d := math.Abs(NormalizeAngle(a) - NormalizeAngle(b))
	if d > math.Pi {
		d = 2*math.Pi - d
	}
	return d
}

// MetersToRadius converts distance in meters to radar units (pixels/cells).
func MetersToRadius(meters, maxRange, radarRadius float64) float64 {
	if meters > maxRange {
		return radarRadius
	}
	return (meters / maxRange) * radarRadius
}
