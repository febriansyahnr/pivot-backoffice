package merchantForbiddenUsecase

import (
	"context"
	"errors"
	"testing"

	merchantforbiddenusecaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckUseCase(t *testing.T) {

	testcases := []struct {
		Name      string
		Usecase   string
		WantErr   bool
		MockSetup func(repo *repoMocks.IMerchantForbiddenUsecaseRepository)
	}{
		{
			Name:    "SUCCESS: Usecase exists",
			Usecase: "DISBURSEMENT",
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository) {
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return([]*merchantforbiddenusecaseModel.MerchantForbiddenUseCase{}, nil)
			},
		},
		{
			Name:    "FAIL: Usecase not exists",
			Usecase: "DISBURSEMENTB",
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository) {
			},
		},
		{
			Name:    "ERROR: Error get forbidden usecase",
			Usecase: "DISBURSEMENT",
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository) {
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, errors.New("error"))

			},
		},
		{
			Name:    "ERROR: Usecase banned",
			Usecase: "DISBURSEMENT",
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository) {
				repo.On(
					"GetForbiddenUsecase", mock.Anything, mock.Anything,
				).Return(
					[]*merchantforbiddenusecaseModel.MerchantForbiddenUseCase{
						&merchantforbiddenusecaseModel.MerchantForbiddenUseCase{
							UseCase: "DISBURSEMENT"},
					}, nil)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			repo := &repoMocks.IMerchantForbiddenUsecaseRepository{}
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			testcase.MockSetup(repo)

			service := New(loggerMock, repo, nil, nil)
			err := service.CheckUseCase(context.TODO(), "", testcase.Usecase)

			if testcase.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
