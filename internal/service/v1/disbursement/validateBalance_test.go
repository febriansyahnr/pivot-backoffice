package disbursementService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestGetAvailableMerchantBalance(t *testing.T) {
	validInput := &disbursementModel.ValidateBalanceRequest{
		DisbursementIDs: []string{uuid.NewString(), uuid.NewString()},
		MerchantID:      uuid.NewString(),
	}
	conf := config.Config{
		Environment: c.EnvironmentStaging,
	}

	ctxOnBehalf := func() context.Context {
		return context.WithValue(context.Background(), c.CtxParentMerchantId, "123456789")
	}

	testCases := []struct {
		name       string
		context    func() context.Context
		mocksSetup func(disbursementRepo *repositoryMocks.IDisbursementRepository,
			orchestratorSvc *serviceMocks.IOrchestratorService,
		)
		input *disbursementModel.ValidateBalanceRequest
		valid bool
	}{
		{
			name: "VALID:Check balance valid",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(100_000.00, nil)

				disbursementRepo.On(
					"SumAmountByIDs", c.ValueCtxMockType(), c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 10_000}, nil)
			},
			input: validInput,
			valid: true,
		},
		{
			name: "INVALID:Check balance invalid due to GetAvailableMerchantBalance error",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(0.00, c.ErrSomeErrorForUnitTest)
			},
			input: validInput,
			valid: false,
		},
		{
			name: "INVALID:Check balance invalid due to SumAmountByIDs error",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(100_000.00, nil)

				disbursementRepo.On(
					"SumAmountByIDs", c.ValueCtxMockType(), c.ArrayStringMockType(),
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: validInput,
			valid: false,
		},
		{
			name: "INVALID:Merchant balance is not enough",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(0.00, nil)

				disbursementRepo.On(
					"SumAmountByIDs", c.ValueCtxMockType(), c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 10_000}, nil)
			},
			input: validInput,
			valid: false,
		},
		{
			name:    "INVALID:Transaction On-Behalf/Main Merchant Balance Error",
			context: ctxOnBehalf,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Times(1).Return(100_000.00, nil)
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Times(1).Return(00.00, c.ErrSomeErrorForUnitTest)

				disbursementRepo.On("SumAmountByIDs", c.ValueCtxMockType(), c.ArrayStringMockType()).Return(&disbursementModel.SumAmountResponse{TotalAmount: 10_000}, nil)
			},
			input: validInput,
			valid: false,
		},
		{
			name:    "INVALID:Transaction On-Behalf/Main Merchant Balance Insufficient",
			context: ctxOnBehalf,
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				disbursementRepo.On(
					"SumAmountByIDs", c.ValueCtxMockType(), c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{ParentFeeCharged: 2_750.00}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Times(1).Return(12_500.00, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Times(1).Return(500.00, nil)
			},
			input: validInput,
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			snapCoreRepoMock := repositoryMocks.NewISnapCoreRepository(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			beneficiaryAccSvcMock := serviceMocks.NewIBeneficiaryAccountService(t)

			tc.mocksSetup(disbursementRepoMock, orchSvcMock)

			svc := New(
				&conf, pdkLoggerMock, merchantRepo, disbursementRepoMock, snapCoreRepoMock, nil,
				WithOrchestratorService(orchSvcMock), WithBeneficiaryAccService(beneficiaryAccSvcMock),
			)
			ctx := context.Background()
			if tc.context != nil {
				ctx = tc.context()
			}
			valid := svc.ValidateBalance(ctx, tc.input)
			assert.Equal(t, tc.valid, valid)

			disbursementRepoMock.AssertExpectations(t)
			snapCoreRepoMock.AssertExpectations(t)
			orchSvcMock.AssertExpectations(t)
			beneficiaryAccSvcMock.AssertExpectations(t)
		})
	}
}
