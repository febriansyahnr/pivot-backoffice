package merchantForbiddenUsecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantForbiddenUseCaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantForbiddenUsecase"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/mock"
)

func Test_RemoveForbiddenUsecase(t *testing.T) {
	testCases := []struct {
		Name      string
		Input     *merchantForbiddenUseCaseModel.MerchantForbiddenUseCase
		MockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		WantErr   bool
	}{
		{
			Name: "Success Remove Merchant Forbidden Usecase",
			Input: &merchantForbiddenUseCaseModel.MerchantForbiddenUseCase{
				MerchantID: uuid.Max.String(),
				UseCase:    "usecase",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			WantErr: false,
		},
		{
			Name: "Error no rows affected",
			Input: &merchantForbiddenUseCaseModel.MerchantForbiddenUseCase{
				MerchantID: uuid.Max.String(),
				UseCase:    "usecase",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, nil)
			},
			WantErr: false,
		},
		{
			Name: "Error sql other errors",
			Input: &merchantForbiddenUseCaseModel.MerchantForbiddenUseCase{
				MerchantID: uuid.Max.String(),
				UseCase:    "usecase",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("error"))
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
			err := repo.RemoveForbiddenUsecase(context.TODO(), test.Input)
			if test.WantErr && err == nil {
				t.Errorf("RegisterForbiddenUsecase error: %v, wantErr: %v", err, test.WantErr)
			}

		})
	}

}
