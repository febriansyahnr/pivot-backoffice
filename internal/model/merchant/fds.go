package merchant

import (
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type FDSConfig struct {
	ProofOfPayment             *FDSFeatureProofOfPayment `json:"proofOfPayment" validate:"required_without=BypassExternalPaymentCheck"`
	BypassExternalPaymentCheck *bool                     `json:"bypassExternalPaymentCheck,omitempty"`
}

type FDSFeatureProofOfPayment struct {
	Velocity FDSRuleVelocityConfig `json:"velocity" validate:"required"`
}

type FDSRuleVelocityConfig struct {
	Enabled   bool               `json:"enabled"` // Placeholder for future improvements.
	Window    FDSWindowConfig    `json:"window" validate:"required"`
	Threshold FDSThresholdConfig `json:"threshold" validate:"required"`
	Action    string             `json:"action" validate:"required,oneof=BLOCK"` // Placeholder for future improvements.
}

type FDSWindowConfig struct {
	Interval int    `json:"interval" validate:"required,min=1"`
	Unit     string `json:"unit" validate:"required,oneof=SECOND MINUTE HOUR DAY"`
}

func (w *FDSWindowConfig) Period() time.Duration {
	if w.Unit == constant.WindowUnitDay {
		return time.Duration(w.Interval) * (24 * time.Hour)
	}

	var unit string

	switch w.Unit {
	default:
		unit = "s"

	case constant.WindowUnitMinute:
		unit = "m"

	case constant.WindowUnitHour:
		unit = "h"
	}

	duration, _ := time.ParseDuration(fmt.Sprintf("%d%s", w.Interval, unit))
	// Error can be safely ignored as format is guaranteed valid by validation
	return duration
}

type FDSThresholdConfig struct {
	Count int `json:"count" validate:"required,min=1"`
}
