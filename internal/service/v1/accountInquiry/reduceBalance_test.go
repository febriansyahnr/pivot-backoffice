package accountinquiry

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mocks_logger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReduceBalance(t *testing.T) {

	onBehalfMetadata := &requestAccountInquiries.Metadata{
		FeeOnBehalf: &feeModel.TrxFeeOnBehalfMetadata{
			FinalAmount: 1_000,
		},
		OnBehalf: &merchant.OnBehalfObject{
			ParentMerchantId: "4739703d-b609-4b18-aa8f-80818813210b",
		},
	}

	testCases := []struct {
		Name      string
		Metadata  *requestAccountInquiries.Metadata
		WantErr   bool
		MockSetup func(
			accTrx *mocks.IOrchestratorService,

			trfSvc *mocks.ITransferService,
		)
	}{
		{
			Name:     "Success account inquiry fee (on-behalf)",
			Metadata: onBehalfMetadata,
			WantErr:  false,
			MockSetup: func(accTrx *mocks.IOrchestratorService, trfSvc *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(1000.0, nil)
				trfSvc.On("Transfer", constant.ValueCtxMockType(), mock.Anything).Return(&transfer.Transfer{UUID: uuid.New()}, nil)
				accTrx.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:    "Success reduce balance",
			WantErr: false,
			MockSetup: func(accTrx *mocks.IOrchestratorService, _ *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(1000.0, nil)
				accTrx.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:    "Error get available balance",
			WantErr: true,
			MockSetup: func(accTrx *mocks.IOrchestratorService, _ *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(0.0, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name:    "Error insufficient balance",
			WantErr: true,
			MockSetup: func(accTrx *mocks.IOrchestratorService, _ *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(1.0, nil)
			},
		},
		{
			Name:    "Error create transaction",
			WantErr: true,
			MockSetup: func(accTrx *mocks.IOrchestratorService, _ *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(1000.0, nil)
				accTrx.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name:     "Error account inquiry fee (on-behalf)",
			Metadata: onBehalfMetadata,
			WantErr:  true,
			MockSetup: func(accTrx *mocks.IOrchestratorService, trfSvc *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(1000.0, nil)
				trfSvc.On(
					"Transfer", constant.ValueCtxMockType(), mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name: "Success reduce balance with automated deduction type",
			Metadata: &requestAccountInquiries.Metadata{
				FeeDetail: &feeModel.FeeMetadataObject{
					FinalAmount:   1_000,
					DeductionType: constant.MerchantFeeDeductionTypeAutomated,
				},
			},
			WantErr: false,
			MockSetup: func(accTrx *mocks.IOrchestratorService, _ *mocks.ITransferService) {
				accTrx.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(1000.0, nil)
				accTrx.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			logger, _ := mocks_logger.NewZapLogger(mocks_logger.Config{})
			accountTransactionSvc := mocks.NewIOrchestratorService(t)
			trfSvc := mocks.NewITransferService(t)

			test.MockSetup(accountTransactionSvc, trfSvc)
			svc := &AccountInquiryService{
				logger:              logger,
				orchestratorService: accountTransactionSvc,
				transferSvc:         trfSvc,
			}
			if test.Metadata == nil {
				test.Metadata = &requestAccountInquiries.Metadata{}
			}
			if test.Metadata.FeeDetail == nil {
				test.Metadata.FeeDetail = &feeModel.FeeMetadataObject{
					FinalAmount: 1_000,
				}
			}
			err := svc.ReduceBalance(context.Background(), uuid.NewString(), uuid.NewString(), test.Metadata)

			if test.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
