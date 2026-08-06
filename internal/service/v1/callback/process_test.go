package callbackService_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	loggerPdkMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	callbackPartnerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/callback"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	callbackPartner "github.com/paper-indonesia/pivot-backoffice/pkg/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCallbackPaymentCallback(t *testing.T) {
	merchantID := uuid.New()

	callbackRequest := `{
		"noReference" : "REFF123456",
		"Status" : "SUCCESS"
	}`

	type UnserializableType struct {
		Ch chan int
	}

	unserializableType := UnserializableType{
		Ch: make(chan int),
	}

	validRequest := &callbackModel.ProcessCallbackRequest{
		Name:       "Virtual Account",
		MerchantID: merchantID,
		Request:    callbackRequest,
	}

	invalidRequest := &callbackModel.ProcessCallbackRequest{
		Name:       "Virtual Account",
		MerchantID: merchantID,
		Request:    unserializableType,
	}

	callbackMasterDB := &callbackModel.Callback{
		UUID:             uuid.New(),
		CallbackMasterID: uuid.New(),
		MerchantID:       merchantID,
		URL:              "https://localhost/v1/payment-callback",
		Description:      "API",
	}

	testCases := []struct {
		name       string
		input      *callbackModel.ProcessCallbackRequest
		mocksSetup func(

			callbackRepo *repositoryMocks.ICallbackRepository,
			callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
			merchantSvc *serviceMocks.IMerchantService)
		wantErr bool
	}{
		{
			name:  "SUCCESS: successfully consume callback",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("some-response", nil)

				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(&validApiKey, nil)

				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: GetOrGenerateCallbackApiKey",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: when decode request",
			input: invalidRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(&validApiKey, nil)

			},
			wantErr: true,
		},
		{
			name:  "ERROR: when find callback",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: partner not yet register callback",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(nil, nil)

			},
			wantErr: true,
		},
		{
			name:  "ERROR: when call api callback partner service",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(&validApiKey, nil)

				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("", constant.ErrSomeErrorForUnitTest)

				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: when create callback logs",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(&validApiKey, nil)

				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: when update callback logs",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(&validApiKey, nil)

				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("", nil)

				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Update callback",
			input: validRequest,
			mocksSetup: func(

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *serviceMocks.IMerchantService) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					constant.StringMockType(),
					constant.UuidMockType()).
					Return(callbackMasterDB, nil)

				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("", constant.ErrSomeErrorForUnitTest)

				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(nil)

				merchantSvc.On("GetOrGenerateCallbackApiKey",
					constant.ValueCtxMockType(),
					constant.StringMockType()).
					Return(&validApiKey, nil)

				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType()).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCache := redisExtMocks.NewIRedisExt(t)
			callbackRepoMock := repositoryMocks.NewICallbackRepository(t)
			callbackPartnerSvc := callbackPartnerMocks.NewICallbackPartner(t)
			merchantSvc := serviceMocks.NewIMerchantService(t)
			tt.mocksSetup(callbackRepoMock, callbackPartnerSvc, merchantSvc)

			callbackSvc := New(mockLogger, mockCache, callbackRepoMock, callbackPartnerSvc, merchantSvc)
			ctx := context.Background()
			err := callbackSvc.ProcessCallback(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			callbackRepoMock.AssertExpectations(t)
		})
	}
}

func TestSendMerchantCallback(t *testing.T) {

	log := loggerPdkMock.NewILogger(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	callbackRepo := repositoryMocks.NewICallbackRepository(t)
	callbackPartnerSvc := callbackPartnerMocks.NewICallbackPartner(t)

	service := New(log, nil, callbackRepo, callbackPartnerSvc, merchantSvc)

	tests := []struct {
		name       string
		request    callbackModel.SendMerchantCallbackRequest
		setupMock  func()
		wantError  error
		wantResult *callbackModel.SendMerchantCallbackResponse
	}{
		{
			name: "ERROR:Snap find callback details",
			request: callbackModel.SendMerchantCallbackRequest{
				IsSnap: true,
			},
			setupMock: func() {
				callbackRepo.On("FindCallbackByNameAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "error when find callback token", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Get or generate callback api key",
			request: callbackModel.SendMerchantCallbackRequest{},
			setupMock: func() {
				merchantSvc.On("GetOrGenerateCallbackApiKey", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Send callback",
			request: callbackModel.SendMerchantCallbackRequest{},
			setupMock: func() {
				merchantSvc.On("GetOrGenerateCallbackApiKey", mock.Anything, mock.Anything).Return(util.ValueToPtr("x-api-key"), nil)
				callbackPartnerSvc.On("Callback", mock.Anything, mock.Anything, mock.Anything).Once().Return("", assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "SUCCESS:Merchant responded with non-2xx status",
			request: callbackModel.SendMerchantCallbackRequest{},
			setupMock: func() {
				callbackPartnerSvc.On("Callback", mock.Anything, mock.Anything, mock.Anything).Once().Return(string(callbackPartner.ErrTestBadRequest.ResponseBody()), callbackPartner.ErrTestBadRequest)
			},
			wantResult: &callbackModel.SendMerchantCallbackResponse{
				StatusCode:   http.StatusBadRequest,
				ResponseBody: callbackPartner.ErrTestBadRequest.ResponseBody(),
			},
		},
		{
			name:    "SUCCESS:Merchant responded with 2xx status",
			request: callbackModel.SendMerchantCallbackRequest{},
			setupMock: func() {
				callbackPartnerSvc.On("Callback", mock.Anything, mock.Anything, mock.Anything).Once().Return(`{"message":"OK"}`, nil) // NOSONAR
			},
			wantResult: &callbackModel.SendMerchantCallbackResponse{
				StatusCode:   http.StatusOK,
				ResponseBody: []byte(`{"message":"OK"}`), // NOSONAR
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.SendMerchantCallback(t.Context(), test.request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			log.AssertExpectations(t)
			merchantSvc.AssertExpectations(t)
			callbackRepo.AssertExpectations(t)
			callbackPartnerSvc.AssertExpectations(t)
		})
	}
}

func TestWriteCallbackLogFromWorkflowTask(t *testing.T) {
	callbackRepo := repositoryMocks.NewICallbackRepository(t)

	service := New(nil, nil, callbackRepo, nil, nil)

	tests := []struct {
		name      string
		log       callbackModel.WorkflowWriteLogRequest
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			log: callbackModel.WorkflowWriteLogRequest{
				IsSnap: true,
				Response: callbackModel.WorkflowSendCallbackResponse{
					StatusCode:     http.StatusInternalServerError,
					AdditionalInfo: &callbackModel.SendMerchantCallbackAdditionalInfo{},
				},
			},
			setupMock: func() {
				callbackRepo.On("CreateCallbackLog", mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("write callback log: %v", assert.AnError),
		},
		{
			name: "SUCCESS", // NOSONAR
			log:  callbackModel.WorkflowWriteLogRequest{},
			setupMock: func() {
				callbackRepo.On("CreateCallbackLog", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.WriteCallbackLogFromWorkflowTask(t.Context(), test.log)

			assert.Equal(t, test.wantError, err)
			if test.wantError == nil {
				assert.NotNil(t, result)
			}
		})
	}
}
