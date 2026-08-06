package merchantForbiddenUsecase

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

	merchantForbiddenUseCaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
)

func Test_GetMerchantForbiddenUsecase(t *testing.T) {

	testCases := []struct {
		Name      string
		Input     *merchantForbiddenUseCaseModel.GetMerchantForbiddenUseCaseRequest
		MockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		WantErr   bool
	}{
		{
			Name: "Success Get Merchant Forbidden Usecase",
			Input: &merchantForbiddenUseCaseModel.GetMerchantForbiddenUseCaseRequest{
				MerchantID: uuid.Max.String(),
				UseCase:    "usecase",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
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
			Name: "Error sql err no rows",
			Input: &merchantForbiddenUseCaseModel.GetMerchantForbiddenUseCaseRequest{
				MerchantID: uuid.Max.String(),
				UseCase:    "usecase",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
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
			Name: "Error sql other errors",
			Input: &merchantForbiddenUseCaseModel.GetMerchantForbiddenUseCaseRequest{
				MerchantID: uuid.Max.String(),
				UseCase:    "usecase",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
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
			_, err := repo.GetForbiddenUsecase(context.TODO(), test.Input)
			if test.WantErr && err == nil {
				t.Errorf("GetMerchantForbiddenUsecase error: %v, wantErr: %v", err, test.WantErr)
			}

		})
	}
}
