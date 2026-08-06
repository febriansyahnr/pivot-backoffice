package callbackService_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	callbackPartnerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/callback"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	merchantSvcMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResendCallback(t *testing.T) {
	merchantID := uuid.NewString()
	event := constant.CallbackEventPayoutDone
	callbackLog := &callbackModel.CallbackLogWithMaster{
		UUID:       uuid.New(),
		MerchantID: merchantID,
		Event:      &event,
		Request:    `{"event":"PAYOUT.PENDING","data":{"test":"OK"}}`,
	}

	testCases := []struct {
		Name       string
		MerchantID string
		WantErr    bool
		MockSetup  func(mockRepo *repositoryMocks.ICallbackRepository, merchantSvc *merchantSvcMocks.IMerchantService, callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner)
	}{
		{
			Name:       "SUCCESS",
			MerchantID: merchantID,
			WantErr:    false,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, merchantSvc *merchantSvcMocks.IMerchantService, callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(callbackLog, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string")).
					Return(&validApiKey, nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return("some-response", nil)

				mockRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)
			},
		},
		{
			Name:       "ERROR: GetCallbackLogByID error",
			MerchantID: merchantID,
			WantErr:    true,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, merchantSvc *merchantSvcMocks.IMerchantService, callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name:       "ERROR: GetOrGenerateCallbackApiKey error",
			MerchantID: merchantID,
			WantErr:    true,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, merchantSvc *merchantSvcMocks.IMerchantService, callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(callbackLog, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string")).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name:       "ERROR: Call Callback error",
			MerchantID: merchantID,
			WantErr:    false,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, merchantSvc *merchantSvcMocks.IMerchantService, callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(callbackLog, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string")).
					Return(&validApiKey, nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return("", constant.ErrSomeErrorForUnitTest)

				mockRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)

			},
		},
		{
			Name:       "ERROR: UpdateCallbackLog error",
			MerchantID: merchantID,
			WantErr:    true,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, merchantSvc *merchantSvcMocks.IMerchantService, callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(callbackLog, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string")).
					Return(&validApiKey, nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return("", constant.ErrSomeErrorForUnitTest)

				mockRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(constant.ErrSomeErrorForUnitTest)

			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewICallbackRepository(t)
			merchantSvc := merchantSvcMocks.NewIMerchantService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCache := redisExtMocks.NewIRedisExt(t)
			callbackPartnerSvc := callbackPartnerMocks.NewICallbackPartner(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo, merchantSvc, callbackPartnerSvc)

			cbService := New(mockLogger, mockCache, mockRepo, callbackPartnerSvc, merchantSvc)
			err := cbService.ResendCallback(ctx, uuid.NewString(), tc.MerchantID)
			if tc.WantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
