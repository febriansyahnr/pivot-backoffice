package callbackService_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	callbackMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/callback"
	rabbitMqExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCallbackURLByMerchantId(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	rmq := rabbitMqExt.NewRabbitMQExt(t)
	rmq.On(
		"PublishActivity", c.ValueCtxMockType(),
		c.PtrStringMockType(), c.PtrStringMockType(), c.StringMockType(), c.StringMockType(), c.MapStrValStringMockType(),
	).Return(nil)

	repo := repoMocks.NewICallbackRepository(t)
	service := New(logger, nil, repo, nil, nil, WithRabbitMQExt(rmq))

	traceID := uuid.NewString()
	data := []callbackModel.CallbackURLSettingResp{
		{
			MasterID:    "19007642-1980-4ffe-8add-462a369fa775",
			MasterName:  "PAYMENT",
			CallbackID:  c.NullString{NullString: sql.NullString{String: "2dd45926-3ae4-4b8d-bf46-26af612dc228", Valid: true}},
			CallbackURL: c.NullString{NullString: sql.NullString{String: "https://example.id/callback", Valid: true}},
		},
	}
	request := &callbackModel.CallbackURLSettingReq{Info: &http.Request{}}
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
		wantResult []callbackModel.CallbackURLSettingResp
	}{
		{
			name: "ERROR:Some internal error",
			setupMocks: func() {
				repo.On(
					"GetCallbackURLByMerchantId", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceID),
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				repo.On(
					"GetCallbackURLByMerchantId", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(data, nil)
			},
			wantResult: data,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if resp, err := service.GetCallbackURLByMerchantId(ctx, request); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetCallbackAPIKeyByMerchantId(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	rmq := rabbitMqExt.NewRabbitMQExt(t)
	rmq.On(
		"PublishActivity", c.ValueCtxMockType(),
		c.PtrStringMockType(), c.PtrStringMockType(), c.StringMockType(), c.StringMockType(), c.MapStrValStringMockType(),
	).Return(nil)

	userSvc := serviceMocks.NewIUserService(t)
	repo := repoMocks.NewICallbackRepository(t)
	vaultTransit := vaultMock.NewIVaultTransit(t)

	service := New(logger, nil, repo, nil, nil, WithRabbitMQExt(rmq), WithUserService(userSvc), WithVaultTransit(vaultTransit))

	request := &callbackModel.CallbackURLSettingReq{Info: &http.Request{}}

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
		wantResult *callbackModel.CallbackAPIKeyResp
	}{
		{
			name: "ERROR:Some internal error",
			setupMocks: func() {
				userSvc.On(
					"CheckCurrentPin", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Invalid PIN",
			setupMocks: func() {
				userSvc.On(
					"CheckCurrentPin", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrInvalidPIN)
			},
			wantErr: "invalid pin",
		},
		{
			name: "ERROR:Get callback api key by merchant id",
			setupMocks: func() {
				userSvc.On(
					"CheckCurrentPin", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)

				repo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "an error occurred on the server. please try again later",
		},
		{
			name: "ERROR:Empty callback API key",
			setupMocks: func() {
				repo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&callbackModel.CallbackAPIKeyResp{}, nil)
			},
			wantErr: "empty callback api key",
		},
		{
			name: "ERROR:Decrypt callback API key",
			setupMocks: func() {
				repo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&callbackModel.CallbackAPIKeyResp{
					APIKey: "vault:v1:....", Version: 1, // NOSONAR
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: "an error occurred on the server. please try again later",
		},
		{
			name: "SUCCESS:Plaintext",
			setupMocks: func() {
				repo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&callbackModel.CallbackAPIKeyResp{APIKey: "this-is-api-key"}, nil) // NOSONAR
			},
			wantResult: &callbackModel.CallbackAPIKeyResp{APIKey: "this-is-api-key"}, // NOSONAR
		},
		{
			name: "SUCCESS:Decrypt",
			setupMocks: func() {
				repo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&callbackModel.CallbackAPIKeyResp{
					APIKey: "vault:v1:....", Version: 1, // NOSONAR
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(&vault.DecryptResponse{
					Plaintext: []byte("abc123"), // NOSONAR
				}, nil)
			},
			wantResult: &callbackModel.CallbackAPIKeyResp{APIKey: "abc123"}, // NOSONAR
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if resp, err := service.GetCallbackAPIKeyByMerchantId(t.Context(), request); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestTestAndSaveCallbackURL(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	rmq := rabbitMqExt.NewRabbitMQExt(t)
	rmq.On(
		"PublishActivity", c.ValueCtxMockType(),
		c.PtrStringMockType(), c.PtrStringMockType(), c.StringMockType(), c.StringMockType(), c.MapStrValStringMockType(),
	).Return(nil)

	callbackRepo := repoMocks.NewICallbackRepository(t)
	callbackPartnerSvc := callbackMock.NewICallbackPartner(t)

	traceID := uuid.NewString()
	secretKey := "SECRET-API-CALLBACK"
	request := &callbackModel.TestAndSaveCallbackURLReq{Info: &http.Request{}}
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)
	service := New(logger, nil, callbackRepo, callbackPartnerSvc, nil, WithRabbitMQExt(rmq))

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
		wantResult *callbackModel.TestAndSaveCallbackURLResp
	}{
		{
			name: "ERROR:Get callback api key by merchant id",
			setupMocks: func() {
				callbackRepo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "an error occurred on the server. please try again later",
		},
		{
			name: "ERROR:Empty callback api key",
			setupMocks: func() {
				callbackRepo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&callbackModel.CallbackAPIKeyResp{}, nil)
			},
			wantErr: "empty callback api key",
		},
		{
			name: "ERROR:Request to callback URL",
			setupMocks: func() {
				callbackRepo.On(
					"GetCallbackAPIKeyByMerchantId", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&callbackModel.CallbackAPIKeyResp{APIKey: secretKey}, nil)

				callbackPartnerSvc.On(
					"Callback", c.ValueCtxMockType(), callbackReqMockType, c.MapStrValStringMockType(),
				).Once().Return("", fmt.Errorf("callback partner: %w", c.ErrSomeErrorForUnitTest))
			},
			wantErr: "callback partner: some error",
		},
		{
			name: "SUCCESS:Request to callback URL (invalid body)",
			setupMocks: func() {
				callbackPartnerSvc.On(
					"Callback", c.ValueCtxMockType(), callbackReqMockType, c.MapStrValStringMockType(),
				).Once().Return(`{"code":"40","errors":"invalid request body"}`, errors.New("invalid request body"))
			},
			wantResult: &callbackModel.TestAndSaveCallbackURLResp{
				Status:    false,
				RequestID: traceID,
				Body: map[string]string{
					"code":   "40",
					"errors": "invalid request body",
				},
			},
		},
		{
			name: "ERROR:Begin transaction",
			setupMocks: func() {
				callbackPartnerSvc.On(
					"Callback", c.ValueCtxMockType(), callbackReqMockType, c.MapStrValStringMockType(),
				).Return(`{"code":"00","body":{"message":"OK"}}`, nil)

				callbackRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("TX: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "ERROR:Rollback transaction",
			setupMocks: func() {
				callbackRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.Background(), nil)

				callbackRepo.On(
					"GetCallbackIdByMerchantAndMasterCallbackId", c.BackgroundCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
				callbackRepo.On("RollbackTransaction", c.BackgroundCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("RT: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "ERROR:Get callback id by merchant and callback master id",
			setupMocks: func() {
				callbackRepo.On("RollbackTransaction", c.BackgroundCtxMockType()).Return(nil)

				callbackRepo.On(
					"GetCallbackIdByMerchantAndMasterCallbackId", c.BackgroundCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("GET: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "ERROR:Update callback URL",
			setupMocks: func() {
				callbackRepo.On(
					"GetCallbackIdByMerchantAndMasterCallbackId", c.BackgroundCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(uuid.NewString(), nil)
				callbackRepo.On(
					"UpdateCallbackURLById", c.BackgroundCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("UPDATE: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "ERROR:Create callback",
			setupMocks: func() {
				callbackRepo.On(
					"GetCallbackIdByMerchantAndMasterCallbackId", c.BackgroundCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return("", nil)
				callbackRepo.On(
					"CreateCallback", c.BackgroundCtxMockType(), ptrCallbackMockType,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("CREATE: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "ERROR:Create callback log",
			setupMocks: func() {
				callbackRepo.On(
					"CreateCallback", c.BackgroundCtxMockType(), ptrCallbackMockType,
				).Return(nil)

				callbackRepo.On(
					"CreateCallbackLog", c.BackgroundCtxMockType(), c.PtrCallbackLogMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("LOG: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "ERROR:Commit transaction",
			setupMocks: func() {
				callbackRepo.On(
					"CreateCallbackLog", c.BackgroundCtxMockType(), c.PtrCallbackLogMockType(),
				).Return(nil)

				callbackRepo.On("CommitTransaction", c.BackgroundCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("CMT: "+c.InternalErrorFmt, traceID),
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				callbackRepo.On("CommitTransaction", c.BackgroundCtxMockType()).Return(nil)
			},
			wantResult: &callbackModel.TestAndSaveCallbackURLResp{
				Status:    true,
				RequestID: traceID,
				Body: map[string]interface{}{
					"code": "00",
					"body": map[string]string{
						"message": "OK",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if resp, err := service.TestAndSaveCallbackURL(ctx, request); test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)

			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult.Status, resp.Status)
				assert.Equal(t, test.wantResult.RequestID, resp.RequestID)
				assert.Equal(t, "***************BACK", resp.Information.CallbackToken)

				wantJson, _ := json.Marshal(test.wantResult.Body)
				actualJson, _ := json.Marshal(resp.Body)
				assert.JSONEq(t, string(wantJson), string(actualJson))

				if resp.Status {
					assert.NotEmpty(t, resp.Information.CallbackID)
					assert.NotEmpty(t, resp.Information.CallbackLogID)
				}
			}
		})
	}
}
