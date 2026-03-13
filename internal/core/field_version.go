package core

import "time"

// FieldVersion tracks the version metadata for a single task field,
// including which device last modified it and when.
type FieldVersion struct {
	DeviceID  string    `yaml:"device_id" json:"device_id"`
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
	Version   uint64    `yaml:"version" json:"version"`
}
