package bluetooth

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"time"
)

// DeviceType distinguishes BLE from Classic Bluetooth.
type DeviceType int

const (
	DeviceTypeBLE     DeviceType = iota
	DeviceTypeClassic
	DeviceTypeWiFi
)

func (dt DeviceType) String() string {
	switch dt {
	case DeviceTypeClassic:
		return "Classic"
	case DeviceTypeWiFi:
		return "WiFi"
	default:
		return "BLE"
	}
}

// Device represents a discovered Bluetooth or WiFi device.
type Device struct {
	MAC       string
	Name      string
	RSSI      float64
	Type      DeviceType
	LastSeen  time.Time
	Angle     float64 // Radians, 0=north, clockwise
	Distance  float64 // Estimated distance in meters
	Elevation float64 // [-1, +1], 0=same level, +1=above, -1=below
	Frequency int     // MHz (e.g. 2437, 5180). Zero for BLE/Classic.
	Channel   int     // WiFi channel number. Zero for BLE/Classic.
}

// Symbol returns the radar character for this device type.
func (d *Device) Symbol() string {
	switch d.Type {
	case DeviceTypeClassic:
		return "B"
	case DeviceTypeWiFi:
		return "W"
	default:
		return "*"
	}
}

// Band returns the WiFi frequency band label ("2.4G", "5G", or "").
func (d *Device) Band() string {
	if d.Frequency >= 5000 {
		return "5G"
	}
	if d.Frequency >= 2400 {
		return "2.4G"
	}
	return ""
}

// DisplayName returns the device name or "[unnamed]" if empty.
func (d *Device) DisplayName() string {
	if d.Name == "" {
		return "[unnamed]"
	}
	return d.Name
}

// MacToAngle derives a consistent angle from a MAC address using a hash.
// Returns radians in [0, 2π), where 0=north, increasing clockwise.
func MacToAngle(mac string) float64 {
	h := sha256.Sum256([]byte(mac))
	val := binary.BigEndian.Uint32(h[:4])
	return float64(val) / float64(math.MaxUint32) * 2 * math.Pi
}

// MacToElevation derives a consistent elevation from a MAC address using a hash.
// Returns a value in [-1, +1], where -1=below, 0=level, +1=above.
// Uses bytes 4-7 of SHA256 (separate from angle which uses 0-3).
func MacToElevation(mac string) float64 {
	h := sha256.Sum256([]byte(mac))
	val := binary.BigEndian.Uint32(h[4:8])
	return float64(val)/float64(math.MaxUint32)*2 - 1
}

// RSSIToDistance estimates distance from RSSI using the log-distance path loss model.
// Formula: d = 10^((measuredPower - rssi) / (10 * n))
func RSSIToDistance(rssi, measuredPower, pathLossExp float64) float64 {
	if rssi >= 0 {
		return 0.1
	}
	d := math.Pow(10, (measuredPower-rssi)/(10*pathLossExp))
	if d < 0.1 {
		return 0.1
	}
	return d
}
