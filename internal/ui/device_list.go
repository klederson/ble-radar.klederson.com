package ui

import (
	"fmt"
	"strings"

	"ble-radar.klederson.com/internal/bluetooth"
)

// RenderDeviceList renders the scrollable device list panel.
func RenderDeviceList(devices []*bluetooth.Device, width, height int, scrollOffset int) string {
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
		for i, d := range devices {
			if i < scrollOffset {
				continue
			}
			entry := renderDeviceEntry(d, innerW)
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

func renderDeviceEntry(d *bluetooth.Device, maxW int) []string {
	symbol := StyleDeviceTypeBLE.Render("*")
	typeTag := StyleDeviceTypeBLE.Render("[BLE]")
	if d.Type == bluetooth.DeviceTypeClassic {
		symbol = StyleDeviceTypeClassic.Render("B")
		typeTag = StyleDeviceTypeClassic.Render("[CLS]")
	}

	name := d.DisplayName()
	// Truncate name to fit
	nameMax := maxW - 8
	if nameMax < 4 {
		nameMax = 4
	}
	if len(name) > nameMax {
		name = name[:nameMax]
	}

	line1 := fmt.Sprintf(" %s %s %s", symbol, StyleDeviceName.Render(name), typeTag)

	mac := d.MAC
	if len(mac) > maxW-2 {
		mac = mac[:maxW-2]
	}
	line2 := fmt.Sprintf("   %s", StyleDeviceMAC.Render(mac))

	rssiStr := fmt.Sprintf("%ddBm", int(d.RSSI))
	distStr := fmt.Sprintf("~%.1fm", d.Distance)
	line3 := fmt.Sprintf("   %s  %s",
		StyleDeviceRSSI.Render(rssiStr),
		StyleDeviceDist.Render(distStr))

	return []string{line1, line2, line3, ""}
}
