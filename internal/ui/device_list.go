package ui

import (
	"fmt"
	"strings"

	"ble-radar.klederson.com/internal/bluetooth"
	"github.com/charmbracelet/lipgloss"
)

// Cursor row style: black text on bright green = unmissable highlight
var cursorRowSty = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000000")).
	Background(ColorMatrixGreen).
	Bold(true)

// Hidden device style: very dim
var hiddenDevSty = lipgloss.NewStyle().
	Foreground(ColorDimGreen)

// RenderDeviceList renders the scrollable device list panel with cursor and visibility controls.
func RenderDeviceList(devices []*bluetooth.Device, width, height int, cursorIndex int, hiddenDevices map[string]bool, isolateMAC string) string {
	innerW := width - 4
	if innerW < 10 {
		innerW = 10
	}

	title := StylePanelTitle.Render(fmt.Sprintf("DEVICES [%d]", len(devices)))
	separator := StyleRadarRing.Render(strings.Repeat("-", innerW))

	lines := []string{title, separator}

	if len(devices) == 0 {
		lines = append(lines, "")
		lines = append(lines, StyleHelp.Render(" No devices..."))
		lines = append(lines, StyleHelp.Render(" Waiting for scan"))
	} else {
		linesPerDevice := 4 // 3 content + 1 blank
		maxVisible := (height - 4) / linesPerDevice
		if maxVisible < 1 {
			maxVisible = 1
		}

		// Compute viewport start so cursor is always visible
		viewStart := 0
		if cursorIndex >= maxVisible {
			viewStart = cursorIndex - maxVisible + 1
		}

		for i := viewStart; i < len(devices); i++ {
			isCursor := i == cursorIndex
			isHidden := hiddenDevices[devices[i].MAC]
			isIsolated := devices[i].MAC == isolateMAC

			entry := renderDeviceEntryFull(devices[i], innerW, isCursor, isHidden, isIsolated)
			for _, l := range entry {
				if len(lines) >= height-3 {
					break
				}
				lines = append(lines, l)
			}
			if len(lines) >= height-3 {
				break
			}
		}
	}

	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return StylePanelBorder.Width(width - 2).Height(height - 2).Render(content)
}

func renderDeviceEntryFull(d *bluetooth.Device, maxW int, isCursor, isHidden, isIsolated bool) []string {
	symbol := "*"
	tag := "[BLE]"
	if d.Type == bluetooth.DeviceTypeClassic {
		symbol = "B"
		tag = "[CLS]"
	}

	name := d.DisplayName()
	nameMax := maxW - 18
	if nameMax < 4 {
		nameMax = 4
	}
	if len(name) > nameMax {
		name = name[:nameMax]
	}

	check := "[x]"
	if isHidden {
		check = "[ ]"
	}

	iso := " "
	if isIsolated {
		iso = "!"
	}

	cursor := "  "
	if isCursor {
		cursor = ">>"
	}

	mac := d.MAC
	if len(mac) > maxW-8 {
		mac = mac[:maxW-8]
	}

	rssiStr := fmt.Sprintf("%ddBm", int(d.RSSI))
	distStr := fmt.Sprintf("~%.1fm", d.Distance)

	rawLine1 := fmt.Sprintf("%s %s %s %s %s %s", cursor, check, symbol, name, iso, tag)
	rawLine2 := fmt.Sprintf("       %s", mac)
	rawLine3 := fmt.Sprintf("       %s  %s", rssiStr, distStr)

	// Pad all raw lines to full width
	rawLine1 = padRaw(rawLine1, maxW)
	rawLine2 = padRaw(rawLine2, maxW)
	rawLine3 = padRaw(rawLine3, maxW)

	if isCursor {
		// Black text on bright green background - unmissable
		sty := cursorRowSty
		return []string{
			sty.Render(rawLine1),
			sty.Render(rawLine2),
			sty.Render(rawLine3),
			"",
		}
	}

	if isHidden {
		// Dim everything for hidden devices
		return []string{
			hiddenDevSty.Render(rawLine1),
			hiddenDevSty.Render(rawLine2),
			hiddenDevSty.Render(rawLine3),
			"",
		}
	}

	// Normal styled rendering
	return renderNormalEntry(d, rawLine1, rawLine2, rawLine3, maxW, isIsolated)
}

func renderNormalEntry(d *bluetooth.Device, raw1, raw2, raw3 string, maxW int, isIsolated bool) []string {
	symbol := StyleDeviceTypeBLE.Render("*")
	typeTag := StyleDeviceTypeBLE.Render("[BLE]")
	if d.Type == bluetooth.DeviceTypeClassic {
		symbol = StyleDeviceTypeClassic.Render("B")
		typeTag = StyleDeviceTypeClassic.Render("[CLS]")
	}

	name := d.DisplayName()
	nameMax := maxW - 18
	if nameMax < 4 {
		nameMax = 4
	}
	if len(name) > nameMax {
		name = name[:nameMax]
	}

	check := StyleCheckOn.Render("[x]")
	iso := " "
	if isIsolated {
		iso = StyleIsolateMarker.Render("!")
	}

	mac := d.MAC
	if len(mac) > maxW-8 {
		mac = mac[:maxW-8]
	}

	rssiStr := fmt.Sprintf("%ddBm", int(d.RSSI))
	distStr := fmt.Sprintf("~%.1fm", d.Distance)

	line1 := fmt.Sprintf("   %s %s %s %s %s", check, symbol, StyleDeviceName.Render(name), iso, typeTag)
	line2 := fmt.Sprintf("       %s", StyleDeviceMAC.Render(mac))
	line3 := fmt.Sprintf("       %s  %s", StyleDeviceRSSI.Render(rssiStr), StyleDeviceDist.Render(distStr))

	return []string{line1, line2, line3, ""}
}

func padRaw(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
