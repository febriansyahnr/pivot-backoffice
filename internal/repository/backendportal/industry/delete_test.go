package industry

import (
	"context"
	"errors"
	"testing"

	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	ctx := context.Background()
	uuid := "uuid-1"

	testCases := []struct {
		name      string
		uuid      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Delete industry",
			uuid: uuid,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			uuid: uuid,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			err := repo.Delete(ctx, tc.uuid)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
