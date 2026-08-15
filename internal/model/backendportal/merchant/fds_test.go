package merchant_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/merchant"

	"github.com/stretchr/testify/assert"
)

func TestFDSWindowConfigPeriod(t *testing.T) {

	tests := []struct {
		data       FDSWindowConfig
		wantResult time.Duration
	}{
		{
			data: FDSWindowConfig{
				Interval: 30,
				Unit:     constant.WindowUnitSecond,
			},
			wantResult: 30 * time.Second,
		},
		{
			data: FDSWindowConfig{
				Interval: 15,
				Unit:     constant.WindowUnitMinute,
			},
			wantResult: 15 * time.Minute,
		},
		{
			data: FDSWindowConfig{
				Interval: 1,
				Unit:     constant.WindowUnitHour,
			},
			wantResult: time.Hour,
		},
		{
			data: FDSWindowConfig{
				Interval: 3,
				Unit:     constant.WindowUnitDay,
			},
			wantResult: 3 * (24 * time.Hour),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantResult, test.data.Period())
	}
}
