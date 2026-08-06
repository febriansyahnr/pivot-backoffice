package accountinquiry

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	mocks_repo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks_service "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestCheckStatusRequestInquiry(t *testing.T) {
	type mocker struct {
		repo                   *mocks_repo.IRequestAccountInquiryRepository
		beneficiaryAccountRepo *mocks_repo.IBeneficiaryAccountRepository
		routingProcessorSvc    *mocks_service.IRoutingProcessorService
	}
	showAdditionalInfo := true
	hideAdditionalInfo := false

	testCases := []struct {
		desc                   string
		merchantID             string
		inquiryID              string
		wantErr                bool
		expectAdditionalInfo   *bool
		expectedVirtualAccount bool
		expectedInquiryStatus  string
		mockSetup              func(m *mocker)
	}{
		{
			desc:    "error when find latest by inquiry id",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, assert.AnError)
			},
		},
		{
			desc:    "error when find latest by inquiry id data not found",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
		},
		{
			desc:    "success when find latest by inquiry id found with status is not pending",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:       "uuid",
						MerchantID: "merchant_id",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusValid,
						},
					},
				}, nil)
			},
		},
		{
			desc:    "status pending when process to processor",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2022400",
					ResponseMessage:        "Request In progress",
					BeneficiaryAccountName: "Yories Yolanda",
				}, nil)
			},
		},
		{
			desc:    "error when process begin transaction",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(nil, assert.AnError)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), assert.AnError)
			},
		},
		{
			desc:    "error when update request account inquiry",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(assert.AnError)
				m.repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error when update account inquiries",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(nil)
				m.beneficiaryAccountRepo.On("GetByID", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("string")).Return(&beneficiaryAccountModel.BeneficiaryAccount{}, nil)
				m.beneficiaryAccountRepo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount")).Return(assert.AnError)
				m.repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error when commit transaction",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(nil)
				m.beneficiaryAccountRepo.On("GetByID", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("string")).Return(&beneficiaryAccountModel.BeneficiaryAccount{}, nil)
				m.beneficiaryAccountRepo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount")).Return(nil)
				m.repo.On("CommitTransaction", mock.Anything).Return(assert.AnError)
				m.repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "success check status request inquiry with status valid - account names match",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						BeneficiaryAccountName: sql.NullString{
							String: "Yories Yolanda",
							Valid:  true,
						},
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(nil)
				m.beneficiaryAccountRepo.On("GetByID", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("string")).Return(&beneficiaryAccountModel.BeneficiaryAccount{}, nil)
				m.beneficiaryAccountRepo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount")).Return(nil)
				m.repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "success check status request inquiry",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(nil)
				m.beneficiaryAccountRepo.On("GetByID", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("string")).Return(&beneficiaryAccountModel.BeneficiaryAccount{}, nil)
				m.beneficiaryAccountRepo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount")).Return(nil)
				m.repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error when GetByID beneficiary account returns error",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(nil)
				m.beneficiaryAccountRepo.On("GetByID", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("string")).Return(nil, assert.AnError)
				m.repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error when GetByID beneficiary account returns nil (not found)",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(nil)
				m.beneficiaryAccountRepo.On("GetByID", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("string")).Return(nil, nil)
				m.repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error when RollbackTransaction fails",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "uuid",
						MerchantID:          "merchant_id",
						BeneficiaryBankCode: "022",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
						},
					},
				}, nil)

				m.routingProcessorSvc.On("AccountInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "2002400",
					ResponseMessage:        "Successfull",
					BeneficiaryAccountName: "Yories Yolanda",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "022",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryBankName:    "BNI",
					ProcessorReference:     constant.SnapCoreProcessor,
				}, nil)

				m.repo.On("BeginTransaction", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(context.Background(), nil)
				m.repo.On("Update", mock.AnythingOfType(constant.MockTypeBackgroundContext), mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster")).Return(assert.AnError)
				m.repo.On("RollbackTransaction", mock.Anything).Return(assert.AnError)
			},
		},
		{
			desc:                   "success non pending and flag on reads metadata virtual account",
			merchantID:             "merchant-flag-on",
			inquiryID:              "inquiry-id",
			wantErr:                false,
			expectAdditionalInfo:   &showAdditionalInfo,
			expectedVirtualAccount: true,
			expectedInquiryStatus:  constant.RequestAccountInquiryStatusValid,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID", mock.Anything, "inquiry-id", "merchant-flag-on").Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:       "request-id",
						MerchantID: "merchant-flag-on",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusValid,
							Valid:  true,
						},
						Metadata: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(`{"snapCoreResponse":{"isVirtualAccount":true}}`),
						},
					},
				}, nil)
			},
		},
		{
			desc:                  "success non pending and flag off hides additional info",
			merchantID:            "merchant-flag-off",
			inquiryID:             "inquiry-id",
			wantErr:               false,
			expectAdditionalInfo:  &hideAdditionalInfo,
			expectedInquiryStatus: constant.RequestAccountInquiryStatusValid,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID", mock.Anything, "inquiry-id", "merchant-flag-off").Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:       "request-id",
						MerchantID: "merchant-flag-off",
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusValid,
							Valid:  true,
						},
						Metadata: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(`{"snapCoreResponse":{"isVirtualAccount":true}}`),
						},
					},
				}, nil)
			},
		},
		{
			desc:                   "success pending and flag on keeps metadata as primary source",
			merchantID:             "merchant-flag-on",
			inquiryID:              "inquiry-id",
			wantErr:                false,
			expectAdditionalInfo:   &showAdditionalInfo,
			expectedVirtualAccount: false,
			expectedInquiryStatus:  constant.RequestAccountInquiryStatusPending,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID", mock.Anything, "inquiry-id", "merchant-flag-on").Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "request-id",
						MerchantID:          "merchant-flag-on",
						BeneficiaryBankCode: "008",
						BeneficiaryAccountNo: sql.NullString{
							String: "1234567890",
							Valid:  true,
						},
						BeneficiaryAccountName: sql.NullString{
							String: "Name",
							Valid:  true,
						},
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
							Valid:  true,
						},
						Metadata: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(`{"snapCoreResponse":{"isVirtualAccount":false}}`),
						},
					},
				}, nil)
				m.routingProcessorSvc.On("AccountInquiry", mock.Anything, mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest")).
					Return(&routingProcessorModel.InquiryAccountResponseData{
						ResponseCode:           "2022400",
						ResponseMessage:        "Request In progress",
						BeneficiaryAccountName: "Name",
						IsVirtualAccount:       true,
					}, nil)
			},
		},
		{
			desc:                   "success pending and flag on falls back to latest response when metadata missing",
			merchantID:             "merchant-flag-on",
			inquiryID:              "inquiry-id",
			wantErr:                false,
			expectAdditionalInfo:   &showAdditionalInfo,
			expectedVirtualAccount: true,
			expectedInquiryStatus:  constant.RequestAccountInquiryStatusPending,
			mockSetup: func(m *mocker) {
				m.repo.On("FindLatestByInquiryID", mock.Anything, "inquiry-id", "merchant-flag-on").Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						UUID:                "request-id",
						MerchantID:          "merchant-flag-on",
						BeneficiaryBankCode: "008",
						BeneficiaryAccountNo: sql.NullString{
							String: "1234567890",
							Valid:  true,
						},
						BeneficiaryAccountName: sql.NullString{
							String: "Name",
							Valid:  true,
						},
						Status: sql.NullString{
							String: constant.RequestAccountInquiryStatusPending,
							Valid:  true,
						},
						Metadata: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(`{}`),
						},
					},
				}, nil)
				m.routingProcessorSvc.On("AccountInquiry", mock.Anything, mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest")).
					Return(&routingProcessorModel.InquiryAccountResponseData{
						ResponseCode:           "2022400",
						ResponseMessage:        "Request In progress",
						BeneficiaryAccountName: "Name",
						IsVirtualAccount:       true,
					}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				repo:                   mocks_repo.NewIRequestAccountInquiryRepository(t),
				beneficiaryAccountRepo: mocks_repo.NewIBeneficiaryAccountRepository(t),
				routingProcessorSvc:    mocks_service.NewIRoutingProcessorService(t),
			}
			logger, _ := logger.NewZapLogger(logger.Config{})

			ffContentConfig := `
backend-portal-account-inquiry-display-virtual-account-flag-for-whitelisted-merchant:
  variations:
    ON: true
    OFF: false
  targeting:
    - name: allowed merchant
      query: merchant_id in ["merchant-flag-on"]
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

			tc.mockSetup(m)

			svc := New(logger, nil, m.repo, nil, nil, nil, nil, WithRoutingProcessorService(m.routingProcessorSvc), WithBeneficiaryAccountRepository(m.beneficiaryAccountRepo))
			resp, err := svc.CheckStatusRequestInquiry(context.Background(), tc.merchantID, tc.inquiryID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				if tc.expectedInquiryStatus != "" {
					assert.Equal(t, tc.expectedInquiryStatus, resp.InquiryResult.Status)
				}
				if tc.expectAdditionalInfo != nil {
					if *tc.expectAdditionalInfo {
						require.NotNil(t, resp.AdditionalInfo)
						assert.Equal(t, tc.expectedVirtualAccount, resp.AdditionalInfo.IsVirtualAccount)
					} else {
						assert.Nil(t, resp.AdditionalInfo)
					}
				}
			}

			m.beneficiaryAccountRepo.AssertExpectations(t)
			m.repo.AssertExpectations(t)
			m.routingProcessorSvc.AssertExpectations(t)
		})
	}
}
