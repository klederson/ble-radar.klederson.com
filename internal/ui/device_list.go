package ui

import (
	"fmt"
	"strings"

	"ble-radar.klederson.com/internal/bluetooth"
	"github.com/charmbracelet/lipgloss"
)

// FilterState holds the current filter settings for the device list.
type FilterState struct {
	BLE     bool   // show BLE devices
	Classic bool   // show Classic devices
	WiFi    bool   // show WiFi devices
	Search  string // text search on name/MAC
	Active  bool   // text input mode
}

// Cursor row style: black text on bright green = unmissable highlight
var cursorRowSty = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000000")).
	Background(ColorMatrixGreen).
	Bold(true)

// Hidden device style: very dim
var hiddenDevSty = lipgloss.NewStyle().
	Foreground(ColorDimGreen)

// RenderDeviceList renders the scrollable device list panel with cursor and visibility controls.
// The filter bar stays fixed at the top; only the device entries scroll.
func RenderDeviceList(devices []*bluetooth.Device, width, height int, cursorIndex int, hiddenDevices map[string]bool, isolateMAC string, filter FilterState) string {
	innerW := width - 4
	if innerW < 10 {
		innerW = 10
	}

	// Fixed header: title + separator + filter bar (3 lines)
	title := StylePanelTitle.Render(fmt.Sprintf("DEVICES [%d]", len(devices)))
	separator := StyleRadarRing.Render(strings.Repeat("-", innerW))
	filterBar := renderFilterBar(filter, innerW)
	headerLines := []string{title, separator, filterBar}
	headerCount := len(headerLines)

	// Total inner height (excluding border top+bottom)
	innerH := height - 2
	if innerH < headerCount+1 {
		innerH = headerCount + 1
	}

	// Space available for device entries
	devSpace := innerH - headerCount
	if devSpace < 1 {
		devSpace = 1
	}

	// Build device entry lines
	var devLines []string
	if len(devices) == 0 {
		devLines = append(devLines, "")
		devLines = append(devLines, StyleHelp.Render(" No devices..."))
		devLines = append(devLines, StyleHelp.Render(" Waiting for scan"))
	} else {
		linesPerDevice := 4 // 3 content + 1 blank
		maxVisible := devSpace / linesPerDevice
		if maxVisible < 1 {
			maxVisible = 1
		}

		// Compute viewport start so cursor is always visible
		viewStart := 0
		if cursorIndex >= maxVisible {
			viewStart = cursorIndex - maxVisible + 1
		}

		count := 0
		for i := viewStart; i < len(devices); i++ {
			isCursor := i == cursorIndex
			isHidden := hiddenDevices[devices[i].MAC]
			isIsolated := devices[i].MAC == isolateMAC

			entry := renderDeviceEntryFull(devices[i], innerW, isCursor, isHidden, isIsolated)
			for _, l := range entry {
				if count >= devSpace {
					break
				}
				devLines = append(devLines, l)
				count++
			}
			if count >= devSpace {
				break
			}
		}
	}

	// Truncate device lines if somehow over budget
	if len(devLines) > devSpace {
		devLines = devLines[:devSpace]
	}

	// Pad device lines to fill remaining space
	for len(devLines) < devSpace {
		devLines = append(devLines, "")
	}

	// Combine: header (fixed) + device entries (scrolled, exact fit)
	all := make([]string, 0, innerH)
	all = append(all, headerLines...)
	all = append(all, devLines...)

	// Hard clamp to innerH (safety)
	if len(all) > innerH {
		all = all[:innerH]
	}

	content := strings.Join(all, "\n")
	rendered := StylePanelBorder.Width(width - 2).Height(innerH).Render(content)

	// Hard clamp rendered output to exactly `height` lines.
	// lipgloss Height() only sets a minimum; it won't truncate overflow.
	outLines := strings.Split(rendered, "\n")
	if len(outLines) > height {
		outLines = outLines[:height]
	}
	for len(outLines) < height {
		outLines = append(outLines, "")
	}
	return strings.Join(outLines, "\n")
}

func renderDeviceEntryFull(d *bluetooth.Device, maxW int, isCursor, isHidden, isIsolated bool) []string {
	symbol := "*"
	tag := "[BLE]"
	switch d.Type {
	case bluetooth.DeviceTypeClassic:
		symbol = "B"
		tag = "[CLS]"
	case bluetooth.DeviceTypeWiFi:
		symbol = "W"
		tag = "[WiFi]"
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
	line3Extra := ""
	if d.Type == bluetooth.DeviceTypeWiFi && d.Band() != "" {
		line3Extra = fmt.Sprintf("  %s ch%d", d.Band(), d.Channel)
	}
	rawLine3 := fmt.Sprintf("       %s  %s%s", rssiStr, distStr, line3Extra)

	// Truncate to maxW to prevent line wrapping inside the panel
	rawLine1 = truncRaw(rawLine1, maxW)
	rawLine2 = truncRaw(rawLine2, maxW)
	rawLine3 = truncRaw(rawLine3, maxW)

	if isCursor {
		sty := cursorRowSty
		return []string{
			sty.Render(rawLine1),
			sty.Render(rawLine2),
			sty.Render(rawLine3),
			"",
		}
	}

	if isHidden {
		return []string{
			hiddenDevSty.Render(rawLine1),
			hiddenDevSty.Render(rawLine2),
			hiddenDevSty.Render(rawLine3),
			"",
		}
	}

	return renderNormalEntry(d, rawLine1, rawLine2, rawLine3, maxW, isIsolated)
}

func renderNormalEntry(d *bluetooth.Device, raw1, raw2, raw3 string, maxW int, isIsolated bool) []string {
	symbol := StyleDeviceTypeBLE.Render("*")
	typeTag := StyleDeviceTypeBLE.Render("[BLE]")
	switch d.Type {
	case bluetooth.DeviceTypeClassic:
		symbol = StyleDeviceTypeClassic.Render("B")
		typeTag = StyleDeviceTypeClassic.Render("[CLS]")
	case bluetooth.DeviceTypeWiFi:
		symbol = StyleDeviceTypeWiFi.Render("W")
		typeTag = StyleDeviceTypeWiFi.Render("[WiFi]")
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
	bandExtra := ""
	if d.Type == bluetooth.DeviceTypeWiFi && d.Band() != "" {
		bandExtra = StyleDeviceTypeWiFi.Render(fmt.Sprintf("  %s ch%d", d.Band(), d.Channel))
	}
	line3 := fmt.Sprintf("       %s  %s", StyleDeviceRSSI.Render(rssiStr), StyleDeviceDist.Render(distStr)) + bandExtra

	return []string{line1, line2, line3, ""}
}

// truncRaw pads or truncates a raw string to exactly w characters.
func truncRaw(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	if len(s) < w {
		return s + strings.Repeat(" ", w-len(s))
	}
	return s
}

func renderFilterBar(f FilterState, maxW int) string {
	toggleSty := func(on bool, label string) string {
		if on {
			return StyleFilterActive.Render("[" + label + "]")
		}
		return StyleFilterInactive.Render("[" + label + "]")
	}

	bar := " " + toggleSty(f.BLE, "1:BLE") + " " + toggleSty(f.Classic, "2:CLS") + " " + toggleSty(f.WiFi, "3:WiFi")

	if f.Active {
		bar += "  " + StyleFilterActive.Render("/"+f.Search+"_")
	} else if f.Search != "" {
		bar += "  " + StyleFilterInactive.Render("/"+f.Search)
	}

	return bar
}
