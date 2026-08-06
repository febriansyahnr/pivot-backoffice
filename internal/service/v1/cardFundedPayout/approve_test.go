package cardFundedPayoutService_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApprovePayout(t *testing.T) {
	log := logMock.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	customerSvc := serviceMocks.NewICustomerService(t)
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)
	paymentMethodRepo := repoMocks.NewIPaymentMethodRepository(t)

	cfg := &config.Config{}
	service := New(cfg, log,
		WithCustomerService(customerSvc),
		WithPaymentRepository(paymentRepo),
		WithUnifiedPaymentService(unifiedPaymentSvc),
		WithDisbursementRepository(disbursementRepo),
		WithStatusHistoriesRepository(statusHistoryRepo),
		WithPaymentMethodRepository(paymentMethodRepo),
	)

	payoutID := "5fc93f16-93d8-4c2c-a0f2-27c48887617a"
	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	vendorID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	cardID := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"
	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	userID := "4dc93f16-93d8-4c2c-a0f2-27c48887617d"
	userName := "Test User" // NOSONAR

	request := model.ApprovePayoutRequest{
		ID:  payoutID,
		CVC: "123", // NOSONAR
		PayoutActionRequest: model.PayoutActionRequest{
			MerchantID: merchantID,
			UserID:     userID,
			UserName:   userName,
		},
	}

	validPaymentMethod := &paymentModel.PaymentMethodWithPivot{
		MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
			CardFundedPayoutConfig: &paymentMethodModel.CardFundedPayoutConfig{
				Enabled:         true,
				ActiveProcessor: constant.CreditCardPartnerProcessorMPGS,
				Processors: map[string]paymentMethodModel.CardPartnerProcessorConfig{
					"MPGS": {Limit: 2000000000},
					"CYBS": {Limit: 3000000000},
				},
			},
			PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
					Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
						{
							PartnerProcessor:     constant.CreditCardPartnerProcessorMPGS,
							CardFundedPayoutType: constant.CardTransactionTypeCIT,
							AcquirerMerchantID:   "CIT-MERCHANT-ID", // NOSONAR
						},
						{
							PartnerProcessor:     constant.CreditCardPartnerProcessorMPGS,
							CardFundedPayoutType: constant.CardTransactionTypeMIT,
							AcquirerMerchantID:   "MIT-MERCHANT-ID", // NOSONAR
						},
					},
				},
			},
		},
	}

	validPayout := &disbursementModel.Disbursement{
		UUID:        payoutID,
		ReferenceID: referenceID,
		MerchantID:  merchantID,
		Currency:    constant.CurrencyIDR,
		Amount:      decimal.NewFromFloat(1_000_000.00),
		TotalAmount: decimal.NewFromFloat(1_005_000.00),
		Fee:         util.ValueToPtr(decimal.NewFromFloat(5000.00)),
		Status:      constant.DisbursementStatusWaiting,
		Remark:      util.ValueToPtr("Test payout"),
		MetadataObj: disbursementModel.Metadata{
			CardFundedDetail: &disbursementModel.CardFundedDetailMetadata{
				VendorID:         vendorID,
				VendorName:       "PT Test Vendor", // NOSONAR
				SettlementMethod: constant.PaymentSettlementMethodStandard,
				Card: &disbursementModel.CardFundedDetailMetadataCard{
					ID:             cardID,
					CardName:       "Airlines", // NOSONAR
					PaymentChannel: "VISA",     // NOSONAR
					IssuingBank:    "BNI",      // NOSONAR
				},
			},
			FeeDetail: feeModel.FeeMetadataObject{
				Percentage: 2.5,
			},
		},
	}

	validCardDetail := &model.GetSavedCardResponse{
		ID:         cardID,
		CardName:   "Airlines",      // NOSONAR
		CardOrigin: "LOCAL",         // NOSONAR
		CardToken:  "token-abc-123", // NOSONAR
	}

	validPaymentResponse := &unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ID:         "payment-session-id",
		PaymentUrl: "https://example.com/payment", // NOSONAR
	}

	ctxTx := t.Context()

	tests := []struct {
		name       string
		request    model.ApprovePayoutRequest
		setupMock  func()
		wantError  error
		wantResult *model.PayoutActionResponse
	}{
		{
			name:    "ERROR:GetActivePaymentMethodByRequest returns error",
			request: request,
			setupMock: func() {
				paymentMethodRepo.On(
					"GetActivePaymentMethodByRequest", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:Payment method not found",
			request: request,
			setupMock: func() {
				paymentMethodRepo.On(
					"GetActivePaymentMethodByRequest", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotFound),
		},
		{
			name:    "ERROR:Partner config not found",
			request: request,
			setupMock: func() {
				paymentMethodRepo.On(
					"GetActivePaymentMethodByRequest", mock.Anything, mock.Anything,
				).Once().Return(&paymentModel.PaymentMethodWithPivot{MerchantConfigObj: nil}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrPartnerConfigNotFound),
		},
		{
			name:    "ERROR:GetDetailForCardFundedPayoutByID returns error",
			request: request,
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(validPaymentMethod, nil)
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, payoutID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to retrieve card-funded payout detail", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:Payout not found (nil)",
			request: request,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, payoutID,
				).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("payout with ID %s not found", payoutID)),
		},
		{
			name:    "ERROR:Payout not found (merchant ID mismatch)",
			request: request,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, payoutID,
				).Once().Return(&disbursementModel.Disbursement{
					UUID:       payoutID,
					MerchantID: "different-merchant-id",
					Status:     constant.DisbursementStatusWaiting,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("payout with ID %s not found", payoutID)),
		},
		{
			name:    "ERROR:Payout status is not WAITING",
			request: request,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, payoutID,
				).Once().Return(&disbursementModel.Disbursement{
					UUID:       payoutID,
					MerchantID: merchantID,
					Status:     constant.DisbursementStatusApproved,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("payout must be in WAITING status; current status is %s", constant.DisbursementStatusApproved)),
		},
		{
			name:    "ERROR:GetCardFundedPayoutSavedCardDetail returns error",
			request: request,
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Return(validPayout, nil)
				customerSvc.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, model.GetSavedCardDetailRequest{
						MerchantID: merchantID,
						CardID:     cardID,
					},
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:BeginTransaction returns error",
			request: request,
			setupMock: func() {
				customerSvc.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, model.GetSavedCardDetailRequest{
						MerchantID: merchantID,
						CardID:     cardID,
					},
				).Return(validCardDetail, nil)
				disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to create a new database transaction session", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:CreateSession returns error",
			request: request,
			setupMock: func() {
				disbursementRepo.On("BeginTransaction", mock.Anything).Return(ctxTx, nil)
				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				paymentRepo.On("HardDeleteCardFundedPayoutPayments", mock.Anything, merchantID, payoutID).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to cancel the payment session", mock.Anything).Once().Return()
				disbursementRepo.On("RollbackTransaction", context.WithoutCancel(ctxTx)).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to rollback changes", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:ApproveInBulk returns error",
			request: request,
			setupMock: func() {
				validPaymentMethod.MerchantConfigObj.CardFundedPayoutConfig.Processors = map[string]paymentMethodModel.CardPartnerProcessorConfig{
					"MPGS": {Limit: 500_000.00},
					"CYBS": {Limit: 500_000.00},
				}
				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).Times(2).Return(validPaymentResponse, nil)
				disbursementRepo.On("ApproveInBulk", ctxTx, []string{payoutID}, userID).Once().Return(assert.AnError)
				paymentRepo.On("HardDeleteCardFundedPayoutPayments", mock.Anything, merchantID, payoutID).Once().Return(nil)
				disbursementRepo.On("RollbackTransaction", context.WithoutCancel(ctxTx)).Once().Return(nil)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:CommitTransaction returns error",
			request: request,
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).Times(2).Return(validPaymentResponse, nil)
				disbursementRepo.On("ApproveInBulk", ctxTx, []string{payoutID}, userID).Once().Return(nil)
				statusHistoryRepo.On("Insert", ctxTx, mock.Anything).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", ctxTx).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to commit changes", mock.Anything).Once().Return()
				paymentRepo.On("HardDeleteCardFundedPayoutPayments", mock.Anything, merchantID, payoutID).Once().Return(nil)
				disbursementRepo.On("RollbackTransaction", context.WithoutCancel(ctxTx)).Once().Return(nil)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "SUCCESS:Approve payout",
			request: request,
			setupMock: func() {
				validPaymentMethod.MerchantConfigObj.CardFundedPayoutConfig.Processors = map[string]paymentMethodModel.CardPartnerProcessorConfig{
					"MPGS": {Limit: 1_000_000.00},
					"CYBS": {Limit: 1_000_000.00},
				}
				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.MatchedBy(func(payload *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) bool {
					return payload.PaymentMethod.CardPaymentMethodDetail.Token != "" &&
						payload.PaymentMethod.CardPaymentMethodDetail.CVC != ""
				})).Once().Return(validPaymentResponse, nil)
				disbursementRepo.On("ApproveInBulk", ctxTx, []string{payoutID}, userID).Once().Return(nil)
				statusHistoryRepo.On("Insert", ctxTx, mock.Anything).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", ctxTx).Once().Return(nil)
			},
			wantError: nil,
			wantResult: &model.PayoutActionResponse{
				ID:          payoutID,
				VendorID:    vendorID,
				VendorName:  "PT Test Vendor",
				ReferenceID: referenceID,
				FeeAmount:   5000.00,
				Amount: commonModel.AmountRequest{
					Currency: constant.CurrencyIDR,
					Value:    1_000_000.00,
				},
				Remarks:           "Test payout",
				SettlementMethod:  constant.PaymentSettlementMethodStandard,
				CardID:            cardID,
				CardName:          "Airlines",
				AuthenticationUrl: util.ValueToPtr("https://example.com/payment"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.ApprovePayout(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if err == nil {
				test.wantResult.ApprovedAt = result.ApprovedAt
				require.Equal(t, test.wantResult, result)
				require.NotNil(t, result.ApprovedAt)
			}

			log.AssertExpectations(t)
			customerSvc.AssertExpectations(t)
			unifiedPaymentSvc.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
		})
	}
}
