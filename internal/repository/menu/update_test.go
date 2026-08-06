package menuRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {

	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   error
	}{
		{
			name: "when the update was succeeded, then should not return an error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*menuModel.Menu"),
				).Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "when failed to update the menu, then should return an error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*menuModel.Menu"),
				).Return(false, errors.New("insert error"))
			},
			wantErr: errors.New("insert error"),
		},
		{
			name: "when no rows affected, then should return an error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.AnythingOfType("*menuModel.Menu"),
				).Return(false, nil)
			},
			wantErr: constant.ErrNoRowsAffected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Update(context.Background(), &menuModel.Menu{})

			assert.Equal(t, tc.wantErr, err)
		})
	}
}
