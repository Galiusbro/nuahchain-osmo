package types

import "time"

const (
	// SensitiveParamChangeDelay defines the minimum delay before applying queued governance parameter changes.
	SensitiveParamChangeDelay = time.Hour

	// DefaultEmergencyThreshold defines the number of signatures required to execute emergency actions.
	// The product request specifies that a single vote should be sufficient, hence the value of 1.
	DefaultEmergencyThreshold = uint32(1)
)
