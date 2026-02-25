package bluetooth

import (
	"sort"
	"sync"
	"time"

	"ble-radar.klederson.com/internal/config"
)

// DeviceStore is a thread-safe store for discovered devices.
type DeviceStore struct {
	mu      sync.RWMutex
	devices map[string]*Device
}

// NewDeviceStore creates a new empty DeviceStore.
func NewDeviceStore() *DeviceStore {
	return &DeviceStore{
		devices: make(map[string]*Device),
	}
}

// Upsert adds or updates a device. If the device already exists, RSSI is
// smoothed using EMA and the angle is preserved for position consistency.
func (s *DeviceStore) Upsert(mac, name string, rssi float64, dtype DeviceType) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if existing, ok := s.devices[mac]; ok {
		// Update existing device with EMA smoothing
		existing.RSSI = existing.RSSI*(1-config.SmoothingAlpha) + rssi*config.SmoothingAlpha
		existing.Distance = RSSIToDistance(existing.RSSI, config.MeasuredPower, config.PathLossExp)
		existing.LastSeen = now
		if name != "" {
			existing.Name = name
		}
		return
	}

	// New device
	angle := MacToAngle(mac)
	dist := RSSIToDistance(rssi, config.MeasuredPower, config.PathLossExp)

	s.devices[mac] = &Device{
		MAC:       mac,
		Name:      name,
		RSSI:      rssi,
		Type:      dtype,
		LastSeen:  now,
		Angle:     angle,
		Distance:  dist,
		Elevation: MacToElevation(mac),
	}
}

// Evict removes devices not seen within the timeout duration.
// Returns the number of evicted devices.
func (s *DeviceStore) Evict(timeout time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-timeout)
	count := 0
	for mac, dev := range s.devices {
		if dev.LastSeen.Before(cutoff) {
			delete(s.devices, mac)
			count++
		}
	}
	return count
}

// Snapshot returns a sorted copy of all devices (strongest RSSI first).
func (s *DeviceStore) Snapshot() []*Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Device, 0, len(s.devices))
	for _, d := range s.devices {
		// Copy device to avoid data races
		cp := *d
		result = append(result, &cp)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RSSI > result[j].RSSI // Strongest first (less negative)
	})
	return result
}

// Count returns the total number of tracked devices.
func (s *DeviceStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.devices)
}

// CountByType returns counts broken down by device type.
func (s *DeviceStore) CountByType() (ble, classic int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.devices {
		if d.Type == DeviceTypeClassic {
			classic++
		} else {
			ble++
		}
	}
	return
}
