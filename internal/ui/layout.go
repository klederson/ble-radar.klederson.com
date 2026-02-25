package ui

import "github.com/charmbracelet/lipgloss"

// ComposeLayout joins the radar panel and device list horizontally,
// with menu bar on top and status bar on bottom.
func ComposeLayout(menuBar, radarPanel, deviceList, statusBar string, width int) string {
	middle := lipgloss.JoinHorizontal(lipgloss.Top, radarPanel, deviceList)
	return lipgloss.JoinVertical(lipgloss.Left, menuBar, middle, statusBar)
}
