package refundProcessorService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/test"
)

func TestBankTransferStrategyProcess(t *testing.T) {

	_, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		request       *refundModel.RefundProcessRequest
		setupMocks    func(*mocks.ISnapCoreRepository, *mockServices.IBeneficiaryAccountService, *mockServices.IOrchestratorService)
		expectedError error
	}{
		{
			name: "success - bank transfer refund processed successfully",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				RefundLedgerID:           "ledger-123",
				PaymentProcessorID:       "bt-123",
				PaymentClientReferenceID: "client-ref-123",
				Refund: &refundModel.Refund{
					UUID:       "refund-123",
					MerchantID: "merchant-123",
					Amount:     1000.0,
					CreatedAt:  time.Now(),
					Reason:     "customer requested",
					MetadataObj: refundModel.MetadataObj{
						TransferDestination: &refundModel.TransferDestination{
							ChannelCode: "BCA",
							ChannelInformation: refundModel.ChannelInformation{
								AccountNumber: "1234567890",
								AccountName:   "John Doe",
							},
							Description: "Refund payment",
						},
					},
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository, beneficiaryAccountSvc *mockServices.IBeneficiaryAccountService, orchestratorSvc *mockServices.IOrchestratorService) {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, &beneficiaryAccountModel.CheckAccountRequest{
					BeneficiaryBankCode:  "014",
					BeneficiaryAccountNo: "1234567890",
					MerchantID:           "merchant-123",
					AdditionalInfo:       map[string]any{},
				}).Return(&beneficiaryAccountModel.Account{}, nil)

				snapCoreRepo.On("BankTransfer", mock.Anything, mock.MatchedBy(func(req *snapCoreModel.BankTransferRequest) bool {
					return req.BTBeneficiaryRequest.BeneficiaryBankCode == "014" &&
						req.BTBeneficiaryRequest.BeneficiaryAccountNo == "1234567890" &&
						req.BTBeneficiaryRequest.BeneficiaryAccountName == "John Doe" &&
						req.Currency == constant.CurrencyIDR &&
						req.Amount.Value == "1000" &&
						req.Remark == "Refund payment"
				}), mock.MatchedBy(func(header *snapCoreModel.BankTransferHeaderRequest) bool {
					return header.ExternalId == "ledger-123" && header.MerchantId == "merchant-123"
				})).Return(&snapCoreModel.BankTransferResponseData{
					UUID:   "bt-resp-123",
					Status: constant.SnapCoreBankTransferStatusSuccess,
				}, nil)

				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, constant.StringMockType(), constant.SnapCoreProcessor, constant.StringMockType(), constant.StringMockType()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "fail - transfer destination not available",
			request: &refundModel.RefundProcessRequest{
				RefundID:       "refund-123",
				RefundLedgerID: "ledger-123",
				Refund: &refundModel.Refund{
					MetadataObj: refundModel.MetadataObj{},
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository, beneficiaryAccountSvc *mockServices.IBeneficiaryAccountService, orchestratorSvc *mockServices.IOrchestratorService) {
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, "ledger-123", constant.SnapCoreProcessor, "", "").Return(nil)
			},
			expectedError: errors.New("ERROR_UNPROCESSABLE_CONTENT | refund transfer destination not available"),
		},
		{
			name: "fail - invalid channel code",
			request: &refundModel.RefundProcessRequest{
				RefundID:       "refund-123",
				RefundLedgerID: "ledger-123",
				Refund: &refundModel.Refund{
					MetadataObj: refundModel.MetadataObj{
						TransferDestination: &refundModel.TransferDestination{
							ChannelCode: "999",
						},
					},
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository, beneficiaryAccountSvc *mockServices.IBeneficiaryAccountService, orchestratorSvc *mockServices.IOrchestratorService) {
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, "ledger-123", constant.SnapCoreProcessor, "", "").Return(nil)
			},
			expectedError: errors.New("ERROR_UNPROCESSABLE_CONTENT | channel code not found"),
		},
		{
			name: "fail - beneficiary account inquiry error",
			request: &refundModel.RefundProcessRequest{
				RefundID:       "refund-123",
				RefundLedgerID: "ledger-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					MetadataObj: refundModel.MetadataObj{
						TransferDestination: &refundModel.TransferDestination{
							ChannelCode: "BCA",
							ChannelInformation: refundModel.ChannelInformation{
								AccountNumber: "1234567890",
								AccountName:   "John Doe",
							},
						},
					},
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository, beneficiaryAccountSvc *mockServices.IBeneficiaryAccountService, orchestratorSvc *mockServices.IOrchestratorService) {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, &beneficiaryAccountModel.CheckAccountRequest{
					BeneficiaryBankCode:  "014",
					BeneficiaryAccountNo: "1234567890",
					MerchantID:           "merchant-123",
					AdditionalInfo:       map[string]any{},
				}).Return(nil, errors.New("account inquiry failed"))

				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, "ledger-123", constant.SnapCoreProcessor, "", "").Return(nil)
			},
			expectedError: errors.New("account inquiry failed"),
		},
		{
			name: "fail - snap core bank transfer error",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				RefundLedgerID:           "ledger-123",
				PaymentProcessorID:       "bt-123",
				PaymentClientReferenceID: "client-ref-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Amount:     1000.0,
					CreatedAt:  time.Now(),
					Reason:     "customer requested",
					MetadataObj: refundModel.MetadataObj{
						TransferDestination: &refundModel.TransferDestination{
							ChannelCode: "BCA",
							ChannelInformation: refundModel.ChannelInformation{
								AccountNumber: "1234567890",
								AccountName:   "John Doe",
							},
							Description: "Refund payment",
						},
					},
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository, beneficiaryAccountSvc *mockServices.IBeneficiaryAccountService, orchestratorSvc *mockServices.IOrchestratorService) {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(&beneficiaryAccountModel.Account{}, nil)

				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Return(&snapCoreModel.BankTransferResponseData{
					UUID: "bt-resp-123",
				}, errors.New("bank transfer failed"))

				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, constant.StringMockType(), constant.SnapCoreProcessor, constant.StringMockType(), constant.StringMockType()).Return(nil)
			},
			expectedError: errors.New("bank transfer failed"),
		},
		{
			name: "fail - bank transfer status pending",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				RefundLedgerID:           "ledger-123",
				PaymentProcessorID:       "bt-123",
				PaymentClientReferenceID: "client-ref-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Amount:     1000.0,
					CreatedAt:  time.Now(),
					Reason:     "customer requested",
					MetadataObj: refundModel.MetadataObj{
						TransferDestination: &refundModel.TransferDestination{
							ChannelCode: "BCA",
							ChannelInformation: refundModel.ChannelInformation{
								AccountNumber: "1234567890",
								AccountName:   "John Doe",
							},
							Description: "Refund payment",
						},
					},
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository, beneficiaryAccountSvc *mockServices.IBeneficiaryAccountService, orchestratorSvc *mockServices.IOrchestratorService) {
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(&beneficiaryAccountModel.Account{}, nil)

				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Return(&snapCoreModel.BankTransferResponseData{
					UUID:   "bt-resp-123",
					Status: constant.SnapCoreBankTransferStatusPending,
				}, nil)

				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, constant.StringMockType(), constant.SnapCoreProcessor, constant.StringMockType(), constant.StringMockType()).Return(nil)
			},
			expectedError: constant.ErrBankTransferStillInPending,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSnapCoreRepo := mocks.NewISnapCoreRepository(t)
			mockBeneficiaryAccountSvc := mockServices.NewIBeneficiaryAccountService(t)
			mockOrchestratorSvc := mockServices.NewIOrchestratorService(t)

			tc.setupMocks(mockSnapCoreRepo, mockBeneficiaryAccountSvc, mockOrchestratorSvc)

			defer mockSnapCoreRepo.AssertExpectations(t)
			defer mockBeneficiaryAccountSvc.AssertExpectations(t)
			defer mockOrchestratorSvc.AssertExpectations(t)

			strategy := &BankTransferStrategy{
				snapCoreRepo:          mockSnapCoreRepo,
				beneficiaryAccountSvc: mockBeneficiaryAccountSvc,
				orchestratorSvc:       mockOrchestratorSvc,
				logger:                pdkLogger,
			}

			err := strategy.Process(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
