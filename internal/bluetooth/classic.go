package bluetooth

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ClassicScanner discovers classic Bluetooth devices via hcitool.
type ClassicScanner struct {
	program  *tea.Program
	running  bool
	cancel   context.CancelFunc
	interval time.Duration
}

// NewClassicScanner creates a classic BT scanner.
func NewClassicScanner(interval time.Duration) *ClassicScanner {
	return &ClassicScanner{
		interval: interval,
	}
}

// Start begins periodic classic BT scans in a goroutine.
func (s *ClassicScanner) Start(p *tea.Program) error {
	s.program = p
	s.running = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.loop(ctx)
	return nil
}

func (s *ClassicScanner) loop(ctx context.Context) {
	for {
		if !s.running {
			return
		}
		s.scan()
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.interval):
		}
	}
}

func (s *ClassicScanner) scan() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "hcitool", "scan", "--flush")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Scanning") {
			continue
		}
		// Format: "AA:BB:CC:DD:EE:FF	Device Name"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 1 {
			continue
		}
		mac := strings.TrimSpace(parts[0])
		name := ""
		if len(parts) >= 2 {
			name = strings.TrimSpace(parts[1])
		}
		if !isValidMAC(mac) {
			continue
		}

		msg := DeviceDiscoveredMsg{
			MAC:  mac,
			Name: name,
			RSSI: -75, // hcitool scan doesn't provide RSSI; use default
			Type: DeviceTypeClassic,
		}
		if s.program != nil {
			s.program.Send(msg)
		}
	}

	_ = cmd.Wait()
}

// Stop halts the classic scanner.
func (s *ClassicScanner) Stop() {
	s.running = false
	if s.cancel != nil {
		s.cancel()
	}
}

func isValidMAC(mac string) bool {
	if len(mac) != 17 {
		return false
	}
	for i, c := range mac {
		if (i+1)%3 == 0 {
			if c != ':' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
	}
	return true
}

// ClassicScannerAvailable checks if hcitool is available on the system.
func ClassicScannerAvailable() bool {
	_, err := exec.LookPath("hcitool")
	return err == nil
}

// ClassicScanErrorMsg reports hcitool errors.
type ClassicScanErrorMsg struct {
	Err error
}

func (e ClassicScanErrorMsg) Error() string {
	return fmt.Sprintf("classic scan error: %v", e.Err)
}
