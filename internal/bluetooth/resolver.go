package bluetooth

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// NameResolver tries to resolve names for unnamed BLE devices in the background.
// It uses hcitool name which sends a name request to the device.
type NameResolver struct {
	program  *tea.Program
	mu       sync.Mutex
	tried    map[string]int // MAC -> attempt count
	resolved map[string]bool
	stop     chan struct{}
}

const (
	maxAttempts    = 2
	resolveTimeout = 4 * time.Second
	resolvePause   = 3 * time.Second
)

// NewNameResolver creates a new resolver.
func NewNameResolver() *NameResolver {
	return &NameResolver{
		tried:    make(map[string]int),
		resolved: make(map[string]bool),
		stop:     make(chan struct{}),
	}
}

// Start begins the resolver background loop.
func (r *NameResolver) Start(p *tea.Program) {
	r.program = p
}

// RequestResolve queues a MAC for background name resolution.
// Safe to call from any goroutine.
func (r *NameResolver) RequestResolve(mac string) {
	r.mu.Lock()
	if r.resolved[mac] || r.tried[mac] >= maxAttempts {
		r.mu.Unlock()
		return
	}
	r.tried[mac]++
	r.mu.Unlock()

	go r.resolve(mac)
}

func (r *NameResolver) resolve(mac string) {
	// Rate limit - don't spam
	time.Sleep(resolvePause)

	select {
	case <-r.stop:
		return
	default:
	}

	name := tryHcitool(mac)
	if name == "" {
		return
	}

	r.mu.Lock()
	r.resolved[mac] = true
	r.mu.Unlock()

	if r.program != nil {
		r.program.Send(DeviceDiscoveredMsg{
			MAC:  mac,
			Name: name,
			RSSI: -100, // placeholder, store will EMA smooth it
			Type: DeviceTypeBLE,
		})
	}
}

func tryHcitool(mac string) string {
	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "hcitool", "name", mac).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Stop terminates the resolver.
func (r *NameResolver) Stop() {
	close(r.stop)
}

// IsResolved returns true if this MAC has been successfully resolved.
func (r *NameResolver) IsResolved(mac string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.resolved[mac]
}

// ShouldResolve returns true if this MAC should be attempted for resolution.
func (r *NameResolver) ShouldResolve(mac string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return !r.resolved[mac] && r.tried[mac] < maxAttempts
}
