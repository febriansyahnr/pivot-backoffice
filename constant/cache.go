package constant

import "time"

const (
	MaxRetryMechanism     int           = 3
	StandardCacheDuration time.Duration = time.Duration(24 * time.Hour) // 24 hours
)
