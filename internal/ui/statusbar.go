package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the bottom status bar.
func RenderStatusBar(width int, scanning bool, total, ble, classic, wifi int, sweepDeg float64, maxRange float64) string {
	status := ""
	if scanning {
		status = StyleStatusScanning.Render("[SCANNING]")
	} else {
		status = StyleStatusPaused.Render("[PAUSED]")
	}

	info := fmt.Sprintf(" Devices: %d  BLE: %d  CLS: %d  WiFi: %d  Sweep: %ddeg  Range: 0-%.0fm",
		total, ble, classic, wifi, int(sweepDeg), maxRange)

	content := status + StyleStatusBar.Foreground(ColorGreen).Render(info)

	gap := width - lipgloss.Width(content)
	if gap < 0 {
		gap = 0
	}
	padding := ""
	for i := 0; i < gap; i++ {
		padding += " "
	}

	return StyleStatusBar.Width(width).Render(content + padding)
}
