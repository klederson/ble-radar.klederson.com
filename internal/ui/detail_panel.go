package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"ble-radar.klederson.com/internal/bluetooth"
	"github.com/charmbracelet/lipgloss"
)

// RenderDetailPanel renders the device detail overlay that replaces the radar area.
func RenderDetailPanel(d *bluetooth.Device, width, height int, rssiHistory []float64) string {
	innerW := width - 4
	if innerW < 20 {
		innerW = 20
	}

	title := StylePanelTitle.Render("DEVICE DETAIL")
	escHint := StyleHelp.Render("[ESC]")
	titleLine := title + strings.Repeat(" ", max(0, innerW-lipgloss.Width(title)-lipgloss.Width(escHint))) + escHint

	sep := StyleRadarRing.Render(strings.Repeat("-", innerW))

	lines := []string{titleLine, sep, ""}

	// Device info fields
	labelSty := lipgloss.NewStyle().Foreground(ColorMidGreen)
	valSty := lipgloss.NewStyle().Foreground(ColorMatrixGreen).Bold(true)

	fields := []struct{ label, value string }{
		{"Name", d.DisplayName()},
		{"MAC", d.MAC},
		{"Type", d.Type.String()},
		{"RSSI", fmt.Sprintf("%d dBm", int(d.RSSI))},
		{"Distance", fmt.Sprintf("~%.1fm", d.Distance)},
		{"Last", formatLastSeen(d.LastSeen)},
	}

	for _, f := range fields {
		label := labelSty.Render(fmt.Sprintf("  %-10s", f.label))
		value := valSty.Render(f.value)
		lines = append(lines, label+value)
	}

	lines = append(lines, "")

	// Signal bar
	barWidth := innerW - 22
	if barWidth < 10 {
		barWidth = 10
	}
	bar := renderSignalBar(d.RSSI, barWidth)
	rssiLabel := valSty.Render(fmt.Sprintf(" %ddBm", int(d.RSSI)))
	lines = append(lines, labelSty.Render("  Signal ")+bar+rssiLabel)

	lines = append(lines, "")

	// RSSI sparkline
	if len(rssiHistory) > 0 {
		sparkW := innerW - 4
		if sparkW < 10 {
			sparkW = 10
		}
		lines = append(lines, labelSty.Render("  RSSI History:"))
		spark := renderSparkline(rssiHistory, sparkW)
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(ColorGreen).Render(spark))
	}

	lines = append(lines, "")

	// Compass
	usedLines := len(lines)
	compassH := height - usedLines - 5 // leave room for label + border
	if compassH < 5 {
		compassH = 5
	}
	compassW := innerW
	if compassW > compassH*3 {
		compassW = compassH * 3 // keep roughly proportional
	}

	compass := RenderPerspective(compassW, compassH, d.Angle, d.Elevation, d.Distance, d.RSSI)
	if compass != "" {
		// Center compass horizontally
		compassLines := strings.Split(compass, "\n")
		pad := (innerW - compassW) / 2
		if pad < 0 {
			pad = 0
		}
		prefix := strings.Repeat(" ", pad)
		for _, cl := range compassLines {
			lines = append(lines, prefix+cl)
		}
	}

	// Direction + distance label centered below compass
	dir := angleToDir(d.Angle)
	vertLabel := elevationLabel(d.Elevation)
	distLabel := fmt.Sprintf("~%.1fm  %s  %s  %ddBm", d.Distance, dir, vertLabel, int(d.RSSI))
	distPad := (innerW - len(distLabel)) / 2
	if distPad < 0 {
		distPad = 0
	}
	lines = append(lines, strings.Repeat(" ", distPad)+valSty.Render(distLabel))

	// Pad to fill height
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return StylePanelActive.Width(width - 2).Height(height - 2).Render(content)
}

func renderSignalBar(rssi float64, width int) string {
	// Map RSSI -100..-30 to 0..width filled bars
	ratio := (rssi + 100.0) / 70.0
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(math.Round(ratio * float64(width)))

	bar := strings.Repeat("|", filled) + strings.Repeat("-", width-filled)
	filledPart := lipgloss.NewStyle().Foreground(lipgloss.Color(proximityColor(rssi))).Render(bar[:filled])
	emptyPart := lipgloss.NewStyle().Foreground(ColorDimGreen).Render(bar[filled:])
	return StyleHelp.Render("[") + filledPart + emptyPart + StyleHelp.Render("]")
}

func renderSparkline(values []float64, width int) string {
	if len(values) == 0 {
		return ""
	}

	chars := []byte{'_', '.', '-', '~', '^'}

	// Find min/max for scaling
	minV, maxV := values[0], values[0]
	for _, v := range values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	rng := maxV - minV
	if rng < 1 {
		rng = 1
	}

	// Take last `width` values
	start := 0
	if len(values) > width {
		start = len(values) - width
	}

	var sb strings.Builder
	for i := start; i < len(values); i++ {
		idx := int((values[i] - minV) / rng * float64(len(chars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		sb.WriteByte(chars[idx])
	}

	return sb.String()
}

func angleToDir(a float64) string {
	for a < 0 {
		a += 2 * math.Pi
	}
	for a >= 2*math.Pi {
		a -= 2 * math.Pi
	}
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int(math.Round(a/(math.Pi/4))) % 8
	return dirs[idx]
}

func elevationLabel(elev float64) string {
	if elev > 0.2 {
		return "above"
	}
	if elev < -0.2 {
		return "below"
	}
	return "level"
}

func formatLastSeen(t time.Time) string {
	d := time.Since(t)
	if d < time.Second {
		return "now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm ago", int(d.Minutes()))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
