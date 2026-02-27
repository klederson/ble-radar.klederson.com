package app

import (
	"strings"
	"time"

	"ble-radar.klederson.com/internal/bluetooth"
	"ble-radar.klederson.com/internal/config"
	"ble-radar.klederson.com/internal/radar"
	"ble-radar.klederson.com/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// shared holds state shared between the Bubble Tea model copies and main.go.
// Because Bubble Tea uses value receivers, pointer fields ensure all copies
// see the same underlying data.
type shared struct {
	store          *bluetooth.DeviceStore
	sweep          *radar.Sweep
	bleScanner     *bluetooth.BLEScanner
	classicScanner *bluetooth.ClassicScanner
	wifiScanner    *bluetooth.WiFiScanner
	mockScanner    *bluetooth.MockScanner
	resolver       *bluetooth.NameResolver
	hiddenDevices  map[string]bool
	rssiHistory    map[string]*RSSIRing
}

// AppModel is the root Bubble Tea model for BLE Radar.
type AppModel struct {
	width  int
	height int

	scanning    bool
	demoMode    bool
	adapter     string
	cursorIndex int
	selectedMAC string
	detailOpen  bool
	isolateMAC  string

	// Filter state
	filterBLE     bool
	filterClassic bool
	filterWiFi    bool
	filterSearch  string
	filterActive  bool
	filteredView  []*bluetooth.Device

	shared *shared

	// Cached snapshot
	devices []*bluetooth.Device
}

// New creates a new AppModel.
func New(demoMode bool, adapter string) AppModel {
	return AppModel{
		scanning:      true,
		demoMode:      demoMode,
		adapter:       adapter,
		filterBLE:     true,
		filterClassic: true,
		filterWiFi:    true,
		shared: &shared{
			store:         bluetooth.NewDeviceStore(),
			sweep:         radar.NewSweep(),
			resolver:      bluetooth.NewNameResolver(),
			hiddenDevices: make(map[string]bool),
			rssiHistory:   make(map[string]*RSSIRing),
		},
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		evictCmd(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case TickMsg:
		m.shared.sweep.Update()
		m.devices = m.shared.store.Snapshot()
		m.filteredView = m.filteredDevices()

		// Record RSSI history
		for _, d := range m.devices {
			ring, ok := m.shared.rssiHistory[d.MAC]
			if !ok {
				ring = NewRSSIRing(60)
				m.shared.rssiHistory[d.MAC] = ring
			}
			ring.Push(d.RSSI)
		}

		// Request name resolution for unnamed devices (real mode only)
		if !m.demoMode {
			for _, d := range m.devices {
				if d.Name == "" && m.shared.resolver.ShouldResolve(d.MAC) {
					m.shared.resolver.RequestResolve(d.MAC)
				}
			}
		}

		// Cursor stability: re-find selectedMAC after re-sort
		if m.selectedMAC != "" {
			found := false
			for i, d := range m.filteredView {
				if d.MAC == m.selectedMAC {
					m.cursorIndex = i
					found = true
					break
				}
			}
			if !found {
				m.clampCursor()
			}
		} else {
			m.clampCursor()
		}

		// Auto-close detail if device gone
		if m.detailOpen && (len(m.filteredView) == 0 || m.cursorIndex >= len(m.filteredView)) {
			m.detailOpen = false
		}

		return m, tickCmd()

	case EvictMsg:
		m.shared.store.Evict(config.DeviceTimeout)

		// Clean up stale entries
		snap := m.shared.store.Snapshot()
		active := make(map[string]bool, len(snap))
		for _, d := range snap {
			active[d.MAC] = true
		}
		for mac := range m.shared.rssiHistory {
			if !active[mac] {
				delete(m.shared.rssiHistory, mac)
			}
		}
		for mac := range m.shared.hiddenDevices {
			if !active[mac] {
				delete(m.shared.hiddenDevices, mac)
			}
		}
		if m.isolateMAC != "" && !active[m.isolateMAC] {
			m.isolateMAC = ""
		}

		return m, evictCmd()

	case bluetooth.DeviceDiscoveredMsg:
		if m.scanning {
			m.shared.store.Upsert(msg.MAC, msg.Name, float64(msg.RSSI), msg.Type, msg.Frequency, msg.Channel)
		}
		return m, nil

	case ScanErrorMsg:
		return m, nil
	}

	return m, nil
}

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filterActive {
		return m.handleKeyFilter(msg)
	}
	if m.detailOpen {
		return m.handleKeyDetail(msg)
	}
	return m.handleKeyNormal(msg)
}

func (m AppModel) handleKeyNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "Q", "ctrl+c":
		m.stopScanners()
		return m, tea.Quit

	case "s", "S":
		if !m.scanning {
			m.scanning = true
		}

	case "p", "P":
		m.scanning = false

	case "up", "k":
		if m.cursorIndex > 0 {
			m.cursorIndex--
			m.syncSelectedMAC()
		}

	case "down", "j":
		if m.cursorIndex < len(m.filteredView)-1 {
			m.cursorIndex++
			m.syncSelectedMAC()
		}

	case "home":
		m.cursorIndex = 0
		m.syncSelectedMAC()

	case "end":
		if len(m.filteredView) > 0 {
			m.cursorIndex = len(m.filteredView) - 1
			m.syncSelectedMAC()
		}

	case "enter":
		if len(m.filteredView) > 0 && m.cursorIndex < len(m.filteredView) {
			m.detailOpen = true
		}

	case "shift+enter", "I":
		// Toggle isolate mode
		if len(m.filteredView) > 0 && m.cursorIndex < len(m.filteredView) {
			mac := m.filteredView[m.cursorIndex].MAC
			if m.isolateMAC == mac {
				m.isolateMAC = ""
			} else {
				m.isolateMAC = mac
			}
		}

	case " ", "v":
		// Toggle device visibility on radar
		if len(m.filteredView) > 0 && m.cursorIndex < len(m.filteredView) {
			mac := m.filteredView[m.cursorIndex].MAC
			if m.shared.hiddenDevices[mac] {
				delete(m.shared.hiddenDevices, mac)
			} else {
				m.shared.hiddenDevices[mac] = true
			}
		}

	case "1":
		m.filterBLE = !m.filterBLE
		m.refreshFilter()

	case "2":
		m.filterClassic = !m.filterClassic
		m.refreshFilter()

	case "3":
		m.filterWiFi = !m.filterWiFi
		m.refreshFilter()

	case "/":
		m.filterActive = true
	}

	return m, nil
}

