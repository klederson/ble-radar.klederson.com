package ui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderCompass renders a compass with an arrow pointing toward a device.
// angle: radians (0=north, clockwise), distance: meters, rssi: dBm.
func RenderCompass(width, height int, angle, distance, rssi float64) string {
	if width < 9 || height < 5 {
		return ""
	}

	grid := make([][]byte, height)
	isArrow := make([][]bool, height)
	for i := range grid {
		grid[i] = make([]byte, width)
		isArrow[i] = make([]bool, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	fcx := float64(width) / 2.0
	fcy := float64(height) / 2.0
	rx := fcx - 2.0 // horizontal radius in columns
	ry := fcy - 2.0 // vertical radius in rows
	if rx < 3 {
		rx = 3
	}
	if ry < 2 {
		ry = 2
	}

	// Draw compass ring
	steps := 80
	for i := 0; i < steps; i++ {
		a := float64(i) * 2 * math.Pi / float64(steps)
		col := int(math.Round(fcx + rx*math.Sin(a)))
		row := int(math.Round(fcy - ry*math.Cos(a)))
		if col >= 0 && col < width && row >= 0 && row < height && grid[row][col] == ' ' {
			grid[row][col] = ringChar(a)
		}
	}

	cx := int(math.Round(fcx))
	cy := int(math.Round(fcy))

	// Cardinal markers
	nRow := cy - int(math.Round(ry)) - 1
	sRow := cy + int(math.Round(ry)) + 1
	eCol := cx + int(math.Round(rx)) + 1
	wCol := cx - int(math.Round(rx)) - 1
	setGrid(grid, width, height, cx, nRow, 'N')
	setGrid(grid, width, height, cx, sRow, 'S')
	setGrid(grid, width, height, eCol, cy, 'E')
	setGrid(grid, width, height, wCol, cy, 'W')

	// Cross hairs (faint axes)
	for r := cy - int(ry) + 1; r < cy+int(ry); r++ {
		if r != cy && grid[r][cx] == ' ' {
			grid[r][cx] = ':'
		}
	}
	for c := cx - int(rx) + 1; c < cx+int(rx); c++ {
		if c != cx && grid[cy][c] == ' ' {
			grid[cy][c] = '.'
		}
	}

	// Center
	setGrid(grid, width, height, cx, cy, '+')

	// Arrow from center toward device angle
	// Length proportional to proximity (closer = longer arrow)
	maxFrac := 0.85
	minFrac := 0.3
	distFrac := math.Min(distance/20.0, 1.0)
	arrowFrac := maxFrac - (maxFrac-minFrac)*distFrac // closer = longer

	sinA := math.Sin(angle)
	cosA := math.Cos(angle)

	// Draw shaft: step from center to tip
	shaftSteps := int(math.Max(rx, ry) * arrowFrac)
	if shaftSteps < 2 {
		shaftSteps = 2
	}

	var tipCol, tipRow int
	for s := 1; s <= shaftSteps; s++ {
		t := float64(s) / float64(shaftSteps) * arrowFrac
		fx := fcx + t*rx*sinA
		fy := fcy - t*ry*cosA
		col := int(math.Round(fx))
		row := int(math.Round(fy))
		if col >= 0 && col < width && row >= 0 && row < height {
			ch := shaftChar(angle)
			grid[row][col] = ch
			isArrow[row][col] = true
			tipCol = col
			tipRow = row
		}
	}

	// Arrowhead at tip
	tipCh := arrowTip(angle)
	if tipCol >= 0 && tipCol < width && tipRow >= 0 && tipRow < height {
		grid[tipRow][tipCol] = tipCh
		isArrow[tipRow][tipCol] = true
	}

	// Small wing lines at tip
	wingLen := 2
	wingAngleL := angle - math.Pi*0.8
	wingAngleR := angle + math.Pi*0.8
	for w := 1; w <= wingLen; w++ {
		t := float64(w) * 0.4
		// Left wing
		wlc := int(math.Round(float64(tipCol) + t*rx/float64(shaftSteps)*math.Sin(wingAngleL)*2))
		wlr := int(math.Round(float64(tipRow) - t*ry/float64(shaftSteps)*math.Cos(wingAngleL)*2))
		if wlc >= 0 && wlc < width && wlr >= 0 && wlr < height {
			grid[wlr][wlc] = shaftChar(wingAngleL)
			isArrow[wlr][wlc] = true
		}
		// Right wing
		wrc := int(math.Round(float64(tipCol) + t*rx/float64(shaftSteps)*math.Sin(wingAngleR)*2))
		wrr := int(math.Round(float64(tipRow) - t*ry/float64(shaftSteps)*math.Cos(wingAngleR)*2))
		if wrc >= 0 && wrc < width && wrr >= 0 && wrr < height {
			grid[wrr][wrc] = shaftChar(wingAngleR)
			isArrow[wrr][wrc] = true
		}
	}

	// Render with colors
	arrowColor := proximityColor(rssi)
	arrowSty := lipgloss.NewStyle().Foreground(lipgloss.Color(arrowColor)).Bold(true)
	ringSty := lipgloss.NewStyle().Foreground(ColorDimGreen)
	axisSty := lipgloss.NewStyle().Foreground(lipgloss.Color("#003300"))
	markSty := lipgloss.NewStyle().Foreground(ColorMatrixGreen).Bold(true)

	var sb strings.Builder
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			ch := grid[row][col]
			switch {
			case ch == 'N' || ch == 'S' || ch == 'E' || ch == 'W':
				sb.WriteString(markSty.Render(string(ch)))
			case ch == '+':
				sb.WriteString(markSty.Render(string(ch)))
			case isArrow[row][col]:
				sb.WriteString(arrowSty.Render(string(ch)))
			case ch == ':' || ch == '.':
				sb.WriteString(axisSty.Render(string(ch)))
			case ch != ' ':
				sb.WriteString(ringSty.Render(string(ch)))
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

func setGrid(grid [][]byte, w, h, col, row int, ch byte) {
	if col >= 0 && col < w && row >= 0 && row < h {
		grid[row][col] = ch
	}
}

func ringChar(a float64) byte {
	for a < 0 {
		a += 2 * math.Pi
	}
	for a >= 2*math.Pi {
		a -= 2 * math.Pi
	}
	sector := int(math.Round(a/(math.Pi/4))) % 8
	switch sector {
	case 0:
		return '-'
	case 1:
		return '\\'
	case 2:
		return '|'
	case 3:
		return '/'
	case 4:
		return '-'
	case 5:
		return '\\'
	case 6:
		return '|'
	case 7:
		return '/'
	}
	return '-'
}

// shaftChar returns the line character for a given angle direction.
func shaftChar(a float64) byte {
	for a < 0 {
		a += 2 * math.Pi
	}
	for a >= 2*math.Pi {
		a -= 2 * math.Pi
	}
	// 8 direction sectors
	sector := int(math.Round(a/(math.Pi/4))) % 8
	switch sector {
	case 0, 4: // N, S
		return '|'
	case 2, 6: // E, W
		return '-'
	case 1, 5: // NE, SW
		return '\\'
	case 3, 7: // SE, NW
		return '/'
	}
	return '|'
}

// arrowTip returns the arrowhead character for a given angle.
func arrowTip(a float64) byte {
	for a < 0 {
		a += 2 * math.Pi
	}
	for a >= 2*math.Pi {
		a -= 2 * math.Pi
	}
	sector := int(math.Round(a/(math.Pi/4))) % 8
	switch sector {
	case 0: // N
		return '^'
	case 1: // NE
		return '/'
	case 2: // E
		return '>'
	case 3: // SE
		return '\\'
	case 4: // S
		return 'v'
	case 5: // SW
		return '/'
	case 6: // W
		return '<'
	case 7: // NW
		return '\\'
	}
	return '*'
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
