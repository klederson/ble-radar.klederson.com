package app

import (
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
	mockScanner    *bluetooth.MockScanner
}

// AppModel is the root Bubble Tea model for BLE Radar.
type AppModel struct {
	width  int
	height int

	scanning     bool
	demoMode     bool
	adapter      string
	scrollOffset int

	shared *shared

	// Cached snapshot
	devices []*bluetooth.Device
}

// New creates a new AppModel.
func New(demoMode bool, adapter string) AppModel {
	return AppModel{
		scanning: true,
		demoMode: demoMode,
		adapter:  adapter,
		shared: &shared{
			store: bluetooth.NewDeviceStore(),
			sweep: radar.NewSweep(),
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
		return m, tickCmd()

	case EvictMsg:
		m.shared.store.Evict(config.DeviceTimeout)
		return m, evictCmd()

	case bluetooth.DeviceDiscoveredMsg:
		if m.scanning {
			m.shared.store.Upsert(msg.MAC, msg.Name, float64(msg.RSSI), msg.Type)
		}
		return m, nil

	case ScanErrorMsg:
		return m, nil
	}

	return m, nil
}

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}

	case "down", "j":
		if m.scrollOffset < len(m.devices)-1 {
			m.scrollOffset++
		}

	case "home":
		m.scrollOffset = 0

	case "end":
		if len(m.devices) > 0 {
			m.scrollOffset = len(m.devices) - 1
		}
	}

	return m, nil
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

	menuBar := ui.RenderMenuBar(m.width, m.adapter, m.scanning)

	innerW := radarW - 4
	innerH := bodyH - 4
	if innerW < 5 {
		innerW = 5
	}
	if innerH < 3 {
		innerH = 3
	}
	radarContent := radar.Render(innerW, innerH, m.devices, m.shared.sweep)
	legend := radar.RenderLegend(innerW)
	radarPanel := ui.RenderRadarPanel(radarW, bodyH, radarContent, legend)

	deviceList := ui.RenderDeviceList(m.devices, listW, bodyH, m.scrollOffset)

	total := m.shared.store.Count()
	ble, classic := m.shared.store.CountByType()
	statusBar := ui.RenderStatusBar(m.width, m.scanning, total, ble, classic,
		m.shared.sweep.Degrees(), config.MaxRange)

	return ui.ComposeLayout(menuBar, radarPanel, deviceList, statusBar, m.width)
}

// StartScanners initializes and starts scanners. Must be called before p.Run().
func (m *AppModel) StartScanners(p *tea.Program) error {
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

	return nil
}

func (m *AppModel) stopScanners() {
	if m.shared.mockScanner != nil {
		m.shared.mockScanner.Stop()
	}
	if m.shared.bleScanner != nil {
		m.shared.bleScanner.Stop()
	}
	if m.shared.classicScanner != nil {
		m.shared.classicScanner.Stop()
	}
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
