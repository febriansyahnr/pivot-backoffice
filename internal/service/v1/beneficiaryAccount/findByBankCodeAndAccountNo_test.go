package beneficiaryAccountService

import (
	"context"

	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocksSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	ffMock "github.com/thomaspoignant/go-feature-flag/testutils/mock"
)

func TestBeneficiaryAccountServiceFindByBankCodeAndAccountNo(t *testing.T) {
	// Valid use case
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	projectRoot, err := util.FindProjectRoot(currentDir, "backend-portal")
	assert.NoError(t, err)

	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	err = ffclient.Init(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: targetPath},
		LeveledLogger:   slog.Default(),
		DataExporter: ffclient.DataExporter{
			FlushInterval:    10 * time.Second,
			MaxEventInMemory: 1000,
			Exporter: &ffMock.Exporter{
				Bulk: true,
			},
		},
	})
	defer ffclient.Close()

	assert.NoError(t, err)

	beneficiaryAccount := &beneficiaryAccountModel.BeneficiaryAccount{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryAccountName: "testing",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "testing bank",
		MetadataObj: beneficiaryAccountModel.Metadata{
			RequestInquiryStatus: constant.RequestAccountInquiryStatusValid,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// accountInquiryExisted := &accountInquiries.AccountInquiries{
	// 	UUID:                   "uuid-uuid-uuid",
	// 	BeneficiaryAccountNo:   "12341234",
	// 	BeneficiaryAccountName: "testing",
	// 	BeneficiaryBankCode:    "1234",
	// 	BeneficiaryBankName:    "testing bank",
	// 	Response:               "{\"account_number\":\"12341234\",\"bank_code\":\"1234\",\"bank_name\":\"testing bank\",\"name\":\"testing\"}",
	// 	CreatedAt:              time.Now(),
	// 	UpdatedAt:              time.Now(),
	// }

	reqPayload := &beneficiaryAccountModel.CheckAccountRequest{
		BeneficiaryAccountNo: "12341234",
		BeneficiaryBankCode:  "1234",
		MerchantID:           "merchant-merchant-merchant",
		AdditionalInfo:       nil,
	}

	reqWhitelistedPayload := &beneficiaryAccountModel.CheckAccountRequest{
		BeneficiaryAccountNo: "111501015932505",
		BeneficiaryBankCode:  "1234",
		MerchantID:           "merchant-merchant-merchant",
		AdditionalInfo:       nil,
	}

	snapRes := &routingProcessorModel.InquiryAccountResponseData{
		ResponseCode:           "1234",
		ResponseMessage:        "success",
		PartnerReferenceNo:     "12345",
		BeneficiaryAccountName: "testing",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "test",
	}

	snapResWithEmptyBankName := &routingProcessorModel.InquiryAccountResponseData{
		ResponseCode:           "1234",
		ResponseMessage:        "success",
		PartnerReferenceNo:     "12345",
		BeneficiaryAccountName: "testing",
		BeneficiaryAccountNo:   "12341234",
		BeneficiaryBankCode:    "002", // BRI
	}
	virtualAccountTrue := true
	virtualAccountFalse := false

	testCases := []struct {
		name                     string
		request                  *beneficiaryAccountModel.CheckAccountRequest
		mockSetup                func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService)
		wantErr                  bool
		expectedIsVirtualAccount *bool
	}{
		{
			name:    "SUCCESS: Beneficiary Account is exist with status is valid",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(beneficiaryAccount, nil)
			},
			wantErr: false,
		},
		{
			name:    "FAILED: failed to get beneficiary account",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "SUCCESS: Get Bank Account Inquiry",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

				ben.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Get Bank Account Inquiry with empty bank name",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapResWithEmptyBankName, nil)

				ben.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Whitelist Account No",
			request: reqWhitelistedPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

			},
			wantErr: false,
		},
		{
			name:    "FAILED: Get Bank Account Inquiry from snapcore",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {

				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(nil, assert.AnError)

			},
			wantErr: true,
		},
		{
			name:    "FAILED: Create beneficiary account",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {

				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

				ben.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "FAILED: Update beneficiary account",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {

				benefEmptyName := *beneficiaryAccount
				benefEmptyName.BeneficiaryAccountName = ""
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&benefEmptyName, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

				ben.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "SUCCESS: Update beneficiary account",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {

				benefEmptyName := *beneficiaryAccount
				benefEmptyName.BeneficiaryAccountName = ""
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&benefEmptyName, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

				ben.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Beneficiary Account is exist, but not valid",
			request: reqPayload,
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				warningBeneficiaryAccount := beneficiaryAccount
				warningBeneficiaryAccount.MetadataObj = beneficiaryAccountModel.Metadata{
					RequestInquiryStatus: constant.RequestAccountInquiryStatusPending,
				}

				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(warningBeneficiaryAccount, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

				ben.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Legacy metadata without virtual account defaults to non-VA",
			request: &beneficiaryAccountModel.CheckAccountRequest{
				BeneficiaryAccountNo: "22223333",
				BeneficiaryBankCode:  "008",
				MerchantID:           "merchant-id",
			},
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On("GetByBankCodeAndAccountNo", mock.Anything, mock.Anything, "008", "22223333").Return(&beneficiaryAccountModel.BeneficiaryAccount{
					UUID:                   "beneficiary-id",
					MerchantID:             "merchant-id",
					BeneficiaryAccountNo:   "22223333",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Name",
					MetadataObj: beneficiaryAccountModel.Metadata{
						RequestInquiryStatus: constant.RequestAccountInquiryStatusValid,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			wantErr:                  false,
			expectedIsVirtualAccount: &virtualAccountFalse,
		},
		{
			name: "SUCCESS: Refresh inquiry updates metadata with latest virtual account value",
			request: &beneficiaryAccountModel.CheckAccountRequest{
				BeneficiaryAccountNo: "22223333",
				BeneficiaryBankCode:  "008",
				MerchantID:           "merchant-id",
			},
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On("GetByBankCodeAndAccountNo", mock.Anything, mock.Anything, "008", "22223333").Return(&beneficiaryAccountModel.BeneficiaryAccount{
					UUID:                   "beneficiary-id",
					MerchantID:             "merchant-id",
					BeneficiaryAccountNo:   "22223333",
					BeneficiaryAccountName: "",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Name",
					MetadataObj: beneficiaryAccountModel.Metadata{
						RequestInquiryStatus: constant.RequestAccountInquiryStatusPending,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
				routingProcessorSvc.On("AccountInquiry", mock.Anything, mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest")).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "success",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryAccountNo:   "22223333",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Name",
					IsVirtualAccount:       true,
				}, nil)
				ben.On(
					"Update",
					mock.Anything,
					mock.MatchedBy(func(updated *beneficiaryAccountModel.BeneficiaryAccount) bool {
						return updated.MetadataObj.IsVirtualAccount &&
							updated.MetadataObj.RequestInquiryStatus == constant.RequestAccountInquiryStatusValid
					}),
				).Return(nil)
			},
			wantErr:                  false,
			expectedIsVirtualAccount: &virtualAccountTrue,
		},
		{
			name: "SUCCESS: New inquiry creates metadata with virtual account value",
			request: &beneficiaryAccountModel.CheckAccountRequest{
				BeneficiaryAccountNo: "22223333",
				BeneficiaryBankCode:  "008",
				MerchantID:           "merchant-id",
			},
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On("GetByBankCodeAndAccountNo", mock.Anything, mock.Anything, "008", "22223333").Return(nil, nil)
				routingProcessorSvc.On("AccountInquiry", mock.Anything, mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest")).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "success",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryAccountNo:   "22223333",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Name",
					IsVirtualAccount:       true,
				}, nil)
				ben.On(
					"Create",
					mock.Anything,
					mock.MatchedBy(func(created *beneficiaryAccountModel.BeneficiaryAccount) bool {
						return created.MetadataObj.IsVirtualAccount &&
							created.MetadataObj.RequestInquiryStatus == constant.RequestAccountInquiryStatusValid
					}),
				).Return(nil)
			},
			wantErr:                  false,
			expectedIsVirtualAccount: &virtualAccountTrue,
		},
	}

	// Additional test cases for derived merchant ID coverage (lines 102-103)
	derivedMerchantTestCases := []struct {
		name      string
		request   *beneficiaryAccountModel.CheckAccountRequest
		context   context.Context
		mockSetup func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService)
		wantErr   bool
	}{
		{
			name:    "SUCCESS: Use derived merchant ID when present in context",
			request: reqPayload,
			context: context.WithValue(context.Background(), constant.CtxDerivedMerchantID, "derived-merchant-123"),
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					"derived-merchant-123", // Should use derived merchant ID, not the one from request
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(beneficiaryAccount, nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Use empty string when derived merchant ID is empty string in context",
			request: reqPayload,
			context: context.WithValue(context.Background(), constant.CtxDerivedMerchantID, ""),
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					"", // Should use empty string as merchant ID since type assertion succeeds
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(beneficiaryAccount, nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Use original merchant ID when derived merchant ID is not string type in context",
			request: reqPayload,
			context: context.WithValue(context.Background(), constant.CtxDerivedMerchantID, 12345), // Non-string type
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					reqPayload.MerchantID, // Should use original merchant ID from request
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(beneficiaryAccount, nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Derived merchant ID used when beneficiary account not found and needs inquiry",
			request: reqPayload,
			context: context.WithValue(context.Background(), constant.CtxDerivedMerchantID, "derived-merchant-456"),
			mockSetup: func(ben *mocks.IBeneficiaryAccountRepository, accountInquiry *mocks.IAccountInquiriesRepository, routingProcessorSvc *mocksSvc.IRoutingProcessorService) {
				ben.On(
					"GetByBankCodeAndAccountNo",
					mock.Anything,
					"derived-merchant-456", // Should use derived merchant ID
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				routingProcessorSvc.On(
					"AccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(snapRes, nil)

				ben.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(nil)
			},
			wantErr: false,
		},
	}

	// Run derived merchant ID test cases
	for _, tc := range derivedMerchantTestCases {
		t.Run(tc.name, func(t *testing.T) {
			benMock := mocks.NewIBeneficiaryAccountRepository(t)
			accountInquiryRepoMock := mocks.NewIAccountInquiriesRepository(t)
			snapMock := mocks.NewISnapCoreRepository(t)
			routingProcSvc := mocksSvc.NewIRoutingProcessorService(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mockSetup(benMock, accountInquiryRepoMock, routingProcSvc)

			svc := New(loggerMock, benMock, accountInquiryRepoMock, snapMock, WithRoutingProcessorService(routingProcSvc), WithConfig(&config.Config{Environment: "test"}))
			account, err := svc.FindByBankCodeAndAccountNo(tc.context, tc.request)

			if !tc.wantErr {
				assert.NoError(t, err)
				require.NotEmpty(t, account)
			} else {
				require.Error(t, err)
				require.Empty(t, account)
			}

			benMock.AssertExpectations(t)
			accountInquiryRepoMock.AssertExpectations(t)
			snapMock.AssertExpectations(t)
			routingProcSvc.AssertExpectations(t)
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			benMock := mocks.NewIBeneficiaryAccountRepository(t)
			accountInquiryRepoMock := mocks.NewIAccountInquiriesRepository(t)
			snapMock := mocks.NewISnapCoreRepository(t)
			routingProcSvc := mocksSvc.NewIRoutingProcessorService(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mockSetup(benMock, accountInquiryRepoMock, routingProcSvc)

			svc := New(loggerMock, benMock, accountInquiryRepoMock, snapMock, WithRoutingProcessorService(routingProcSvc), WithConfig(&config.Config{Environment: "test"}))
			account, err := svc.FindByBankCodeAndAccountNo(context.Background(), tc.request)

			if !tc.wantErr {
				assert.NoError(t, err)
				require.NotEmpty(t, account)
				if tc.expectedIsVirtualAccount != nil {
					assert.Equal(t, *tc.expectedIsVirtualAccount, account.MetadataObj.IsVirtualAccount)
				}
			} else {
				require.Error(t, err)
				require.Empty(t, account)
			}

			benMock.AssertExpectations(t)
			accountInquiryRepoMock.AssertExpectations(t)
			snapMock.AssertExpectations(t)

			routingProcSvc.AssertExpectations(t)
		})
	}
}
