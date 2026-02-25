package radar

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"

	"ble-radar.klederson.com/internal/bluetooth"
	"ble-radar.klederson.com/internal/config"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorBright    = lipgloss.Color("#00FF41")
	colorMid       = lipgloss.Color("#008F11")
	colorDim       = lipgloss.Color("#004A0A")
	colorDeviceBLE = lipgloss.Color("#00FFAA")
	colorDeviceCls = lipgloss.Color("#33FF66")
	colorLabelDim  = lipgloss.Color("#008F11")

	styleCenter   = lipgloss.NewStyle().Foreground(colorBright).Bold(true)
	styleRing     = lipgloss.NewStyle().Foreground(colorMid)
	styleDot      = lipgloss.NewStyle().Foreground(colorDim)
	styleBLEDev   = lipgloss.NewStyle().Foreground(colorDeviceBLE).Bold(true)
	styleClassDev = lipgloss.NewStyle().Foreground(colorDeviceCls).Bold(true)
	styleLegBLE   = lipgloss.NewStyle().Foreground(colorDeviceBLE)
	styleLegClass = lipgloss.NewStyle().Foreground(colorDeviceCls)
	styleLabelBLE = lipgloss.NewStyle().Foreground(colorDeviceBLE)
	styleLabelCls = lipgloss.NewStyle().Foreground(colorDeviceCls)
	styleLabelDim = lipgloss.NewStyle().Foreground(colorLabelDim)
)

const maxLabelLen = 8

type devPos struct {
	col, row int
	dev      *bluetooth.Device
	label    string
	labelCol int
	labelRow int
}

// Render produces the complete radar display as a styled string.
func Render(width, height int, devices []*bluetooth.Device, sweep *Sweep) string {
	if width < 10 || height < 5 {
		return ""
	}

	centerX := width / 2
	centerY := height / 2
	radius := float64(min(centerX-1, int(float64(centerY-1)/config.AspectRatio)))
	if radius < 3 {
		radius = 3
	}

	ringRadii := make([]float64, config.RingCount)
	for i := range ringRadii {
		ringRadii[i] = radius * float64(i+1) / float64(config.RingCount)
	}

	// Pre-compute device positions and labels with collision avoidance
	dps := buildDevicePositions(devices, centerX, centerY, radius, width)

	// Build a lookup map for label cells: key = row*width+col → index into dps + char offset
	type labelCell struct {
		dpIdx   int
		charIdx int
	}
	labelMap := make(map[int]labelCell)
	for i, dp := range dps {
		if dp.label == "" {
			continue
		}
		for ci := 0; ci < len(dp.label); ci++ {
			key := dp.labelRow*width + dp.labelCol + ci
			labelMap[key] = labelCell{dpIdx: i, charIdx: ci}
		}
	}

	var sb strings.Builder
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			// Label cells (highest priority after device symbol)
			key := row*width + col
			if lc, ok := labelMap[key]; ok {
				dp := dps[lc.dpIdx]
				ch := dp.label[lc.charIdx]
				sb.WriteString(styleLabelFor(dp.dev, sweep, col, row, centerX, centerY, ch))
				continue
			}
			sb.WriteString(renderCell(col, row, centerX, centerY, radius, ringRadii, sweep, dps))
		}
		if row < height-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// buildDevicePositions computes positions and resolves label collisions.
func buildDevicePositions(devices []*bluetooth.Device, centerX, centerY int, radius float64, width int) []devPos {
	dps := make([]devPos, 0, len(devices))

	// Track occupied row segments: map[row] → list of (startCol, endCol)
	type segment struct{ start, end int }
	occupied := make(map[int][]segment)

	for _, d := range devices {
		devRadius := MetersToRadius(d.Distance, config.MaxRange, radius)
		dc := centerX + int(math.Round(devRadius*math.Sin(d.Angle)))
		dr := centerY - int(math.Round(devRadius*math.Cos(d.Angle)*config.AspectRatio))

		label := deviceCallsign(d)

		// Try placing label to the right
		lc := dc + 2
		lr := dr

		if lc+len(label) >= width {
			lc = dc - len(label) - 1
		}
		if lc < 0 {
			lc = 0
		}

		// Check collision with existing labels
		collision := false
		for _, seg := range occupied[lr] {
			// Check if our label range [lc, lc+len(label)) overlaps [seg.start, seg.end)
			if lc < seg.end && lc+len(label) > seg.start {
				collision = true
				break
			}
		}

		if collision {
			// Try one row below
			lr = dr + 1
			collision = false
			for _, seg := range occupied[lr] {
				if lc < seg.end && lc+len(label) > seg.start {
					collision = true
					break
				}
			}
		}

		if collision {
			// Try one row above
			lr = dr - 1
			collision = false
			for _, seg := range occupied[lr] {
				if lc < seg.end && lc+len(label) > seg.start {
					collision = true
					break
				}
			}
		}

		if collision {
			// Give up on label for this device to keep radar clean
			label = ""
		}

		dp := devPos{
			col:      dc,
			row:      dr,
			dev:      d,
			label:    label,
			labelCol: lc,
			labelRow: lr,
		}
		dps = append(dps, dp)

		// Mark device symbol position as occupied
		occupied[dr] = append(occupied[dr], segment{dc, dc + 1})

		// Mark label region as occupied
		if label != "" {
			occupied[lr] = append(occupied[lr], segment{lc, lc + len(label)})
		}
	}

	return dps
}

