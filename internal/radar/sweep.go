package radar

import (
	"math"
	"time"

	"ble-radar.klederson.com/internal/config"
)

// Sweep manages the rotating sweep line state.
type Sweep struct {
	Angle     float64 // Current angle in radians [0, 2π)
	StartTime time.Time
}

// NewSweep creates a new sweep starting at 0 degrees (north).
func NewSweep() *Sweep {
	return &Sweep{
		Angle:     0,
		StartTime: time.Now(),
	}
}

// Update advances the sweep angle based on elapsed time.
func (s *Sweep) Update() {
	elapsed := time.Since(s.StartTime).Seconds()
	rps := float64(config.SweepSpeedRPM) / 60.0 // rotations per second
	s.Angle = math.Mod(elapsed*rps*2*math.Pi, 2*math.Pi)
}

// Degrees returns the current sweep angle in degrees.
func (s *Sweep) Degrees() float64 {
	return s.Angle * 180 / math.Pi
}

// Intensity returns the glow intensity [0, 1] for a given cell angle.
// The sweep has a trailing glow of SweepTrailDeg degrees.
// Returns 0 if the cell is outside the sweep trail.
func (s *Sweep) Intensity(cellAngle float64) float64 {
	// Calculate how far behind the sweep this angle is
	diff := NormalizeAngle(s.Angle - cellAngle)
	if diff < 0 {
		diff += 2 * math.Pi
	}

	trailRad := config.SweepTrailDeg * math.Pi / 180.0
	if diff > trailRad {
		return 0
	}

	// Linear falloff: 1.0 at sweep head → 0.0 at trail end
	return 1.0 - diff/trailRad
}
