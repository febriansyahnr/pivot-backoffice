package accountinquiry

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mocks_repo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	mocks_logger "github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRequestAccountInquiry(t *testing.T) {
	validPayload := requestAccountInquiries.RequestAccountInquiriesHttpRequest{
		MerchantID:  "test",
		ChannelCode: "MANDIRI",
		ChannelInformation: requestAccountInquiries.ChannelInformation{
			AccountName:   "test",
			AccountNumber: "test",
		},
	}
	validSnapInquiryResp := &routingProcessorModel.InquiryAccountResponseData{
		ResponseCode:           "200xx00",
		ResponseMessage:        "success",
		BeneficiaryAccountName: "test",
		BeneficiaryBankCode:    "008",
	}
	warningSnapInquiryResp := &routingProcessorModel.InquiryAccountResponseData{
		ResponseCode:           "200xx00",
		ResponseMessage:        "success",
		BeneficiaryAccountName: "test warning",
		BeneficiaryBankCode:    "008",
	}

	ctxValue := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, 1)
	ctxOnBehalf := context.WithValue(context.Background(), c.CtxParentMerchantId, "e9e0cb94-8a92-43e6-b8ff-2b4e56dd6215")
	showAdditionalInfo := true
	hideAdditionalInfo := false

	testCases := []struct {
		desc                   string
		ctx                    context.Context
		requestPayload         *requestAccountInquiries.RequestAccountInquiriesHttpRequest
		wantErr                bool
		expectAdditionalInfo   *bool
		expectedVirtualAccount bool
		mockSetup              func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository)
	}{
		{
			desc:    "error get transaction fee on behalf",
			ctx:     ctxOnBehalf,
			wantErr: true,
			mockSetup: func(_ *mocks.IRoutingProcessorService, _ *mocks_repo.IRequestAccountInquiryRepository, _ *mocks_repo.IAccountInquiriesRepository, _ *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetTransactionFeeOnBehalf", c.ValueCtxMockType(), mock.Anything).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "error get fee calculation and detail",
			wantErr: true,
			mockSetup: func(_ *mocks.IRoutingProcessorService, _ *mocks_repo.IRequestAccountInquiryRepository, _ *mocks_repo.IAccountInquiriesRepository, _ *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(0.0, nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "error get available merchant balance",
			wantErr: true,
			mockSetup: func(_ *mocks.IRoutingProcessorService, _ *mocks_repo.IRequestAccountInquiryRepository, _ *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(0.0, nil, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(0.0, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "error insufficient balance pre",
			wantErr: true,
			mockSetup: func(_ *mocks.IRoutingProcessorService, _ *mocks_repo.IRequestAccountInquiryRepository, _ *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(400.0, nil, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(200.0, nil)
			},
		},
		{
			desc:    "error_when_create_accountInquiryRepo",
			wantErr: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(c.ErrSomeErrorForUnitTest)

				// These mock setups should NOT be required because after the beneficiary.Create error,
				// the function should return an error before reaching the BeginTransaction call
				// No need to mock BeginTransaction, RollbackTransaction, etc.
			},
		},
		{
			desc:    "error_requestAccountInquiry_with_update_account_inquiries",
			wantErr: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On(
					"GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&beneficiaryModel.BeneficiaryAccount{
					UUID:                 "uuid-uuid-uuid",
					BeneficiaryAccountNo: "12345",
				}, nil)
				beneficiaryRepo.On("Update", mock.Anything, mock.Anything).Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "when_the_database_error_on_getting_beneficiary_data,_then_should_return_err",
			wantErr: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				// Standard setup for fee and balance checking
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				// Setup for routing processor
				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				// Set up expectation for database error on GetByBankCodeAndAccountNo
				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "when the database error on create beneficiary data, then should return err",
			wantErr: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "error insufficient balance post",
			wantErr: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{FinalAmount: 100.00}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Times(1).Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Times(1).Return(0.0, nil)
			},
		},
		{
			desc:    "error createRequestInquiries",
			wantErr: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "success create requestAccountInquiryRepo",
			wantErr: false,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			desc:    "success requestAccountInquiry with update account inquiries",
			wantErr: false,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On(
					"GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&beneficiaryModel.BeneficiaryAccount{
					UUID:                 "uuid-uuid-uuid",
					BeneficiaryAccountNo: "12345",
				}, nil)
				beneficiaryRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			desc:    "success requestAccountInquiry with status invalid",
			wantErr: false,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(nil, c.ErrSomeErrorForUnitTest)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			desc:    "success requestAccountInquiry with status warning",
			wantErr: false,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(warningSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			desc:    "success create requestAccountInquiryRepo (on-behalf)",
			ctx:     ctxOnBehalf,
			wantErr: false,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetTransactionFeeOnBehalf", c.ValueCtxMockType(), mock.Anything).Return(&feeModel.TrxFeeOnBehalfMetadata{FinalAmount: 100.00}, nil)
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{FinalAmount: 100.00}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(validSnapInquiryResp, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			desc: "success request account inquiry with additional info when feature flag enabled",
			requestPayload: &requestAccountInquiries.RequestAccountInquiriesHttpRequest{
				MerchantID:  "merchant-flag-on",
				ChannelCode: "MANDIRI",
				ChannelInformation: requestAccountInquiries.ChannelInformation{
					AccountName:   "test",
					AccountNumber: "1234567890",
				},
			},
			wantErr:                false,
			expectAdditionalInfo:   &showAdditionalInfo,
			expectedVirtualAccount: true,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Times(2).Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "200xx00",
					ResponseMessage:        "success",
					BeneficiaryAccountName: "test",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "BCA",
					IsVirtualAccount:       true,
				}, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			desc: "success request account inquiry without additional info when feature flag disabled",
			requestPayload: &requestAccountInquiries.RequestAccountInquiriesHttpRequest{
				MerchantID:  "merchant-flag-off",
				ChannelCode: "MANDIRI",
				ChannelInformation: requestAccountInquiries.ChannelInformation{
					AccountName:   "test",
					AccountNumber: "1234567890",
				},
			},
			wantErr:              false,
			expectAdditionalInfo: &hideAdditionalInfo,
			mockSetup: func(routingProcSvc *mocks.IRoutingProcessorService, reqAccountRepo *mocks_repo.IRequestAccountInquiryRepository, accountInquiryRepo *mocks_repo.IAccountInquiriesRepository, accountTransactionSvc *mocks.IOrchestratorService, feeSvc *mocks.IFeeService, beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(100.0, &feeModel.FeeMetadataObject{}, nil)
				accountTransactionSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).
					Times(2).Return(100.0, nil)

				routingProcSvc.On("AccountInquiry", mock.Anything, mock.Anything).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "200xx00",
					ResponseMessage:        "success",
					BeneficiaryAccountName: "test",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "BCA",
					IsVirtualAccount:       true,
				}, nil)

				beneficiaryRepo.On("GetByBankCodeAndAccountNo", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)
				beneficiaryRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				reqAccountRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxValue, nil)
				reqAccountRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)
				reqAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				accountTransactionSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			logger, _ := mocks_logger.NewZapLogger(mocks_logger.Config{})
			snapCore := mocks_repo.NewISnapCoreRepository(t)
			routingProcSvc := mocks.NewIRoutingProcessorService(t)
			reqAccountRepo := mocks_repo.NewIRequestAccountInquiryRepository(t)
			accountInquiryRepo := mocks_repo.NewIAccountInquiriesRepository(t)
			beneficiaryRepo := mocks_repo.NewIBeneficiaryAccountRepository(t)
			accountTransactionSvc := mocks.NewIOrchestratorService(t)
			feeSvc := mocks.NewIFeeService(t)
			trfSvc := mocks.NewITransferService(t)

			ffContentConfig := `
backend-portal-account-inquiry-display-virtual-account-flag-for-whitelisted-merchant:
  variations:
    ON: true
    OFF: false
  targeting:
    - name: allowed merchant
      query: merchant_id in ["test","merchant-flag-on"]
      variation: ON
  defaultRule:
    variation: OFF`
			f, err := os.CreateTemp(os.TempDir(), "account-inquiry-display-virtual-account-flag-for-whitelisted-merchant-*.yaml")
			require.NoError(t, err)
			defer func() { require.NoError(t, os.Remove(f.Name())) }()
			defer func() { require.NoError(t, f.Close()) }()

			_, err = f.WriteString(ffContentConfig)
			require.NoError(t, err)

			err = ffclient.Init(ffclient.Config{
				FileFormat: "YAML",
				Retriever: &fileretriever.Retriever{
					Path: f.Name(),
				},
			})
			require.NoError(t, err)
			defer ffclient.Close()

			tc.mockSetup(routingProcSvc, reqAccountRepo, accountInquiryRepo, accountTransactionSvc, feeSvc, beneficiaryRepo)

			// Set up additional mocks for specific test cases
			if tc.desc == "success create requestAccountInquiryRepo (on-behalf)" {
				// Mock the Transfer method for the on-behalf flow
				trfSvc.On("Transfer", mock.Anything, mock.Anything).Return(&transfer.Transfer{UUID: uuid.New()}, nil)
			}

			svc := New(
				logger, snapCore, reqAccountRepo, accountInquiryRepo, accountTransactionSvc, nil, feeSvc,
				WithBeneficiaryAccountRepository(beneficiaryRepo),
				WithRoutingProcessorService(routingProcSvc),
				WithConfig(&config.Config{Environment: "test"}),
				WithTransferService(trfSvc),
			)

			if tc.ctx == nil {
				tc.ctx = context.Background()
			}
			payload := validPayload
			if tc.requestPayload != nil {
				payload = *tc.requestPayload
			}
			if res, err := svc.RequestAccountInquiry(tc.ctx, payload); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				if tc.expectAdditionalInfo != nil {
					if *tc.expectAdditionalInfo {
						require.NotNil(t, res.AdditionalInfo)
						assert.Equal(t, tc.expectedVirtualAccount, res.AdditionalInfo.IsVirtualAccount)
					} else {
						assert.Nil(t, res.AdditionalInfo)
					}
				}
			}

		})
	}
}