func (m AppModel) handleKeyDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "Q", "ctrl+c":
		m.stopScanners()
		return m, tea.Quit

	case "esc", "enter":
		m.detailOpen = false

	case "up", "k":
		if m.cursorIndex > 0 {
			m.cursorIndex--
			m.syncSelectedMAC()
		}

	case "down", "j":
		if m.cursorIndex < len(m.filteredView)-1 {
			m.cursorIndex++
			m.syncSelectedMAC()
		}
	}

	return m, nil
}

func (m *AppModel) syncSelectedMAC() {
	if m.cursorIndex >= 0 && m.cursorIndex < len(m.filteredView) {
		m.selectedMAC = m.filteredView[m.cursorIndex].MAC
	}
}

func (m *AppModel) clampCursor() {
	if len(m.filteredView) == 0 {
		m.cursorIndex = 0
		m.selectedMAC = ""
		return
	}
	if m.cursorIndex >= len(m.filteredView) {
		m.cursorIndex = len(m.filteredView) - 1
	}
	if m.cursorIndex < 0 {
		m.cursorIndex = 0
	}
	m.selectedMAC = m.filteredView[m.cursorIndex].MAC
}

// visibleDevices returns the devices that should appear on the radar.
func (m AppModel) visibleDevices() []*bluetooth.Device {
	base := m.filteredView

	if m.isolateMAC != "" {
		for _, d := range base {
			if d.MAC == m.isolateMAC {
				return []*bluetooth.Device{d}
			}
		}
		return nil
	}

	if len(m.shared.hiddenDevices) == 0 {
		return base
	}

	result := make([]*bluetooth.Device, 0, len(base))
	for _, d := range base {
		if !m.shared.hiddenDevices[d.MAC] {
			result = append(result, d)
		}
	}
	return result
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing BLE Radar..."
	}

	menuH := 1
	statusH := 1
	bodyH := m.height - menuH - statusH
	if bodyH < 5 {
		bodyH = 5
	}

	radarW := m.width * 3 / 4
	if radarW < 30 {
		radarW = 30
	}
	listW := m.width - radarW
	if listW < 15 {
		listW = 15
		radarW = m.width - listW
	}

	menuBar := ui.RenderMenuBar(m.width, m.adapter, m.scanning, m.detailOpen, m.filterActive)

	var leftPanel string
	if m.detailOpen && m.cursorIndex >= 0 && m.cursorIndex < len(m.filteredView) {
		d := m.filteredView[m.cursorIndex]
		var history []float64
		if ring, ok := m.shared.rssiHistory[d.MAC]; ok {
			history = ring.Values()
		}
		leftPanel = ui.RenderDetailPanel(d, radarW, bodyH, history)
	} else {
		innerW := radarW - 4
		innerH := bodyH - 4
		if innerW < 5 {
			innerW = 5
		}
		if innerH < 3 {
			innerH = 3
		}
		radarContent := radar.Render(innerW, innerH, m.visibleDevices(), m.shared.sweep)
		legend := radar.RenderLegend(innerW)
		leftPanel = ui.RenderRadarPanel(radarW, bodyH, radarContent, legend)
	}

	filter := ui.FilterState{
		BLE:     m.filterBLE,
		Classic: m.filterClassic,
		WiFi:    m.filterWiFi,
		Search:  m.filterSearch,
		Active:  m.filterActive,
	}
	deviceList := ui.RenderDeviceList(m.filteredView, listW, bodyH, m.cursorIndex, m.shared.hiddenDevices, m.isolateMAC, filter)

	total := m.shared.store.Count()
	ble, classic, wifi := m.shared.store.CountByType()
	statusBar := ui.RenderStatusBar(m.width, m.scanning, total, ble, classic, wifi,
		m.shared.sweep.Degrees(), config.MaxRange)

	return ui.ComposeLayout(menuBar, leftPanel, deviceList, statusBar, m.width)
}

