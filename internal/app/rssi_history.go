package app

// RSSIRing is a circular buffer for RSSI history values.
type RSSIRing struct {
	buf   []float64
	pos   int
	count int
}

// NewRSSIRing creates a new circular buffer with the given capacity.
func NewRSSIRing(capacity int) *RSSIRing {
	return &RSSIRing{
		buf: make([]float64, capacity),
	}
}

// Push adds a value to the ring buffer.
func (r *RSSIRing) Push(val float64) {
	r.buf[r.pos] = val
	r.pos = (r.pos + 1) % len(r.buf)
	if r.count < len(r.buf) {
		r.count++
	}
}

// Values returns all stored values in chronological order.
func (r *RSSIRing) Values() []float64 {
	if r.count == 0 {
		return nil
	}
	result := make([]float64, r.count)
	if r.count < len(r.buf) {
		copy(result, r.buf[:r.count])
	} else {
		start := r.pos
		n := copy(result, r.buf[start:])
		copy(result[n:], r.buf[:start])
	}
	return result
}

// Last returns the most recent value, or 0 if empty.
func (r *RSSIRing) Last() float64 {
	if r.count == 0 {
		return 0
	}
	idx := (r.pos - 1 + len(r.buf)) % len(r.buf)
	return r.buf[idx]
}

// Len returns the number of stored values.
func (r *RSSIRing) Len() int {
	return r.count
}
