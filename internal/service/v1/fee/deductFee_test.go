package feeService_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeductBalanceForIndirectFeeType(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)

	service := New(
		logger, nil, merchantRepo,
		WithOrchestratorService(orchestratorSvc),
		WithAccountTransactionRepository(accountTransactionRepo),
	)

	merchantId := uuid.NewString()
	reference, method := c.TypeDisbursement, ""
	startDate := time.Date(2024, 1, 1, 2, 9, 46, 0, time.UTC).Add(time.Second)
	endDate := time.Date(2024, 2, 28, 23, 59, 59, 0, tz).UTC()

	accumulateTrxFees := &orchestrator_model.AccumulateTransactionFees{
		AccountName:    c.TypeDisbursement,
		TotalRows:      1,
		TotalFees:      2_500,
		TotalTaxes:     250,
		TransactionIds: []string{"123456"},
	}

	tests := []struct {
		name      string
		date      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Get merchant fee list",
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeListForBalanceDeduction", c.ValueCtxMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("get merchant list: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeListForBalanceDeduction", c.ValueCtxMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name: "SUCCESS:There is no balance deduction schedule",
			date: "2024-02-26 00:15:39",
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeListForBalanceDeduction", c.ValueCtxMockType(),
				).Return([]merchant.MerchantFeeForBalanceDeduction{
					{
						MerchantId:     merchantId,
						Reference:      reference,
						Method:         method,
						DeductionDay:   31,
						LastDeductDate: nil,
						CreatedAt:      time.Date(2024, 1, 1, 2, 9, 46, 0, time.UTC),
					},
				}, nil)
			},
		},
		{
			name: "ERROR:Accumulate transaction fees",
			date: "2024-02-29 03:15:39",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetAccumulateTransactionFees", c.ValueCtxMockType(), merchantId, reference, method, startDate, endDate,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("get accumulate transaction fees: %v", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS:No transactions",
			date: "2024-02-29 02:30:40",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetAccumulateTransactionFees", c.ValueCtxMockType(), merchantId, reference, method, startDate, endDate,
				).Once().Return(&orchestrator_model.AccumulateTransactionFees{}, nil)
				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, reference, endDate,
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:Get available merchant balance",
			date: "2024-02-29 02:30:40",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetAccumulateTransactionFees", c.ValueCtxMockType(), merchantId, reference, method, startDate, endDate,
				).Return(accumulateTrxFees, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, accumulateTrxFees.AccountName,
				).Once().Return(0.00, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Insufficient balance",
			date: "2024-02-29 02:30:40",
			setupMock: func() {

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, accumulateTrxFees.AccountName,
				).Once().Return(2_000.00, nil)
			},
		},
		{
			name: "ERROR:Deduct balance for indirect fee type",
			date: "2024-02-29 01:45:16",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, accumulateTrxFees.AccountName,
				).Return(2_500.00, nil)

				accountTransactionRepo.On(
					"DeductBalanceForIndirectFeeType", c.ValueCtxMockType(), merchantId, accumulateTrxFees.TransactionIds,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Update deduction last date",
			date: "2024-02-29 04:45:16",
			setupMock: func() {
				accountTransactionRepo.On(
					"DeductBalanceForIndirectFeeType", c.ValueCtxMockType(), merchantId, accumulateTrxFees.TransactionIds,
				).Return(nil)

				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, reference, endDate,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("update last deduct date: %v", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS:Deduct balance",
			date: "2024-02-29 05:39:40",
			setupMock: func() {
				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, reference, endDate,
				).Return(nil)

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			date, err := time.Time{}, error(nil)
			if test.date != "" {
				date, err = time.ParseInLocation(time.DateTime, test.date, tz)
			}
			require.NoError(t, err)

			assert.Equal(t, test.wantErr, service.DeductBalanceForIndirectFeeType(context.Background(), date))
		})
	}
}
