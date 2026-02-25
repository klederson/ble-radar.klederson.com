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
)

func (dt DeviceType) String() string {
	if dt == DeviceTypeClassic {
		return "Classic"
	}
	return "BLE"
}

// Device represents a discovered Bluetooth device.
type Device struct {
	MAC      string
	Name     string
	RSSI     float64
	Type     DeviceType
	LastSeen time.Time
	Angle    float64 // Radians, 0=north, clockwise
	Distance float64 // Estimated distance in meters
}

// Symbol returns the radar character for this device type.
func (d *Device) Symbol() string {
	if d.Type == DeviceTypeClassic {
		return "B"
	}
	return "*"
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
