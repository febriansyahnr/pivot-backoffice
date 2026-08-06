package ipWhitelistRepository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	id := uuid.NewString()

	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					constant.TimeMockType(),
					constant.TimeMockType(),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					constant.TimeMockType(),
					constant.TimeMockType(),
					mock.AnythingOfType("string"),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockMysql := mysqlMocks.NewIMySqlExt(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			tc.mockSetup(mockMysql)

			err := New(logger, mockMysql).Delete(context.Background(), id)
			if (err != nil) != tc.wantErr {
				t.Errorf("IPWhitelist.Delete() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