func deviceCallsign(d *bluetooth.Device) string {
	if d.Name != "" {
		name := d.Name
		if len(name) > maxLabelLen {
			name = name[:maxLabelLen]
		}
		return name
	}
	h := sha256.Sum256([]byte(d.MAC))
	return fmt.Sprintf("#%02X%X", h[0], h[1]&0x0F)
}

func styleLabelFor(d *bluetooth.Device, sweep *Sweep, col, row, centerX, centerY int, ch byte) string {
	intensity := sweep.Intensity(CellAngle(col, row, centerX, centerY))
	s := string(ch)

	if d.Name == "" {
		if intensity > 0.5 {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC33")).Render(s)
		}
		return styleLabelDim.Render(s)
	}

	if d.Type == bluetooth.DeviceTypeClassic {
		if intensity > 0.5 {
			return lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render(s)
		}
		return styleLabelCls.Render(s)
	}

	if intensity > 0.5 {
		return lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render(s)
	}
	return styleLabelBLE.Render(s)
}

func renderCell(col, row, centerX, centerY int, radius float64, ringRadii []float64, sweep *Sweep, devPositions []devPos) string {
	dist := CellDistance(col, row, centerX, centerY)
	angle := CellAngle(col, row, centerX, centerY)

	for _, dp := range devPositions {
		if col == dp.col && row == dp.row {
			return renderDevice(dp.dev, sweep, angle)
		}
	}

	if dist > radius+0.5 {
		return " "
	}

	if col == centerX && row == centerY {
		return styleCenter.Render("+")
	}

	if col == centerX && dist <= radius {
		return renderSweepChar('|', sweep, angle)
	}
	if row == centerY && dist <= radius {
		return renderSweepChar('-', sweep, angle)
	}

	for _, ringR := range ringRadii {
		if math.Abs(dist-ringR) < 0.8 {
			ch := RingChar(angle)
			return renderSweepChar(ch, sweep, angle)
		}
	}

	if dist <= radius {
		return renderInteriorCell(sweep, angle)
	}

	return " "
}

func renderDevice(d *bluetooth.Device, sweep *Sweep, cellAngle float64) string {
	intensity := sweep.Intensity(cellAngle)

	if d.Type == bluetooth.DeviceTypeClassic {
		if intensity > 0.5 {
			return lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render("B")
		}
		return styleClassDev.Render("B")
	}

	if intensity > 0.5 {
		return lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render("*")
	}
	return styleBLEDev.Render("*")
}

func renderSweepChar(ch rune, sweep *Sweep, angle float64) string {
	intensity := sweep.Intensity(angle)
	color := sweepColor(intensity)
	if color == "" {
		return styleRing.Render(string(ch))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(ch))
}

func renderInteriorCell(sweep *Sweep, angle float64) string {
	intensity := sweep.Intensity(angle)
	color := sweepColor(intensity)
	if color == "" {
		return styleDot.Render(".")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(".")
}

func sweepColor(intensity float64) string {
	if intensity <= 0 {
		return ""
	}
	if intensity > 0.8 {
		return "#00FF41"
	}
	if intensity > 0.5 {
		return "#00CC33"
	}
	if intensity > 0.3 {
		return "#00AA22"
	}
	return "#005511"
}

// RenderLegend produces the radar legend line.
func RenderLegend(width int) string {
	legend := "   " +
		styleLegBLE.Render("* BLE") +
		"  " +
		styleLegClass.Render("B Classic")

	pad := (width - lipgloss.Width(legend)) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + legend
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
