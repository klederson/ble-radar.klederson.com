package ui

// RenderRadarPanel wraps radar content with a styled border.
// The actual radar rendering is done externally to avoid import cycles.
func RenderRadarPanel(width, height int, radarContent, legend string) string {
	content := radarContent + "\n" + legend
	return StylePanelBorder.Width(width - 2).Height(height - 2).Render(content)
}
