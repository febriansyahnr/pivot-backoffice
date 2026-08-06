package constant

import "time"

const (
	RateLimitConfigVariableMatchTypeExact    string = "EXACT"
	RateLimitConfigVariableMatchTypeContains string = "CONTAINS"
	RateLimitConfigVariableMatchTypePrefix   string = "PREFIX"

	RateLimitConfigTimeSecond string = "SECOND"
	RateLimitConfigTimeMinute string = "MINUTE"
	RateLimitConfigTimeHour   string = "HOUR"
	RateLimitConfigTimeDaily  string = "DAILY"

	RateLimitConfigTimeSecondDuration time.Duration = 1 * time.Second
	RateLimitConfigTimeMinuteDuration time.Duration = 1 * time.Minute
	RateLimitConfigTimeHourDuration   time.Duration = 1 * time.Hour
	RateLimitConfigTimeDailyDuration  time.Duration = 24 * time.Hour
)
