package paymentService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	logger "github.com/paper-indonesia/pdk/v2/logger"
)

func TestPostCreateFeeTransaction(t *testing.T) {
	log, _ := logger.NewZapLogger(logger.Config{})
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)
	feeSvc := serviceMock.NewIFeeService(t)
	paymentSvc := New(nil, log, nil, nil, nil, nil, nil,
		WithOrchestratorService(orchestratorSvc),
		WithFeeService(feeSvc),
	)

	testCases := []struct {
		name          string
		wantErr       bool
		inputPayment  *paymentModel.Payment
		setupMock     func()
		expectedError string
	}{
		{
			name:         "SUCCESS: Return on empty payment metadata",
			wantErr:      false,
			inputPayment: &paymentModel.Payment{},
		},
		{
			name:    "ERROR: PostAccountTransaction service",
			wantErr: true,
			inputPayment: &paymentModel.Payment{
				Metadata: &map[string]any{
					"feeDetail": map[string]interface{}{},
				},
			},
			setupMock: func() {
				feeSvc.On("CalculateFee",
					constant.ValueCtxMockType(), constant.PtrGetFeeRequestMockType(), constant.PtrFeeMetadataObjectMockType(),
				).Return(1_000.00, 0.00)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedError: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "SUCCESS: Create fee for DIRECT deduction",
			wantErr: false,
			inputPayment: &paymentModel.Payment{
				Metadata: &map[string]any{
					"feeDetail": feeModel.FeeMetadataObject{
						DeductionType: constant.MerchantFeeDeductionTypeDirect,
					},
				},
			},
			setupMock: func() {
				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Create fee for INDIRECT deduction",
			wantErr: false,
			inputPayment: &paymentModel.Payment{
				Metadata: &map[string]any{
					"feeDetail": feeModel.FeeMetadataObject{
						DeductionType: constant.MerchantFeeDeductionTypeAutomated,
					},
				},
			},
			setupMock: func() {
				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Create fee for indirect manual",
			wantErr: false,
			inputPayment: &paymentModel.Payment{
				Metadata: &map[string]any{
					"feeDetail": feeModel.FeeMetadataObject{
						DeductionType: constant.MerchantFeeDeductionTypeAutomated,
					},
				},
			},
			setupMock: func() {
				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			ctx := context.WithValue(context.Background(), constant.CtxParentMerchantId, uuid.NewString())
			err := paymentSvc.PostCreateFeeTransaction(ctx, tt.inputPayment, &paymentModel.PostCreateFeeTransactionRequest{
				SettlementTransactionMetadata: &settlementModel.AccountTransactionMetadataObject{},
				FeeTransactionID:              uuid.New(),
				LinkedTransactionID:           uuid.New(),
				Status:                        constant.StatusSuccess,
				Channel:                       paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Currency:                      "IDR",
				TransactionAmount:             1_000_000,
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
