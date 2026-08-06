package callbackService

import (
	"context"
	"errors"
	"testing"

	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"

	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCallbackRegisterCallback(t *testing.T) {
	request := &callbackModel.RegisterCallbackRequest{
		MerchantID:  uuid.New(),
		Name:        "Virtual Account",
		URL:         "https://localhost/v1/payment-callback",
		Description: "API",
	}

	callbackMasterDB := &callbackModel.CallbackMaster{
		UUID:        uuid.New(),
		Name:        "Virtual Account",
		Description: "API",
	}

	testCases := []struct {
		name       string
		input      *callbackModel.RegisterCallbackRequest
		mocksSetup func(callbackMasterRepo *repositoryMocks.ICallbackRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully register callback (callback master already created before)",
			input: request,
			mocksSetup: func(callbackMasterRepo *repositoryMocks.ICallbackRepository) {
				callbackMasterRepo.On("FindCallbackMasterByName",
					mock.Anything,
					mock.AnythingOfType("string")).
					Return(callbackMasterDB, nil)

				callbackMasterRepo.On("CreateCallback",
					mock.Anything,
					mock.AnythingOfType("*callback_model.Callback")).
					Return(nil)

			},
			wantErr: false,
		},
		{
			name:  "SUCCESS: successfully register callback (callback master not yet create | fresh data)",
			input: request,
			mocksSetup: func(callbackMasterRepo *repositoryMocks.ICallbackRepository) {
				callbackMasterRepo.On("FindCallbackMasterByName",
					mock.Anything,
					mock.AnythingOfType("string")).
					Return(nil, nil)

				callbackMasterRepo.On("CreateCallbackMaster",
					mock.Anything,
					mock.AnythingOfType("*callback_model.CallbackMaster")).
					Return(nil)

				callbackMasterRepo.On("CreateCallback",
					mock.Anything,
					mock.AnythingOfType("*callback_model.Callback")).
					Return(nil)

			},
			wantErr: false,
		},
		{
			name:  "ERROR: when find callback master",
			input: request,
			mocksSetup: func(callbackMasterRepo *repositoryMocks.ICallbackRepository) {
				callbackMasterRepo.On("FindCallbackMasterByName",
					mock.Anything,
					mock.AnythingOfType("string")).
					Return(nil, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: when create callback",
			input: request,
			mocksSetup: func(callbackMasterRepo *repositoryMocks.ICallbackRepository) {
				callbackMasterRepo.On("FindCallbackMasterByName",
					mock.Anything,
					mock.AnythingOfType("string")).
					Return(callbackMasterDB, nil)

				callbackMasterRepo.On("CreateCallback",
					mock.Anything,
					mock.AnythingOfType("*callback_model.Callback")).
					Return(errors.New("some-error"))

			},
			wantErr: true,
		},
		{
			name:  "ERROR: when create callback master",
			input: request,
			mocksSetup: func(callbackMasterRepo *repositoryMocks.ICallbackRepository) {
				callbackMasterRepo.On("FindCallbackMasterByName",
					mock.Anything,
					mock.AnythingOfType("string")).
					Return(nil, nil)

				callbackMasterRepo.On("CreateCallbackMaster",
					mock.Anything,
					mock.AnythingOfType("*callback_model.CallbackMaster")).
					Return(errors.New("some-error"))

			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCache := redisExtMocks.NewIRedisExt(t)
			callbackRepoMock := repositoryMocks.NewICallbackRepository(t)
			tt.mocksSetup(callbackRepoMock)

			callbackSvc := New(mockLogger, mockCache, callbackRepoMock, nil, nil)
			ctx := context.Background()
			resp, err := callbackSvc.RegisterCallback(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, resp)
				assert.NoError(t, err)
			}

			callbackRepoMock.AssertExpectations(t)
		})
	}
}
