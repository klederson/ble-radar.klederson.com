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
	{"HomeNetwork_2G", DeviceTypeWiFi},
	{"XFINITY-7A3F", DeviceTypeWiFi},
	{"TP-Link_5GHz", DeviceTypeWiFi},
	{"AndroidAP", DeviceTypeWiFi},
	{"Starlink_WiFi", DeviceTypeWiFi},
}

type mockDevice struct {
	mac       string
	name      string
	dtype     DeviceType
	baseRSSI  float64
	phase     float64
	amplitude float64
	active    bool
	freq      int
	channel   int
}

// MockScanner generates fake devices for demo mode.
type MockScanner struct {
	program  *tea.Program
	devices  []mockDevice
	running  bool
	cancel   context.CancelFunc
}

// 5 GHz channel options for mock WiFi devices.
var wifi5GChannels = []int{36, 40, 44, 48, 149, 153, 157, 161}

// NewMockScanner creates a mock scanner with random fake devices.
func NewMockScanner() *MockScanner {
	// Separate templates by type to guarantee representation
	var bleTmpls, clsTmpls, wifiTmpls []int
	for i, t := range mockDeviceTemplates {
		switch t.Type {
		case DeviceTypeBLE:
			bleTmpls = append(bleTmpls, i)
		case DeviceTypeClassic:
			clsTmpls = append(clsTmpls, i)
		case DeviceTypeWiFi:
			wifiTmpls = append(wifiTmpls, i)
		}
	}

	// Pick guaranteed minimums from each type
	var picked []int
	blePerm := rand.Perm(len(bleTmpls))
	for i := 0; i < 5 && i < len(blePerm); i++ {
		picked = append(picked, bleTmpls[blePerm[i]])
	}
	clsPerm := rand.Perm(len(clsTmpls))
	for i := 0; i < 2 && i < len(clsPerm); i++ {
		picked = append(picked, clsTmpls[clsPerm[i]])
	}
	wifiPerm := rand.Perm(len(wifiTmpls))
	for i := 0; i < 3 && i < len(wifiPerm); i++ {
		picked = append(picked, wifiTmpls[wifiPerm[i]])
	}

	// Add a few more random ones up to 12-15 total
	total := 12 + rand.Intn(4)
	allPerm := rand.Perm(len(mockDeviceTemplates))
	usedSet := make(map[int]bool, len(picked))
	for _, p := range picked {
		usedSet[p] = true
	}
	for _, idx := range allPerm {
		if len(picked) >= total {
			break
		}
		if !usedSet[idx] {
			picked = append(picked, idx)
			usedSet[idx] = true
		}
	}

	devices := make([]mockDevice, len(picked))
	for i, ti := range picked {
		tmpl := mockDeviceTemplates[ti]
		md := mockDevice{
			mac:       randomMAC(),
			name:      tmpl.Name,
			dtype:     tmpl.Type,
			baseRSSI:  -40 - rand.Float64()*50, // -40 to -90 dBm
			phase:     rand.Float64() * 2 * math.Pi,
			amplitude: 3 + rand.Float64()*8, // 3-11 dBm fluctuation
			active:    true,
		}
		if tmpl.Type == DeviceTypeWiFi {
			if rand.Intn(2) == 0 {
				// 2.4 GHz
				md.freq = 2412 + rand.Intn(11)*5
				md.channel = (md.freq - 2407) / 5
			} else {
				// 5 GHz
				ch := wifi5GChannels[rand.Intn(len(wifi5GChannels))]
				md.channel = ch
				md.freq = 5000 + ch*5
			}
		}
		devices[i] = md
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
			MAC:       d.mac,
			Name:      name,
			RSSI:      int16(rssi),
			Type:      d.dtype,
			Frequency: d.freq,
			Channel:   d.channel,
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