func TestAccountInquiryService_UpsertAccountInquiriesIntoBeneficiary(t *testing.T) {
	validSnapCoreResp := &snapCoreModel.InquiryAccountResponseData{
		ResponseCode:           "200xx00",
		ResponseMessage:        "success",
		BeneficiaryAccountName: "John Doe",
		BeneficiaryAccountNo:   "1234567890",
		BeneficiaryBankCode:    "008",
		BeneficiaryBankName:    "BCA",
	}

	merchantID := "test-merchant-id"
	ctx := context.Background()
	beneficiaryID := uuid.NewString()

	// Make sure to not use AssertExpectations on the mock in error cases
	// since we want to verify only the specific error path behavior

	testCases := []struct {
		desc               string
		mockSetup          func(beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository)
		wantErr            bool
		wantID             string
		verifyExpectations bool // Add flag to indicate which test cases should verify expectations
	}{
		{
			desc: "error when getting existing beneficiary account",
			mockSetup: func(beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				beneficiaryRepo.On("GetByBankCodeAndAccountNo", ctx, merchantID, validSnapCoreResp.BeneficiaryBankCode, validSnapCoreResp.BeneficiaryAccountNo).
					Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			wantID:  "",
		},
		{
			desc: "successfully update existing beneficiary account",
			mockSetup: func(beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				existingBeneficiary := &beneficiaryModel.BeneficiaryAccount{
					UUID:                 beneficiaryID,
					MerchantID:           merchantID,
					BeneficiaryAccountNo: "1234567890",
					BeneficiaryBankCode:  "008",
					MetadataObj:          beneficiaryModel.Metadata{},
					Metadata:             types.NullJSONText{},
				}
				beneficiaryRepo.On("GetByBankCodeAndAccountNo", ctx, merchantID, validSnapCoreResp.BeneficiaryBankCode, validSnapCoreResp.BeneficiaryAccountNo).
					Return(existingBeneficiary, nil)
				beneficiaryRepo.On("Update", ctx, mock.MatchedBy(func(ba *beneficiaryModel.BeneficiaryAccount) bool {
					return ba.UUID == beneficiaryID &&
						ba.BeneficiaryAccountName == validSnapCoreResp.BeneficiaryAccountName &&
						ba.BeneficiaryBankCode == validSnapCoreResp.BeneficiaryBankCode &&
						ba.BeneficiaryAccountNo == validSnapCoreResp.BeneficiaryAccountNo &&
						ba.MetadataObj.RequestInquiryStatus == c.RequestAccountInquiryStatusValid &&
						ba.Metadata.Valid == true
				})).Return(nil)
			},
			wantErr:            false,
			wantID:             beneficiaryID,
			verifyExpectations: true,
		},
		{
			desc: "error when updating existing beneficiary account",
			mockSetup: func(beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				existingBeneficiary := &beneficiaryModel.BeneficiaryAccount{
					UUID:                 beneficiaryID,
					MerchantID:           merchantID,
					BeneficiaryAccountNo: "1234567890",
					BeneficiaryBankCode:  "008",
				}
				beneficiaryRepo.On("GetByBankCodeAndAccountNo", ctx, merchantID, validSnapCoreResp.BeneficiaryBankCode, validSnapCoreResp.BeneficiaryAccountNo).
					Return(existingBeneficiary, nil)
				beneficiaryRepo.On("Update", ctx, mock.Anything).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:            true,
			wantID:             beneficiaryID,
			verifyExpectations: true,
		},
		{
			desc: "successfully create new beneficiary account",
			mockSetup: func(beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				beneficiaryRepo.On("GetByBankCodeAndAccountNo", ctx, merchantID, validSnapCoreResp.BeneficiaryBankCode, validSnapCoreResp.BeneficiaryAccountNo).
					Return(nil, nil)
				beneficiaryRepo.On("Create", ctx, mock.MatchedBy(func(ba *beneficiaryModel.BeneficiaryAccount) bool {
					return ba.MerchantID == merchantID &&
						ba.BeneficiaryAccountName == validSnapCoreResp.BeneficiaryAccountName &&
						ba.BeneficiaryBankCode == validSnapCoreResp.BeneficiaryBankCode &&
						ba.BeneficiaryAccountNo == validSnapCoreResp.BeneficiaryAccountNo
				})).Return(nil)
			},
			wantErr:            false,
			wantID:             "",
			verifyExpectations: true,
		},
		{
			desc: "error when creating new beneficiary account",
			mockSetup: func(beneficiaryRepo *mocks_repo.IBeneficiaryAccountRepository) {
				beneficiaryRepo.On("GetByBankCodeAndAccountNo", ctx, merchantID, validSnapCoreResp.BeneficiaryBankCode, validSnapCoreResp.BeneficiaryAccountNo).
					Return(nil, nil)
				beneficiaryRepo.On("Create", ctx, mock.Anything).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:            true,
			wantID:             "",
			verifyExpectations: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			logger, _ := mocks_logger.NewZapLogger(mocks_logger.Config{})
			beneficiaryRepo := mocks_repo.NewIBeneficiaryAccountRepository(t)

			tc.mockSetup(beneficiaryRepo)

			svc := AccountInquiryService{
				beneficiaryRepo: beneficiaryRepo,
				logger:          logger,
			}

			gotID, err := svc.UpsertAccountInquiriesIntoBeneficiary(ctx, merchantID, validSnapCoreResp)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.wantID != "" {
					assert.Equal(t, tc.wantID, gotID)
				} else {
					assert.NotEmpty(t, gotID) // for the creation case, we get a new UUID
				}
			}
		})
	}
}
