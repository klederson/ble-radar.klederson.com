package bluetooth

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// WiFiScanner discovers nearby WiFi access points.
// Prefers nmcli (no root needed), falls back to iw (needs root).
type WiFiScanner struct {
	program  *tea.Program
	iface    string
	running  bool
	cancel   context.CancelFunc
	interval time.Duration
	useNmcli bool
}

// NewWiFiScanner creates a WiFi scanner. If iface is empty, auto-detects.
func NewWiFiScanner(iface string, interval time.Duration) *WiFiScanner {
	useNmcli := nmcliAvailable()
	if iface == "" && !useNmcli {
		iface = detectWiFiInterface()
	}
	return &WiFiScanner{
		iface:    iface,
		interval: interval,
		useNmcli: useNmcli,
	}
}

// Start begins periodic WiFi scans in a goroutine.
func (s *WiFiScanner) Start(p *tea.Program) error {
	s.program = p
	s.running = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.loop(ctx)
	return nil
}

func (s *WiFiScanner) loop(ctx context.Context) {
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

func (s *WiFiScanner) scan() {
	var msgs []DeviceDiscoveredMsg
	if s.useNmcli {
		msgs = s.scanNmcli()
	} else {
		msgs = s.scanIW()
	}
	for _, msg := range msgs {
		if s.program != nil {
			s.program.Send(msg)
		}
	}
}

// scanNmcli uses nmcli (works without root).
func (s *WiFiScanner) scanNmcli() []DeviceDiscoveredMsg {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use cached results from NetworkManager (it rescans automatically).
	// Calling rescan here causes flicker as the cache clears momentarily.
	cmd := exec.CommandContext(ctx, "nmcli", "-t", "-f", "BSSID,SSID,FREQ,CHAN,SIGNAL", "dev", "wifi", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	return parseNmcliScan(string(out))
}

// parseNmcliScan parses nmcli terse output.
// Format per line: BSSID:SSID:FREQ:CHAN:SIGNAL
// In terse mode, literal colons in values are escaped as \:
func parseNmcliScan(output string) []DeviceDiscoveredMsg {
	var results []DeviceDiscoveredMsg

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Split on unescaped colons: replace \: with placeholder, split, restore
		const placeholder = "\x00"
		escaped := strings.ReplaceAll(line, `\:`, placeholder)
		parts := strings.Split(escaped, ":")
		// Restore colons in each part
		for i := range parts {
			parts[i] = strings.ReplaceAll(parts[i], placeholder, ":")
		}

		if len(parts) < 5 {
			continue
		}

		bssid := strings.TrimSpace(parts[0])
		ssid := strings.TrimSpace(parts[1])
		freqStr := strings.TrimSpace(parts[2])
		chanStr := strings.TrimSpace(parts[3])
		sigStr := strings.TrimSpace(parts[4])

		mac := strings.ToUpper(bssid)
		if !isValidMAC(mac) {
			continue
		}

		freq, _ := strconv.Atoi(strings.TrimSuffix(freqStr, " MHz"))
		channel, _ := strconv.Atoi(chanStr)

		signal, err := strconv.Atoi(sigStr)
		rssi := int16(-80) // default
		if err == nil {
			// nmcli SIGNAL is 0-100 percentage; convert to approximate dBm
			// 100% ~ -30dBm, 0% ~ -100dBm
			rssi = int16(-100 + signal*70/100)
		}

		results = append(results, DeviceDiscoveredMsg{
			MAC:       mac,
			Name:      ssid,
			RSSI:      rssi,
			Type:      DeviceTypeWiFi,
			Frequency: freq,
			Channel:   channel,
		})
	}

	return results
}

// scanIW uses iw (requires root).
func (s *WiFiScanner) scanIW() []DeviceDiscoveredMsg {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "iw", "dev", s.iface, "scan")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	return parseIWScan(string(out))
}

// parseIWScan parses the output of `iw dev <iface> scan`.
func parseIWScan(output string) []DeviceDiscoveredMsg {
	var results []DeviceDiscoveredMsg

	scanner := bufio.NewScanner(strings.NewReader(output))

	var current *DeviceDiscoveredMsg
	for scanner.Scan() {
		line := scanner.Text()

		// New BSS block: "BSS aa:bb:cc:dd:ee:ff(on wlan0)"
		if strings.HasPrefix(line, "BSS ") {
			if current != nil && isValidMAC(current.MAC) {
				results = append(results, *current)
			}
			mac := strings.TrimPrefix(line, "BSS ")
			if idx := strings.IndexByte(mac, '('); idx >= 0 {
				mac = mac[:idx]
			}
			mac = strings.TrimSpace(mac)
			mac = strings.ToUpper(mac)
			current = &DeviceDiscoveredMsg{
				MAC:  mac,
				RSSI: -80,
				Type: DeviceTypeWiFi,
			}
			continue
		}

		if current == nil {
			continue
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "SSID: ") {
			current.Name = strings.TrimPrefix(trimmed, "SSID: ")
		} else if strings.HasPrefix(trimmed, "freq: ") {
			if v, err := strconv.Atoi(strings.TrimPrefix(trimmed, "freq: ")); err == nil {
				current.Frequency = v
			}
		} else if strings.HasPrefix(trimmed, "signal: ") {
			sigStr := strings.TrimPrefix(trimmed, "signal: ")
			sigStr = strings.TrimSuffix(sigStr, " dBm")
			sigStr = strings.TrimSpace(sigStr)
			if v, err := strconv.ParseFloat(sigStr, 64); err == nil {
				current.RSSI = int16(v)
			}
		} else if strings.HasPrefix(trimmed, "DS Parameter set: channel ") {
			chStr := strings.TrimPrefix(trimmed, "DS Parameter set: channel ")
			if v, err := strconv.Atoi(chStr); err == nil {
				current.Channel = v
			}
		} else if strings.HasPrefix(trimmed, "primary channel: ") && current.Channel == 0 {
			chStr := strings.TrimPrefix(trimmed, "primary channel: ")
			if v, err := strconv.Atoi(chStr); err == nil {
				current.Channel = v
			}
		}
	}

	if current != nil && isValidMAC(current.MAC) {
		results = append(results, *current)
	}

	return results
}

// Stop halts the WiFi scanner.
func (s *WiFiScanner) Stop() {
	s.running = false
	if s.cancel != nil {
		s.cancel()
	}
}

// WiFiScannerAvailable checks if nmcli or iw is available on the system.
func WiFiScannerAvailable() bool {
	return nmcliAvailable() || iwAvailable()
}

func nmcliAvailable() bool {
	_, err := exec.LookPath("nmcli")
	return err == nil
}

func iwAvailable() bool {
	_, err := exec.LookPath("iw")
	return err == nil
}

// detectWiFiInterface finds the first wireless interface via `iw dev`.
func detectWiFiInterface() string {
	out, err := exec.Command("iw", "dev").Output()
	if err != nil {
		return "wlan0"
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Interface ") {
			return strings.TrimPrefix(line, "Interface ")
		}
	}
	return "wlan0"
}
