package types

import "cosmossdk.io/errors"

var (
	ErrPremiumPlanNotFound    = errors.Register(ModuleName, 2100, "premium plan not found")
	ErrPremiumOverdue         = errors.Register(ModuleName, 2101, "premium overdue")
	ErrPremiumUnauthorized    = errors.Register(ModuleName, 2102, "unauthorized")
	ErrInvalidPremiumSchedule = errors.Register(ModuleName, 2103, "invalid premium schedule")
)