// StartScanners initializes and starts scanners. Must be called before p.Run().
func (m *AppModel) StartScanners(p *tea.Program) error {
	m.shared.resolver.Start(p)

	if m.demoMode {
		m.shared.mockScanner = bluetooth.NewMockScanner()
		return m.shared.mockScanner.Start(p)
	}

	m.shared.bleScanner = bluetooth.NewBLEScanner()
	if err := m.shared.bleScanner.Start(p); err != nil {
		return err
	}

	if bluetooth.ClassicScannerAvailable() {
		m.shared.classicScanner = bluetooth.NewClassicScanner(
			time.Duration(config.ClassicScanSec) * time.Second)
		_ = m.shared.classicScanner.Start(p)
	}

	if bluetooth.WiFiScannerAvailable() {
		m.shared.wifiScanner = bluetooth.NewWiFiScanner("",
			time.Duration(config.WiFiScanSec)*time.Second)
		_ = m.shared.wifiScanner.Start(p)
	}

	return nil
}

func (m *AppModel) stopScanners() {
	if m.shared.resolver != nil {
		m.shared.resolver.Stop()
	}
	if m.shared.mockScanner != nil {
		m.shared.mockScanner.Stop()
	}
	if m.shared.bleScanner != nil {
		m.shared.bleScanner.Stop()
	}
	if m.shared.classicScanner != nil {
		m.shared.classicScanner.Stop()
	}
	if m.shared.wifiScanner != nil {
		m.shared.wifiScanner.Stop()
	}
}

// filteredDevices returns a filtered copy of devices based on type toggles and search.
func (m AppModel) filteredDevices() []*bluetooth.Device {
	result := make([]*bluetooth.Device, 0, len(m.devices))
	search := strings.ToLower(m.filterSearch)
	for _, d := range m.devices {
		// Type filter
		switch d.Type {
		case bluetooth.DeviceTypeBLE:
			if !m.filterBLE {
				continue
			}
		case bluetooth.DeviceTypeClassic:
			if !m.filterClassic {
				continue
			}
		case bluetooth.DeviceTypeWiFi:
			if !m.filterWiFi {
				continue
			}
		}
		// Text search
		if search != "" {
			nameLower := strings.ToLower(d.Name)
			macLower := strings.ToLower(d.MAC)
			if !strings.Contains(nameLower, search) && !strings.Contains(macLower, search) {
				continue
			}
		}
		result = append(result, d)
	}
	return result
}

func (m AppModel) handleKeyFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc", "enter":
		m.filterActive = false
	case "backspace":
		if len(m.filterSearch) > 0 {
			m.filterSearch = m.filterSearch[:len(m.filterSearch)-1]
			m.refreshFilter()
		}
	default:
		// Accept printable ASCII characters
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.filterSearch += key
			m.refreshFilter()
		}
	}
	return m, nil
}

func (m *AppModel) refreshFilter() {
	savedMAC := m.selectedMAC
	m.filteredView = m.filteredDevices()
	// Try to re-anchor cursor by MAC
	if savedMAC != "" {
		for i, d := range m.filteredView {
			if d.MAC == savedMAC {
				m.cursorIndex = i
				return
			}
		}
	}
	m.clampCursor()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(config.TargetFPS), func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func evictCmd() tea.Cmd {
	return tea.Tick(config.EvictInterval, func(t time.Time) tea.Msg {
		return EvictMsg(t)
	})
}
