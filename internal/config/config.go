package config

import "time"

const (
	// RSSI to distance estimation
	MeasuredPower = -59.0 // RSSI at 1 meter (dBm)
	PathLossExp   = 2.5   // Path loss exponent (N)

	// Radar display
	MaxRange      = 30.0 // Maximum range in meters
	AspectRatio   = 0.5  // Terminal char aspect correction (chars are ~2:1 tall)
	RingCount     = 4    // Number of concentric rings
	SweepSpeedRPM = 30   // Sweep rotations per minute (1 rotation per 2 seconds)
	SweepTrailDeg = 60.0 // Sweep trail angle in degrees
	TargetFPS     = 30   // Target frames per second

	// Device management
	DeviceTimeout  = 30 * time.Second // Remove devices not seen for this long
	EvictInterval  = 5 * time.Second  // How often to run eviction
	SmoothingAlpha = 0.3              // EMA smoothing factor (30% new, 70% old)

	// Scanner
	ScanInterval   = 100 * time.Millisecond // BLE scan callback throttle
	ClassicScanSec = 8                      // hcitool scan duration in seconds

	// Demo mode
	DemoDeviceMin = 8  // Minimum fake devices
	DemoDeviceMax = 12 // Maximum fake devices

	// App
	AppName    = "BLE-RADAR"
	AppVersion = "1.0"
)
