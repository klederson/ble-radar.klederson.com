package bluetooth

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"tinygo.org/x/bluetooth"
)

// DeviceDiscoveredMsg is sent via tea.Program.Send when a device is found.
type DeviceDiscoveredMsg struct {
	MAC  string
	Name string
	RSSI int16
	Type DeviceType
}

// BLEScanner handles Bluetooth Low Energy scanning.
type BLEScanner struct {
	adapter *bluetooth.Adapter
	program *tea.Program
	running bool
}

// NewBLEScanner creates a scanner for the given adapter name (e.g., "hci0").
func NewBLEScanner() *BLEScanner {
	return &BLEScanner{
		adapter: bluetooth.DefaultAdapter,
	}
}

// Start begins BLE scanning in a goroutine. Discovered devices are sent
// as tea messages via program.Send().
func (s *BLEScanner) Start(p *tea.Program) error {
	s.program = p

	if err := s.adapter.Enable(); err != nil {
		return fmt.Errorf("failed to enable BLE adapter: %w (try running with sudo or setcap cap_net_admin+ep)", err)
	}

	s.running = true
	go func() {
		_ = s.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if !s.running {
				return
			}
			msg := DeviceDiscoveredMsg{
				MAC:  result.Address.String(),
				Name: result.LocalName(),
				RSSI: result.RSSI,
				Type: DeviceTypeBLE,
			}
			if s.program != nil {
				s.program.Send(msg)
			}
		})
	}()

	return nil
}

// Stop halts the BLE scanner.
func (s *BLEScanner) Stop() {
	s.running = false
	_ = s.adapter.StopScan()
}
