package partition_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/partition"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReorganizeMonthlyRangePartition(t *testing.T) {
	logger := loggerMocks.NewILogger(t)
	service := serviceMocks.NewITablePartitionService(t)

	handler := New(logger, nil, service)

	loc, err := time.LoadLocation(constant.TimeLoc)
	require.NoError(t, err)

	tests := []struct {
		name      string
		request   partitionModel.ReorganizeRangePartitionRequest
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Invalid timezone", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				Datetime: types.Time{Time: time.Now().UTC()},
			},
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Running reorganize table partition", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: fmt.Errorf("calculations and processes using the %s time zone. your timezone UTC", constant.TimeLoc),
		},
		{
			name: "ERROR:Operation not allowed", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				Datetime: types.Time{Time: time.Date(2025, 10, 22, 17, 40, 0, 0, loc)},
			},
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Running reorganize table partition", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: fmt.Errorf("operation allowed only at the end of the month. your datetime 2025-10-22 17:40:00"),
		},
		{
			name: "ERROR:Some error", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				Datetime: types.Time{Time: time.Date(2025, 9, 30, 01, 15, 0, 0, loc)},
			},
			setupMock: func() {
				service.On("ReorganizeMonthlyRangePartition", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On(
					"Info", mock.Anything, "Running reorganize table partition", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				Datetime: types.Time{Time: time.Date(2025, 10, 31, 01, 15, 0, 0, loc)},
			},
			setupMock: func() {
				service.On("ReorganizeMonthlyRangePartition", mock.Anything, mock.Anything).Once().Return(nil)
				logger.On(
					"Info", mock.Anything, "Running reorganize table partition", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, handler.ReorganizeMonthlyRangePartition(t.Context(), test.request))

			logger.AssertExpectations(t)
			service.AssertExpectations(t)
		})
	}
}
