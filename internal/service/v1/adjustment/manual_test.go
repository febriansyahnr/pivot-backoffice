package adjustment_test

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/adjustment"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	gcsPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateManualTopup(t *testing.T) {
	pdkLog := pdkLoggerMock.NewILogger(t)
	gcsMock := gcsPkgMock.NewGCSService(t)
	rabbitMock := rabbitPkgMock.NewRabbitMQExt(t)
	adjustRepoMock := repoMocks.NewIAdjustmentRepository(t)
	merchantRepoMock := repoMocks.NewIMerchantRepository(t)
	orchestraMock := serviceMocks.NewIOrchestratorService(t)
	merchantTopUpSvc := serviceMocks.NewIMerchantTopUpService(t)

	rabbitMock.On(
		"Publish", c.ValueCtxMockType(), c.StringMockType(), c.PtrStringMockType(), mock.Anything,
	).Return(nil)

	service := New(config.SlackConfig{}, adjustRepoMock, merchantRepoMock)
	WithLogger(service, pdkLog)
	WithGCSService(service, gcsMock)
	WithRabbitMQ(service, rabbitMock)
	WithOrchestratorService(service, orchestraMock)
	WithMerchantTopUpCallbackService(service, merchantTopUpSvc)

	merchantID := "4ae4e1d2-b483-4d9b-b590-8ddc125a230b"
	merchantName := "Test Merchant" // NOSONAR
	data := &adjustment.ManualTopupRequest{
		File:       &multipart.FileHeader{},
		MerchantID: merchantID,
	}

	tests := []struct {
		name         string
		data         *adjustment.ManualTopupRequest
		modifierMock func()
		wantErr      string
	}{
		{
			name: "ERROR:Find merchant by ID/Invalid session",
			modifierMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantID,
				).Once().Return(nil, errors.New("Invalid session"))
			},
			wantErr: "Invalid session",
		},
		{
			name: "ERROR:Find merchant by ID/Not found",
			modifierMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantID,
				).Once().Return(nil, nil)
			},
			wantErr: "merchant id not found",
		},
		{
			name: "ERROR:Get available merchant balance",
			modifierMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.Merchant{UUID: merchantID, Name: merchantName}, nil)
				orchestraMock.On(
					"GetAvailableMerchantBalance", mock.Anything, merchantID, c.AccountNameDisbursement,
				).Once().Return(0.0, assert.AnError)
				pdkLog.On(
					"Error", mock.Anything, "Failed while getting available merchant balance", mock.Anything,
				).Once().Return()
			},
			wantErr: "ERROR_DATABASE | an error occurred on the server. please try again later",
		},
		{
			name: "ERROR:Upload proof of transfer",
			modifierMock: func() {
				orchestraMock.On(
					"GetAvailableMerchantBalance", mock.Anything, merchantID, c.AccountNameDisbursement,
				).Return(10_000.0, nil)
				gcsMock.On(
					"UploadProofOfTransfer", c.ValueCtxMockType(), c.FileHeaderMockType(), c.BoolMockType(),
				).Once().Return("", errors.New("invalid credential"))
			},
			wantErr: "invalid credential",
		},
		{
			name: "ERROR:Begin transaction",
			modifierMock: func() {
				gcsMock.On(
					"UploadProofOfTransfer", c.ValueCtxMockType(), c.FileHeaderMockType(), c.BoolMockType(),
				).Return("https://example.id/storage", nil)

				adjustRepoMock.On(
					"BeginTransaction", c.ValueCtxMockType(),
				).Once().Return(nil, errors.New("begin transaction: invalid session"))
			},
			wantErr: "begin transaction: invalid session",
		},
		{
			name: "ERROR:Rollback transaction",
			modifierMock: func() {
				adjustRepoMock.On(
					"BeginTransaction", c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				adjustRepoMock.On(
					"CreateAdjustment", c.ValueCtxMockType(), manualAdjHistoryMockType,
				).Once().Return(errors.New("create manual topup: invalid session"))
				adjustRepoMock.On(
					"RollbackTransaction", c.ValueCtxMockType(),
				).Once().Return(errors.New("rollback transaction: invalid session"))
			},
			wantErr: "rollback transaction: invalid session",
		},
		{
			name: "ERROR:Post account transaction",
			modifierMock: func() {
				adjustRepoMock.On(
					"RollbackTransaction", c.ValueCtxMockType(),
				).Return(nil)
				adjustRepoMock.On(
					"CreateAdjustment", c.ValueCtxMockType(), manualAdjHistoryMockType,
				).Return(nil)
				orchestraMock.On(
					"PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(errors.New("post account transaction: invalid session"))
			},
			wantErr: "post account transaction: invalid session",
		},
		{
			name: "ERROR:Commit transaction",
			modifierMock: func() {
				orchestraMock.On(
					"PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", c.ValueCtxMockType(),
				).Once().Return(errors.New("commit transaction: invalid session"))
			},
			wantErr: "commit transaction: invalid session",
		},
		{
			name: "SUCCESS",
			modifierMock: func() {
				adjustRepoMock.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)
			},
		},
		{
			name: "SUCCESS: With warning publish merchant callback",
			data: &adjustment.ManualTopupRequest{
				MerchantID:   merchantID,
				File:         &multipart.FileHeader{},
				SendCallback: true,
			},
			modifierMock: func() {
				merchantTopUpSvc.On(
					"SendCallback", mock.Anything, c.CallbackEventMerchantTopUpSuccess, mock.Anything,
				).Once().Return(assert.AnError)
				pdkLog.On(
					"Warn", mock.Anything, "Failed to publish merchant callback delivery message", mock.Anything, mock.Anything,
				).Once().Return()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			if test.data == nil {
				test.data = data
			}
			id, err := service.CreateManualTopup(context.Background(), test.data)
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.NoError(t, uuid.Validate(id))

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
