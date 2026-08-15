package disbursementRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListBulk(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *disbursementModel.GetBulkDisbursementFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.BulkDisbursementWithAggregate"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  &disbursementModel.GetBulkDisbursementFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.BulkDisbursementWithAggregate"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter:  &disbursementModel.GetBulkDisbursementFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.BulkDisbursementWithAggregate"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter: &disbursementModel.GetBulkDisbursementFilterRequest{
				MerchantID:     uuid.NewString(),
				Status:         constant.BulkDisbursementStatusPending,
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
				Sort:           "asc",
				SortBy:         "created_at",
				ReferenceID:    "test-uuid-123",
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.BulkDisbursementWithAggregate"),
					mock.AnythingOfType("string"),
				).Return(errors.New("some-error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)

			},
			filter:  &disbursementModel.GetBulkDisbursementFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetListBulk(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
