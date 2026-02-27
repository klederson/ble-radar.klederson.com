package ui

import "github.com/charmbracelet/lipgloss"

// Matrix color palette
var (
	ColorMatrixGreen = lipgloss.Color("#00FF41")
	ColorGreen       = lipgloss.Color("#00CC33")
	ColorMidGreen    = lipgloss.Color("#008F11")
	ColorDimGreen    = lipgloss.Color("#004A0A")
	ColorBlack       = lipgloss.Color("#000000")
	ColorDeviceBLE   = lipgloss.Color("#00FFAA")
	ColorDeviceClass = lipgloss.Color("#33FF66")
	ColorDeviceWiFi  = lipgloss.Color("#FFCC00")
	ColorBorderBright= lipgloss.Color("#00FF41")
	ColorBorderNorm  = lipgloss.Color("#00AA22")
	ColorError       = lipgloss.Color("#FF3300")
	ColorWarning     = lipgloss.Color("#FFAA00")
	ColorSweepBright = lipgloss.Color("#00FF41")
	ColorSweepMid    = lipgloss.Color("#00AA22")
	ColorSweepDim    = lipgloss.Color("#005511")
)

// Pre-built styles
var (
	StyleMenuBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#002200")).
			Foreground(ColorMatrixGreen).
			Bold(true).
			Padding(0, 1)

	StyleMenuKey = lipgloss.NewStyle().
			Foreground(ColorMatrixGreen).
			Bold(true)

	StyleMenuLabel = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#002200")).
			Foreground(ColorGreen).
			Padding(0, 1)

	StyleStatusScanning = lipgloss.NewStyle().
				Foreground(ColorMatrixGreen).
				Bold(true)

	StyleStatusPaused = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	StylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderNorm)

	StylePanelActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderBright)

	StylePanelTitle = lipgloss.NewStyle().
			Foreground(ColorMatrixGreen).
			Bold(true).
			Padding(0, 1)

	StyleDeviceName = lipgloss.NewStyle().
			Foreground(ColorMatrixGreen).
			Bold(true)

	StyleDeviceMAC = lipgloss.NewStyle().
			Foreground(ColorMidGreen)

	StyleDeviceRSSI = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleDeviceDist = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleDeviceTypeBLE = lipgloss.NewStyle().
				Foreground(ColorDeviceBLE)

	StyleDeviceTypeClassic = lipgloss.NewStyle().
				Foreground(ColorDeviceClass)

	StyleDeviceTypeWiFi = lipgloss.NewStyle().
				Foreground(ColorDeviceWiFi)

	StyleFilterActive = lipgloss.NewStyle().
				Foreground(ColorMatrixGreen).
				Bold(true)

	StyleFilterInactive = lipgloss.NewStyle().
				Foreground(ColorDimGreen)

	StyleRadarCenter = lipgloss.NewStyle().
			Foreground(ColorMatrixGreen).
			Bold(true)

	StyleRadarRing = lipgloss.NewStyle().
			Foreground(ColorMidGreen)

	StyleRadarDot = lipgloss.NewStyle().
			Foreground(ColorDimGreen)

	StyleRadarBLEDevice = lipgloss.NewStyle().
				Foreground(ColorDeviceBLE).
				Bold(true)

	StyleRadarClassicDevice = lipgloss.NewStyle().
				Foreground(ColorDeviceClass).
				Bold(true)

	StyleLegend = lipgloss.NewStyle().
			Foreground(ColorMidGreen)

	StyleLegendBLE = lipgloss.NewStyle().
			Foreground(ColorDeviceBLE)

	StyleLegendClassic = lipgloss.NewStyle().
			Foreground(ColorDeviceClass)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorDimGreen)

	StyleCursorLine = lipgloss.NewStyle().
			Background(lipgloss.Color("#003300"))

	StyleCheckOn = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleCheckOff = lipgloss.NewStyle().
			Foreground(ColorDimGreen)

	StyleIsolateMarker = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)
)
