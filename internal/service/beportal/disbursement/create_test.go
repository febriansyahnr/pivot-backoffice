package disbursementService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	validRequest := &disbursementModel.CreateSingleRequest{}
	conf := config.Config{
		Environment: c.EnvironmentStaging,
	}

	mockRepo := repositoryMocks.NewIDisbursementRepository(t)
	mockStatusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	feeSvc := serviceMocks.NewIFeeService(t)
	beneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)
	feeSvc.On(
		"GetTransactionFeeOnBehalf", c.ValueCtxMockType(), mock.Anything,
	).Return(nil, nil)

	// General status history mock that will handle any calls
	mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

	ctx := context.WithValue(context.Background(), c.CtxParentMerchantId, uuid.NewString())

	feeDecimal := 1000.0
	merchant := &merchantModel.Merchant{}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func()
		input     *disbursementModel.CreateSingleRequest
	}{
		{
			name:      "ERROR: Merchant data not found",
			wantErr:   true,
			mockSetup: func() { /* Empty Function */ },
			input:     validRequest,
		},
		{
			name:  "ERROR: Find beneficiary account",
			input: validRequest,
			mockSetup: func() {
				ctx = context.WithValue(ctx, c.CtxMerchantData, merchant)

				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType,
				).Once().Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: beneficiary account inquiry error is deferred to processing",
			input: &disbursementModel.CreateSingleRequest{
				ReferenceID:         "sample-ref-id",
				BeneficiaryBankCode: "DANA",
				CreatedFrom:         c.DisbursementCreatedFromOpenApi,
			},
			wantErr: false,
			mockSetup: func() {
				ctx = context.WithValue(ctx, c.CtxMerchantData, merchant)

				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType,
				).Once().Return(nil, assert.AnError)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Once().Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)

				mockRepo.On("Insert", c.ValueCtxMockType(), DisbursementMockType).
					Once().Return(nil)
			},
		},
		{
			name:  "ERROR: Payout destination not eligible",
			input: validRequest,
			mockSetup: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType,
				).Once().Return(&beneficiaryAccountModel.Account{
					BeneficiaryBankCode:    "002",
					BeneficiaryAccountNo:   "123450000001",
					BeneficiaryAccountName: "TEST",
					MetadataObj: beneficiaryAccountModel.Metadata{
						IsVirtualAccount: true,
					},
				}, nil)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: GetFeeCalculationAndDetail",
			wantErr: true,
			mockSetup: func() {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType).
					Return(nil, nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Once().Return(feeDecimal, nil, c.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Insert",
			wantErr: true,
			mockSetup: func() {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType).
					Return(nil, nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)

				mockRepo.On("Insert", c.ValueCtxMockType(), DisbursementMockType).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func() {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType).
					Return(nil, nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)

				mockRepo.On("Insert", c.ValueCtxMockType(), DisbursementMockType).
					Return(nil)
			},
			input: validRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			svc := New(&conf, pdkLoggerMock, nil, mockRepo, nil, nil,
				WithStatusHistoriesRepository(mockStatusHistoriesRepo),
				WithFeeService(feeSvc),
				WithBeneficiaryAccService(beneficiaryAccountSvc),
			)
			response, err := svc.CreateSingle(ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}
		})
	}
}
