package callbackService_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	callbackPartnerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/callback"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	merchantSvcMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapVACallback(t *testing.T) {
	merchantID := uuid.New()
	baseUrl := "https://localhost"
	callbackPath := "/v1/payment-callback"
	description := "SNAP API"
	cacheAccessTokenBody := "{\"responseCode\":\"2007300\",\"responseMessage\":\"Successful\",\"accessToken\":\"token\",\"tokenType\":\"Bearer\",\"expiresIn\":\"900\"}"

	validRequest := &callbackModel.ProcessCallbackRequest{
		Name:       constant.CallbackNamePayment,
		Event:      constant.CallbackEventPaymentVirtualAccountPaid,
		MerchantID: merchantID,
		Request: map[string]any{
			"referenceId": "BAYAR123456",
			"customer": map[string]any{
				"customerId": "CUST123456",
				"name":       "John Doe",
			},
			"merchantID": merchantID,
			"virtualAccount": map[string]any{
				"isSnap": true,
			},
			"totalAmount": map[string]any{
				"currency": "IDR",
				"value":    "100000.00",
			},
		},
		IsSnap: true,
	}

	vaultTransit := vaultMock.NewIVaultTransit(t)

	testCases := []struct {
		name       string
		request    *callbackModel.ProcessCallbackRequest
		mocksSetup func(
			request *callbackModel.ProcessCallbackRequest,
			cacheMock *redisExtMocks.IRedisExt,

			callbackRepo *repositoryMocks.ICallbackRepository,
			callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
			merchantSvc *merchantSvcMocks.IMerchantService,
			merchantRepo *repositoryMocks.IMerchantRepository,
		)
		wantErr bool
	}{
		{
			name:    "SUCCESS: snap va callback",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Return(&redis.StringCmd{})
				privateKey, _ := generateMockPrivateKey()
				merchantSvc.On("GetSnapPrivateKey",
					mock.Anything,
					mock.Anything,
				).Return(privateKey, nil)
				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return(`{"accessToken":"token"}`, nil)
				cacheMock.On("Set",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&redis.StatusCmd{})
				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType(),
				).Return(nil)
				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType(),
				).Return(nil)
				merchantRepo.On("GetMerchantSnapPKCS8KeyByMerchantID",
					mock.Anything,
					constant.StringMockType(),
				).Return(&merchantModel.MerchantAuth{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: snap va callback with cached b2b token",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				strCmd := redis.NewStringCmd(context.Background(), "token")
				strCmd.SetVal("token")
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd)
				strCmd2 := redis.NewStringCmd(context.Background(), "tokenBody")
				strCmd2.SetVal(cacheAccessTokenBody)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd2)
				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return(`{"responseCode":"400"}`, nil)
				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType(),
				).Return(nil)
				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType(),
				).Return(nil)
				merchantRepo.On("GetMerchantSnapPKCS8KeyByMerchantID",
					mock.Anything,
					constant.StringMockType(),
				).Return(&merchantModel.MerchantAuth{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "FAILED: error when get snap private key",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Return(&redis.StringCmd{})
				merchantSvc.On("GetSnapPrivateKey",
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("error"))
			},
			wantErr: true,
		},
		{
			name:    "FAILED: error send callback when call callback access token",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Return(&redis.StringCmd{})
				privateKey, _ := generateMockPrivateKey()
				merchantSvc.On("GetSnapPrivateKey",
					mock.Anything,
					mock.Anything,
				).Return(privateKey, nil)
				callbackRepo.On("CreateCallbackLog", mock.Anything, mock.Anything).Return(nil)
				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("", fmt.Errorf("error")).Once()
				callbackRepo.On("UpdateCallbackLog", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name:    "FAILED: error when create callback log",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				strCmd := redis.NewStringCmd(context.Background(), "token")
				strCmd.SetVal("token")
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd)
				strCmd2 := redis.NewStringCmd(context.Background(), "tokenBody")
				strCmd2.SetVal(cacheAccessTokenBody)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd2)
				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("", errors.New("error"))
				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					mock.Anything,
				).Return(nil)
				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					mock.Anything,
				).Return(nil)
				merchantRepo.On("GetMerchantSnapPKCS8KeyByMerchantID",
					mock.Anything,
					constant.StringMockType(),
				).Return(&merchantModel.MerchantAuth{}, nil)
			},
			wantErr: true,
		},
		{
			name:    "FAILED: error when create callback log and update callback log",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				strCmd := redis.NewStringCmd(context.Background(), "token")
				strCmd.SetVal("token")
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd)
				strCmd2 := redis.NewStringCmd(context.Background(), "tokenBody")
				strCmd2.SetVal(cacheAccessTokenBody)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd2)
				callbackPartnerSvc.On("Callback",
					mock.Anything,
					callbackReqMockType,
					constant.MapStrValStringMockType(),
				).Return("", errors.New("error"))
				callbackRepo.On("CreateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType(),
				).Return(nil)
				callbackRepo.On("UpdateCallbackLog",
					mock.Anything,
					constant.PtrCallbackLogMockType(),
				).Return(errors.New("error"))
				merchantRepo.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, constant.StringMockType(),
				).Return(&merchantModel.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(&vault.DecryptResponse{}, nil)
			},
			wantErr: true,
		},
		{
			name:    "FAILED: error when get snap key from table",
			request: validRequest,
			mocksSetup: func(
				request *callbackModel.ProcessCallbackRequest,
				cacheMock *redisExtMocks.IRedisExt,

				callbackRepo *repositoryMocks.ICallbackRepository,
				callbackPartnerSvc *callbackPartnerMocks.ICallbackPartner,
				merchantSvc *merchantSvcMocks.IMerchantService,
				merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)
				strCmd := redis.NewStringCmd(context.Background(), "token")
				strCmd.SetVal("token")
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd)
				strCmd2 := redis.NewStringCmd(context.Background(), "tokenBody")
				strCmd2.SetVal(cacheAccessTokenBody)
				cacheMock.On("Get",
					mock.Anything,
					mock.Anything,
				).Once().Return(strCmd2)
				merchantRepo.On("GetMerchantSnapPKCS8KeyByMerchantID",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, errors.New("error when get secret"))
			},
			wantErr: true,
		},
		{
			name:    "FAILED: Decrypt merchant secrets",
			request: validRequest,
			mocksSetup: func(request *callbackModel.ProcessCallbackRequest, cacheMock *redisExtMocks.IRedisExt, callbackRepo *repositoryMocks.ICallbackRepository, _ *callbackPartnerMocks.ICallbackPartner, _ *merchantSvcMocks.IMerchantService, merchantRepo *repositoryMocks.IMerchantRepository) {
				callbackRepo.On("FindCallbackByNameAndMerchantID",
					mock.Anything, mock.Anything, mock.Anything,
				).Return(&callbackModel.Callback{
					UUID:             uuid.New(),
					CallbackMasterID: uuid.New(),
					MerchantID:       request.MerchantID,
					BaseURL:          &baseUrl,
					URL:              baseUrl + callbackPath,
					Description:      description,
				}, nil)

				strCmd := redis.NewStringCmd(context.Background(), "token")
				strCmd.SetVal("token")
				cacheMock.On("Get", mock.Anything, mock.Anything).Once().Return(strCmd)

				strCmd2 := redis.NewStringCmd(context.Background(), "tokenBody")
				strCmd2.SetVal(cacheAccessTokenBody)
				cacheMock.On("Get", mock.Anything, mock.Anything).Once().Return(strCmd2)

				merchantRepo.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, constant.StringMockType(),
				).Return(&merchantModel.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
				}, nil)
				vaultTransit.On(
					"Decrypt", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
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
			merchantSvc := merchantSvcMocks.NewIMerchantService(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			tt.mocksSetup(tt.request, mockCache, callbackRepoMock, callbackPartnerSvc, merchantSvc, merchantRepo)

			callbackSvc := New(mockLogger, mockCache, callbackRepoMock, callbackPartnerSvc, merchantSvc, WithMerchantRepository(merchantRepo), WithVaultTransit(vaultTransit))

			err := callbackSvc.ProcessCallback(t.Context(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			vaultTransit.AssertExpectations(t)
			callbackRepoMock.AssertExpectations(t)
		})
	}
}

func generateMockPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
