package types

import "time"

// IsActive returns true if the pause should currently be enforced.
func (p PauseInfo) IsActive(now time.Time) bool {
	if !p.Paused {
		return false
	}
	if p.ResumeAt == nil {
		return true
	}
	return now.Before(*p.ResumeAt)
}

// WithTimestamps returns a copy of the pause info with the provided resume and update timestamps.
func (p PauseInfo) WithTimestamps(resumeAt *time.Time, updatedAt time.Time) PauseInfo {
	p.ResumeAt = nil
	if resumeAt != nil {
		ts := resumeAt.UTC()
		p.ResumeAt = &ts
	}
	ts := updatedAt.UTC()
	p.UpdatedAt = &ts
	return p
}

// IsFrozen returns true if the freeze is currently enforced.
func (f FreezeInfo) IsFrozen(now time.Time) bool {
	if !f.Frozen {
		return false
	}
	if f.UnfreezeAt == nil {
		return true
	}
	return now.Before(*f.UnfreezeAt)
}

// WithTimestamps applies the provided timestamps to the freeze info.
func (f FreezeInfo) WithTimestamps(unfreezeAt *time.Time, updatedAt time.Time) FreezeInfo {
	f.UnfreezeAt = nil
	if unfreezeAt != nil {
		ts := unfreezeAt.UTC()
		f.UnfreezeAt = &ts
	}
	ts := updatedAt.UTC()
	f.UpdatedAt = &ts
	return f
}

// Ready indicates whether the pending params are ready to be applied.
func (pp PendingParams) Ready(now time.Time, height int64) bool {
	if pp.ApplyHeight != 0 && uint64(height) >= pp.ApplyHeight {
		return true
	}
	if !pp.ApplyTime.IsZero() && !now.Before(pp.ApplyTime) {
		return true
	}
	return false
}
