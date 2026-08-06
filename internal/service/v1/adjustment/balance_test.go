package adjustment_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/adjustment"
	gcsPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	normalAction     = "NORMAL"
	adjustmentAction = "ADJUSTMENT"
)

func TestCreateBalanceAdjustmentFromManualTopUp(t *testing.T) {
	gcsMock := gcsPkgMock.NewGCSService(t)
	rabbitMock := rabbitPkgMock.NewRabbitMQExt(t)
	adjustRepoMock := repoMocks.NewIAdjustmentRepository(t)
	merchantRepoMock := repoMocks.NewIMerchantRepository(t)
	orchestraMock := serviceMocks.NewIOrchestratorService(t)

	rabbitMock.On(
		"Publish", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), constant.PtrStringMockType(), mock.Anything,
	).Return(nil)

	service := New(config.SlackConfig{}, adjustRepoMock, merchantRepoMock)
	WithGCSService(service, gcsMock)
	WithRabbitMQ(service, rabbitMock)
	WithOrchestratorService(service, orchestraMock)

	data := &adjustment.BalanceAdjustmentRequest{
		MerchantID:   uuid.NewString(),
		Currency:     "IDR",
		Amount:       -1000000,
		CreatedBy:    "John Doe",
		Notes:        "Adjust previous manual top up",
		AdjustmentID: uuid.NewString(),
	}

	tests := []struct {
		name         string
		data         *adjustment.BalanceAdjustmentRequest
		modifierMock func()
		wantErr      string
	}{
		{
			name: "ERROR: Find merchant by ID/Invalid session",
			modifierMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Find merchant by ID/Not found",
			modifierMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
			wantErr: "merchant not found",
		},
		{
			name: "ERROR: Related Adjustment error",
			modifierMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&merchantModel.Merchant{UUID: data.MerchantID}, nil)

				adjustRepoMock.On(
					"FindByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Related Adjustment not found",
			modifierMock: func() {
				adjustRepoMock.On(
					"FindByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
			wantErr: "related adjustment ID not found",
		},
		{
			name: "ERROR: Related Adjustment not top up mode",
			modifierMock: func() {
				adjustRepoMock.On(
					"FindByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(&adjustment.ManualAdjustmentHistory{Action: adjustmentAction}, nil)
			},
			wantErr: "can't perform to this related adjustment ID",
		},
		{
			name: "ERROR: Merchant is not valid",
			modifierMock: func() {
				adjustRepoMock.On(
					"FindByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(&adjustment.ManualAdjustmentHistory{Action: normalAction, MerchantID: "invalid"}, nil)
			},
			wantErr: "merchant is not valid",
		},
		{
			name: "ERROR: CalculateAmountBalanceAdjustmentForTopUp service",
			modifierMock: func() {
				bankAccount, _ := json.Marshal(adjustment.BankAccount{
					Name:      "Mandiri",
					AccNumber: "1234",
				})
				adjustRepoMock.On(
					"FindByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&adjustment.ManualAdjustmentHistory{Action: normalAction, MerchantID: data.MerchantID, BankAccount: string(bankAccount)}, nil)

				adjustRepoMock.On(
					"CalculateAmountBalanceAdjustmentForTopUp", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(0.0, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Deduct amount exceed top up amount",
			modifierMock: func() {
				adjustRepoMock.On(
					"CalculateAmountBalanceAdjustmentForTopUp", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(0.0, nil)
			},
			wantErr: "amount deduction exceeds topup amount",
		},
		{
			name: "ERROR: Begin transaction",
			modifierMock: func() {
				adjustRepoMock.On(
					"CalculateAmountBalanceAdjustmentForTopUp", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(3000000.0, nil)

				adjustRepoMock.On(
					"BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Rollback transaction",
			modifierMock: func() {
				adjustRepoMock.On(
					"BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.AnythingOfType(constant.MockTypeValueContextReference), manualAdjHistoryMockType,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				adjustRepoMock.On(
					"RollbackTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Post account transaction",
			modifierMock: func() {
				adjustRepoMock.On(
					"RollbackTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.AnythingOfType(constant.MockTypeValueContextReference), manualAdjHistoryMockType,
				).Return(nil)
				orchestraMock.On(
					"PostAccountTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*orchestrator_model.CreateAccountTransactionRequest"),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Commit transaction",
			modifierMock: func() {
				orchestraMock.On(
					"PostAccountTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*orchestrator_model.CreateAccountTransactionRequest"),
				).Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			modifierMock: func() {
				adjustRepoMock.On(
					"CommitTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(nil)
			},
			wantErr: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			if test.data == nil {
				test.data = data
			}
			id, err := service.CreateBalanceAdjustmentFromManualTopUp(context.Background(), test.data)
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
