package xbPayoutService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	_ "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory" // imported for type checking in mock
)

func TestCreateSession(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
		XbCoreProcessorConfig: config.XbCoreProcessorConfig{
			BaseUSDRate: 15000.0,
		},
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository)
	}{
		{
			name:    "ERROR: XbCore CreateSender",
			wantErr: true,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: XbCore CreateBeneficiary",
			wantErr: true,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{
					UUID: uuid.New(),
				}, nil)

				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Duplicate Reference ID",
			wantErr: true,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				// Mock existing disbursement found - should return early without calling other services
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID: "existing-payout",
					},
				}, nil)
				// No other mocks needed since function should return early
			},
		},
		{
			name:    "ERROR: XbCore CreatePayoutSession",
			wantErr: true,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				// Mock CreateSender for empty senderId
				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{
					UUID: uuid.New(),
				}, nil)

				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{
					UUID: uuid.New(),
				}, nil)

				beneficiaryAccountRepo.On("Upsert",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(nil)

				xbCoreProcessorRepo.On("CreatePayoutSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreatePayoutSessionRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: FeeService GetFeeCalculationAndDetail",
			wantErr: true,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				// Mock CreateSender for empty senderId
				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{
					UUID: uuid.New(),
				}, nil)

				// Mock CreateBeneficiary for empty beneficiaryId
				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{
					UUID: uuid.New(),
				}, nil)

				beneficiaryAccountRepo.On("Upsert",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(nil)

				xbCoreProcessorRepo.On("CreatePayoutSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreatePayoutSessionRequest"),
				).Return(&xbCoreProcessorModel.CreatePayoutSessionResponseData{}, nil)

				feeSvc.On("GetFeeCalculationAndDetail",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*feeModel.GetFeeRequest"),
				).Once().Return(0.0, nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: DisbursementRepo Insert",
			wantErr: true,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				// Mock CreateSender for empty senderId
				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{
					UUID: uuid.New(),
				}, nil)

				// Mock CreateBeneficiary for empty beneficiaryId
				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{
					UUID: uuid.New(),
				}, nil)

				beneficiaryAccountRepo.On("Upsert",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(nil)

				xbCoreProcessorRepo.On("CreatePayoutSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreatePayoutSessionRequest"),
				).Return(&xbCoreProcessorModel.CreatePayoutSessionResponseData{}, nil)

				feeSvc.On("GetFeeCalculationAndDetail",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*feeModel.GetFeeRequest"),
				).Return(0.0, &feeModel.FeeMetadataObject{}, nil)

				xbCoreProcessorRepo.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetFxRateRequest"),
				).Return(&xbCoreProcessorModel.GetFxRateResponseData{
					MarkupFxRate: decimal.NewFromFloat(15000.0),
				}, nil)

				disbursementRepo.On("Insert",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementModel.Disbursement"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func(disbursementRepo *repositoryMock.IDisbursementRepository, beneficiaryAccountRepo *repositoryMock.IBeneficiaryAccountRepository, xbCoreProcessorRepo *repositoryMock.IXbCoreProcessorRepository, feeSvc *serviceMock.IFeeService, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				disbursementRepo.On("FindByMerchantAndReference",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				// Mock CreateSender for empty senderId
				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{
					UUID: uuid.New(),
				}, nil)

				// Mock CreateBeneficiary for empty beneficiaryId
				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{
					UUID: uuid.New(),
				}, nil)

				beneficiaryAccountRepo.On("Upsert",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(nil)

				xbCoreProcessorRepo.On("CreatePayoutSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreatePayoutSessionRequest"),
				).Return(&xbCoreProcessorModel.CreatePayoutSessionResponseData{}, nil)

				feeSvc.On("GetFeeCalculationAndDetail",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*feeModel.GetFeeRequest"),
				).Return(0.0, &feeModel.FeeMetadataObject{}, nil)

				xbCoreProcessorRepo.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetFxRateRequest"),
				).Return(&xbCoreProcessorModel.GetFxRateResponseData{
					MarkupFxRate: decimal.NewFromFloat(15000.0),
				}, nil)

				disbursementRepo.On("Insert",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementModel.Disbursement"),
				).Return(nil)

				statusHistoriesRepo.On("Insert",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh mocks for each test case
			disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
			beneficiaryAccountRepo := repositoryMock.NewIBeneficiaryAccountRepository(t)
			xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)
			feeSvc := serviceMock.NewIFeeService(t)
			statusHistoriesRepo := repositoryMock.NewIStatusHistoriesRepository(t)

			tc.setupMock(disbursementRepo, beneficiaryAccountRepo, xbCoreProcessorRepo, feeSvc, statusHistoriesRepo)

			svc := New(log, disbursementRepo, beneficiaryAccountRepo, xbCoreProcessorRepo, WithFeeService(feeSvc), WithConfig(cfg), WithStatusHistories(statusHistoriesRepo))
			_, err := svc.CreateSession(context.Background(), &xbModel.CreatePayoutSessionRequest{
				SenderData:      &xbModel.CreateSenderRequest{},
				BeneficiaryData: &xbModel.CreateBeneficiaryRequest{},
				CNAPS:           "989584027708",
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
