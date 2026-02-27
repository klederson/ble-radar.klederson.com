package ui

import "strings"

// RenderRadarPanel wraps radar content with a styled border.
// The actual radar rendering is done externally to avoid import cycles.
func RenderRadarPanel(width, height int, radarContent, legend string) string {
	content := radarContent + "\n" + legend
	rendered := StylePanelBorder.Width(width - 2).Height(height - 2).Render(content)

	// Clamp to exactly height lines so both panels match.
	outLines := strings.Split(rendered, "\n")
	if len(outLines) > height {
		outLines = outLines[:height]
	}
	for len(outLines) < height {
		outLines = append(outLines, "")
	}
	return strings.Join(outLines, "\n")
}
