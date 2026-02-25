package app

import "time"

// TickMsg triggers a frame update for animation.
type TickMsg time.Time

// EvictMsg triggers device eviction.
type EvictMsg time.Time

// ScanErrorMsg reports scanner errors.
type ScanErrorMsg struct {
	Err error
}
