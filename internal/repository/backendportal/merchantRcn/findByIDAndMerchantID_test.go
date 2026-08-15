package merchantRcn

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRcnRepository_FindByIDAndMerchantID(t *testing.T) {
	testCases := []struct {
		Name       string
		ID         string
		MerchantID string
		MockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		WantErr    bool
	}{
		{
			Name:       "Success Get Merchant Forbidden Usecase",
			MerchantID: uuid.Max.String(),
			ID:         uuid.Max.String(),
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:       "Error sql err no rows",
			MerchantID: uuid.Max.String(),
			ID:         uuid.Max.String(),
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			WantErr: false,
		},
		{
			Name:       "Error sql other errors",
			MerchantID: uuid.Max.String(),
			ID:         uuid.Max.String(),
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
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
			_, err := repo.FindByIDAndMerchantID(context.Background(), test.ID, test.MerchantID)
			if test.WantErr && err == nil {
				t.Errorf("GetMerchantForbiddenUsecase error: %v, wantErr: %v", err, test.WantErr)
			}

		})
	}
}
