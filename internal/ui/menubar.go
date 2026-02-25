package ui

import (
	"fmt"

	"ble-radar.klederson.com/internal/config"
	"github.com/charmbracelet/lipgloss"
)

// RenderMenuBar renders the top menu bar.
func RenderMenuBar(width int, adapter string, scanning bool) string {
	title := fmt.Sprintf(" %s v%s ", config.AppName, config.AppVersion)

	keys := []struct{ key, label string }{
		{"S", "can"},
		{"P", "ause"},
		{"F", "ilter"},
		{"H", "elp"},
		{"Q", "uit"},
	}

	menu := ""
	for _, k := range keys {
		menu += "  " + StyleMenuKey.Render("["+k.key+"]") + StyleMenuLabel.Render(k.label)
	}

	status := ""
	if scanning {
		status = StyleStatusScanning.Render("SCANNING")
	} else {
		status = StyleStatusPaused.Render("PAUSED")
	}

	adapterInfo := StyleMenuLabel.Render(fmt.Sprintf("Adapter: %s", adapter))

	left := StyleMenuKey.Render(title) + menu
	right := status + "  " + adapterInfo + " "

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	padding := ""
	for i := 0; i < gap; i++ {
		padding += " "
	}

	return StyleMenuBar.Width(width).Render(left + padding + right)
}
