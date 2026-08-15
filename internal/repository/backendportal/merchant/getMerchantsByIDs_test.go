package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantsByIDs(t *testing.T) {
	testCases := []struct {
		Name      string
		MockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		WantErr   bool
	}{
		{
			Name: "SUCCESS ",
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "Error sql no rows",
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)

			},
			WantErr: false,
		},

		{
			Name: "Error return other errors",
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("error"))

			},
			WantErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			test.MockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetMerchantsByIDs(context.Background(), []string{uuid.NewString(), uuid.NewString()})

			if test.WantErr {
				assert.NotNil(t, err)
			}
		})
	}

}
