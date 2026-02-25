package bluetooth

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var mockDeviceTemplates = []struct {
	Name string
	Type DeviceType
}{
	{"iPhone 15 Pro", DeviceTypeBLE},
	{"Galaxy S24 Ultra", DeviceTypeBLE},
	{"Pixel 9 Pro", DeviceTypeBLE},
	{"AirPods Pro", DeviceTypeBLE},
	{"Galaxy Buds Pro", DeviceTypeClassic},
	{"MacBook Air", DeviceTypeBLE},
	{"Apple Watch", DeviceTypeBLE},
	{"Fitbit Charge 6", DeviceTypeBLE},
	{"Sony WH-1000XM5", DeviceTypeClassic},
	{"JBL Flip 6", DeviceTypeClassic},
	{"Tile Tracker", DeviceTypeBLE},
	{"Tesla Model 3", DeviceTypeBLE},
	{"Nintendo Switch", DeviceTypeClassic},
	{"iPad Pro", DeviceTypeBLE},
	{"OnePlus Buds 3", DeviceTypeBLE},
}

type mockDevice struct {
	mac       string
	name      string
	dtype     DeviceType
	baseRSSI  float64
	phase     float64
	amplitude float64
	active    bool
}

// MockScanner generates fake devices for demo mode.
type MockScanner struct {
	program  *tea.Program
	devices  []mockDevice
	running  bool
	cancel   context.CancelFunc
}

// NewMockScanner creates a mock scanner with random fake devices.
func NewMockScanner() *MockScanner {
	count := 8 + rand.Intn(5) // 8-12 devices
	devices := make([]mockDevice, count)

	perm := rand.Perm(len(mockDeviceTemplates))
	for i := 0; i < count; i++ {
		tmpl := mockDeviceTemplates[perm[i%len(mockDeviceTemplates)]]
		devices[i] = mockDevice{
			mac:       randomMAC(),
			name:      tmpl.Name,
			dtype:     tmpl.Type,
			baseRSSI:  -40 - rand.Float64()*50, // -40 to -90 dBm
			phase:     rand.Float64() * 2 * math.Pi,
			amplitude: 3 + rand.Float64()*8, // 3-11 dBm fluctuation
			active:    true,
		}
	}

	return &MockScanner{devices: devices}
}

// Start begins the mock scanner.
func (s *MockScanner) Start(p *tea.Program) error {
	s.program = p
	s.running = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.loop(ctx)
	return nil
}

func (s *MockScanner) loop(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	t := 0.0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !s.running {
				return
			}
			t += 0.2
			s.emitDevices(t)
		}
	}
}

func (s *MockScanner) emitDevices(t float64) {
	for i := range s.devices {
		d := &s.devices[i]

		// Randomly toggle device visibility (appear/disappear)
		if rand.Float64() < 0.005 {
			d.active = !d.active
		}
		if !d.active {
			continue
		}

		// Sinusoidal RSSI fluctuation + noise
		rssi := d.baseRSSI + d.amplitude*math.Sin(t*0.5+d.phase) + (rand.Float64()-0.5)*4

		name := d.name
		// Some devices occasionally have empty names (realistic)
		if rand.Float64() < 0.05 {
			name = ""
		}

		msg := DeviceDiscoveredMsg{
			MAC:  d.mac,
			Name: name,
			RSSI: int16(rssi),
			Type: d.dtype,
		}
		if s.program != nil {
			s.program.Send(msg)
		}
	}
}

// Stop halts the mock scanner.
func (s *MockScanner) Stop() {
	s.running = false
	if s.cancel != nil {
		s.cancel()
	}
}

func randomMAC() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", b[0], b[1], b[2], b[3], b[4], b[5])
}
