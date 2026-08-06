package merchantHandler_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/console/handler/merchant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMigrateBalanceHistoryToDataReporting(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	reportingSvc := serviceMocks.NewIReportingService(t)

	handler := New(logger, nil, reportingSvc)

	tz, _ := time.LoadLocation(constant.TimeLoc)
	logger.On(
		"Info", mock.Anything, "Running balance history migration to data reporting", mock.Anything, mock.Anything,
	).Return()

	tests := []struct {
		name         string
		startDateStr string
		endDateStr   string
		setupMock    func()
		wantError    error
	}{
		{
			name:         "ERROR: invalid start date format",
			startDateStr: "2024-13-01", // NOSONAR
			endDateStr:   "2024-01-31", // NOSONAR
			setupMock:    func() { /* Empty Function */ },
			wantError: func() error {
				_, err := time.ParseInLocation(time.DateTime, "2024-13-01 00:00:00", tz)
				return fmt.Errorf("parse start date: %w", err)
			}(),
		},
		{
			name:         "ERROR: invalid end date format",
			startDateStr: "2024-01-01", // NOSONAR
			endDateStr:   "2024-01-32", // NOSONAR
			setupMock:    func() { /* Empty Function */ },
			wantError: func() error {
				_, err := time.ParseInLocation(time.DateTime, "2024-01-32 23:59:59", tz)
				return fmt.Errorf("parse end date: %w", err)
			}(),
		},
		{
			name:         "ERROR: start date greater than end date",
			startDateStr: "2024-01-31", // NOSONAR
			endDateStr:   "2024-01-01", // NOSONAR
			setupMock:    func() { /* Empty Function */ },
			wantError:    errors.New("start date must not be after end date"),
		},
		{
			name:         "ERROR: service returns error",
			startDateStr: "2024-01-01", // NOSONAR
			endDateStr:   "2024-01-31", // NOSONAR
			setupMock: func() {
				reportingSvc.On("MigrateBalanceHistoryToDataReporting",
					mock.Anything,
					mock.MatchedBy(func(t time.Time) bool { return t.UTC().Format(time.DateTime) == "2023-12-31 17:00:00" }),
					mock.MatchedBy(func(t time.Time) bool { return t.UTC().Format(time.DateTime) == "2024-01-31 16:59:59" }),
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:         "SUCCESS: migrate balance history successfully",
			startDateStr: "2024-01-01", // NOSONAR
			endDateStr:   "2024-01-31", // NOSONAR
			setupMock: func() {
				reportingSvc.On("MigrateBalanceHistoryToDataReporting",
					mock.Anything,
					mock.MatchedBy(func(t time.Time) bool { return t.UTC().Format(time.DateTime) == "2023-12-31 17:00:00" }),
					mock.MatchedBy(func(t time.Time) bool { return t.UTC().Format(time.DateTime) == "2024-01-31 16:59:59" }),
				).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, handler.MigrateBalanceHistoryToDataReporting(t.Context(), test.startDateStr, test.endDateStr))
		})
	}
}
