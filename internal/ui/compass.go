package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderPerspective renders a first-person perspective view showing a device's
// position in 3D space relative to the user. Grid lines converge toward a
// vanishing point at the center, and the device is rendered as a marker at its
// projected screen position.
//
// angle: radians (0=north, clockwise) - mapped to horizontal position
// elevation: [-1, +1] where -1=below, 0=level, +1=above
// distance: meters
// rssi: dBm
func RenderPerspective(width, height int, angle, elevation, distance, rssi float64) string {
	if width < 9 || height < 5 {
		return ""
	}

	grid := make([][]byte, height)
	cellType := make([][]int, height) // 0=empty, 1=grid, 2=label, 3=marker
	for i := range grid {
		grid[i] = make([]byte, width)
		cellType[i] = make([]int, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	vpX := width / 2  // vanishing point X
	vpY := height / 2 // vanishing point Y

	// Draw perspective floor lines from bottom corners to vanishing point
	drawLine(grid, cellType, width, height, 0, height-1, vpX, vpY, '/', 1)
	drawLine(grid, cellType, width, height, width-1, height-1, vpX, vpY, '\\', 1)

	// Draw additional perspective lines for depth illusion
	drawLine(grid, cellType, width, height, width/4, height-1, vpX, vpY, '/', 1)
	drawLine(grid, cellType, width, height, width*3/4, height-1, vpX, vpY, '\\', 1)

	// Draw ceiling lines from top corners to vanishing point
	drawLine(grid, cellType, width, height, 0, 0, vpX, vpY, '\\', 1)
	drawLine(grid, cellType, width, height, width-1, 0, vpX, vpY, '/', 1)

	// Horizon line (dashed) at vertical center
	for c := 0; c < width; c++ {
		if grid[vpY][c] == ' ' {
			if c%2 == 0 {
				grid[vpY][c] = '-'
			} else {
				grid[vpY][c] = ' '
			}
			cellType[vpY][c] = 1
		}
	}

	// Vertical center line (dotted)
	for r := 0; r < height; r++ {
		if grid[r][vpX] == ' ' {
			grid[r][vpX] = ':'
			cellType[r][vpX] = 1
		}
	}

	// Vanishing point marker
	grid[vpY][vpX] = '+'
	cellType[vpY][vpX] = 1

	// Map device angle to X in [-1, +1]
	// angle is in radians (0=north, clockwise)
	// Map so that: 0/2pi (north) = 0 (center), pi/2 (east) = +1 (right),
	// pi (south) = 0 (behind, wrap), 3pi/2 (west) = -1 (left)
	xNorm := math.Sin(angle) // -1 to +1, positive = right
	yNorm := elevation       // -1 to +1, positive = above

	// Perspective projection
	// Normalize distance to [0,1] range: 0=close, 1=far
	zNorm := math.Min(distance/20.0, 1.0)
	focalLength := 1.0
	perspScale := focalLength / (focalLength + zNorm*3.0)

	halfW := float64(width)/2.0 - 2.0
	halfH := float64(height)/2.0 - 2.0

	screenX := float64(vpX) + xNorm*perspScale*halfW
	screenY := float64(vpY) - yNorm*perspScale*halfH

	// Clamp to grid bounds
	devCol := int(math.Round(screenX))
	devRow := int(math.Round(screenY))
	if devCol < 1 {
		devCol = 1
	}
	if devCol >= width-1 {
		devCol = width - 2
	}
	if devRow < 1 {
		devRow = 1
	}
	if devRow >= height-1 {
		devRow = height - 2
	}

	// Device marker character based on distance
	marker := deviceMarker(distance)
	grid[devRow][devCol] = marker
	cellType[devRow][devCol] = 3

	// Position hint labels
	if xNorm < -0.2 {
		placeLabel(grid, cellType, width, height, 1, vpY, "<")
	} else if xNorm > 0.2 {
		placeLabel(grid, cellType, width, height, width-2, vpY, ">")
	}

	if yNorm > 0.2 {
		lbl := "ABOVE"
		col := vpX - len(lbl)/2
		placeLabel(grid, cellType, width, height, col, 0, lbl)
	}
	if yNorm < -0.2 {
		lbl := "BELOW"
		col := vpX - len(lbl)/2
		placeLabel(grid, cellType, width, height, col, height-1, lbl)
	}

	// Distance label near marker
	distStr := fmt.Sprintf("~%.1fm", distance)
	lblRow := devRow + 1
	if lblRow >= height {
		lblRow = devRow - 1
	}
	lblCol := devCol - len(distStr)/2
	if lblCol < 0 {
		lblCol = 0
	}
	if lblCol+len(distStr) > width {
		lblCol = width - len(distStr)
	}
	placeLabel(grid, cellType, width, height, lblCol, lblRow, distStr)

	// Render with colors
	markerColor := proximityColor(rssi)
	markerSty := lipgloss.NewStyle().Foreground(lipgloss.Color(markerColor)).Bold(true)
	gridSty := lipgloss.NewStyle().Foreground(ColorDimGreen)
	labelSty := lipgloss.NewStyle().Foreground(ColorMidGreen)
	axisSty := lipgloss.NewStyle().Foreground(lipgloss.Color("#003300"))

	var sb strings.Builder
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			ch := grid[row][col]
			ct := cellType[row][col]
			switch {
			case ct == 3:
				sb.WriteString(markerSty.Render(string(ch)))
			case ct == 2:
				sb.WriteString(labelSty.Render(string(ch)))
			case ch == '+':
				sb.WriteString(labelSty.Render(string(ch)))
			case ch == ':' || (ch == '-' && row == vpY):
				sb.WriteString(axisSty.Render(string(ch)))
			case ct == 1:
				sb.WriteString(gridSty.Render(string(ch)))
			case ch != ' ':
				sb.WriteString(gridSty.Render(string(ch)))
			default:
				sb.WriteByte(' ')
			}
		}
		if row < height-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// drawLine uses parametric stepping to draw a line between two points.
func drawLine(grid [][]byte, cellType [][]int, w, h, x0, y0, x1, y1 int, ch byte, ct int) {
	dx := x1 - x0
	dy := y1 - y0
	steps := intAbs(dx)
	if intAbs(dy) > steps {
		steps = intAbs(dy)
	}
	if steps == 0 {
		return
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		col := int(math.Round(float64(x0) + t*float64(dx)))
		row := int(math.Round(float64(y0) + t*float64(dy)))
		if col >= 0 && col < w && row >= 0 && row < h {
			if grid[row][col] == ' ' {
				// Pick line character based on slope
				grid[row][col] = lineChar(dx, dy)
				cellType[row][col] = ct
			}
		}
	}
}

// lineChar picks an appropriate ASCII character based on line direction.
func lineChar(dx, dy int) byte {
	if dx == 0 {
		return ':'
	}
	if dy == 0 {
		return '-'
	}
	slope := float64(dy) / float64(dx)
	absSlope := math.Abs(slope)
	if absSlope > 2.0 {
		return ':'
	}
	if absSlope < 0.5 {
		return '-'
	}
	if (dx > 0 && dy > 0) || (dx < 0 && dy < 0) {
		return '\\'
	}
	return '/'
}

// deviceMarker returns the marker character based on distance.
func deviceMarker(distance float64) byte {
	if distance < 2.0 {
		return '@'
	}
	if distance < 5.0 {
		return 'O'
	}
	if distance < 10.0 {
		return 'o'
	}
	return '.'
}

// placeLabel writes a string into the grid at the given position.
func placeLabel(grid [][]byte, cellType [][]int, w, h, col, row int, label string) {
	if row < 0 || row >= h {
		return
	}
	for i, ch := range label {
		c := col + i
		if c >= 0 && c < w {
			grid[row][c] = byte(ch)
			cellType[row][c] = 2
		}
	}
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// proximityColor maps RSSI to a green shade (brighter = closer).
func proximityColor(rssi float64) string {
	if rssi > -50 {
		return "#00FF41"
	}
	if rssi > -60 {
		return "#00CC33"
	}
	if rssi > -70 {
		return "#00AA22"
	}
	if rssi > -80 {
		return "#008F11"
	}
	return "#005511"
}
