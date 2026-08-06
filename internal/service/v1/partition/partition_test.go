package partitionService_test

import (
	"testing"

	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/partition"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReorganizeDailyRangePartition(t *testing.T) {
	repo := mocks.NewITablePartitionRepository(t)

	service := New(repo)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On("ReorganizeMonthlyRangePartition", mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("ReorganizeMonthlyRangePartition", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, service.ReorganizeMonthlyRangePartition(t.Context(), partitionModel.ReorganizeRangePartitionRequest{}))
		})
	}
}
