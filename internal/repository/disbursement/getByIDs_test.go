package disbursementRepository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByIDs(t *testing.T) {
	disbursementIDs := []string{uuid.NewString(), uuid.NewString()}

	testCase := []struct {
		name      string
		payload   []string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:    "SUCCESS",
			payload: disbursementIDs,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.Disbursement"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: when empty",
			payload: []string{},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Mysql error",
			payload: disbursementIDs,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.Disbursement"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetByIDs(ctx, tc.payload)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
