package merchantTopUp

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTopupSimulation(t *testing.T) {
	merchantId := "7d37b6e7-8af4-4218-8643-34fc3b3eb17e"
	request := snapCoreModel.TopupSimulationRequest{
		MerchantId: merchantId,
		VANumber:   "123456",
		TotalAmount: snapCoreModel.Amount{
			Value:    "10000.00",
			Currency: "IDR",
		},
	}
	logger := loggerMock.NewILogger(t)
	snapCore := repoMocks.NewISnapCoreRepository(t)
	merchantTopUpRepo := repoMocks.NewIMerchantTopUpRepository(t)

	config := &config.Config{}

	service := &merchantTopUpService{
		config: config, snapCore: snapCore, merchantTopUpRepo: merchantTopUpRepo, logger: logger,
	}

	testCases := []struct {
		name      string
		request   snapCoreModel.TopupSimulationRequest
		env       string
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "ERROR:Forbidden in production environment",
			env:       constant.EnvironmentProduction,
			setupMock: func() { /* Empty */ },
			wantErr:   true,
		},
		{
			name: "ERROR:Get topup reference detail",
			setupMock: func() {
				merchantTopUpRepo.On("GetByReferenceNumber", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Error", mock.Anything, "Failed while get merchant top up reference by reference number", mock.Anything,
				).Once().Return()
			},
			wantErr: true,
		},
		{
			name: "ERROR:Topup reference not found",
			setupMock: func() {
				merchantTopUpRepo.On("GetByReferenceNumber", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Forbidden access",
			setupMock: func() {
				merchantTopUpRepo.On("GetByReferenceNumber", mock.Anything, mock.Anything).Once().Return(&model.MerchantTopUp{}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				merchantTopUpRepo.On(
					"GetByReferenceNumber", mock.Anything, mock.Anything,
				).Return(&model.MerchantTopUp{
					MerchantID: merchantId,
				}, nil)
				snapCore.On("TopUpSimulation", mock.Anything, request).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Error", mock.Anything, "Failed to create top up simulation VA", mock.Anything,
				).Once().Return()
			},
			wantErr: true,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				snapCore.On("TopUpSimulation", mock.Anything, request).Return(&snapCoreModel.TopupSimulationResponseData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.request.VANumber == "" {
				tc.request = request
			}
			if tc.env == "" {
				tc.env = constant.EnvironmentStaging
			}

			tc.setupMock()
			config.Environment = tc.env

			_, err := service.CreateTopupSimulation(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			logger.AssertExpectations(t)
			snapCore.AssertExpectations(t)
		})
	}
}
