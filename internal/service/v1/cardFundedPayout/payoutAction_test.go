package cardFundedPayoutService_test

import (
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
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

func TestCreatePayout(t *testing.T) {
	log := logMock.NewILogger(t)
	feeSvc := serviceMocks.NewIFeeService(t)
	vendorSvc := serviceMocks.NewIVendorService(t)
	customerSvc := serviceMocks.NewICustomerService(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)

	service := New(nil, log,
		WithCustomerService(customerSvc),
		WithVendorService(vendorSvc),
		WithFeeService(feeSvc),
		WithDisbursementRepository(disbursementRepo),
		WithStatusHistoriesRepository(statusHistoryRepo),
	)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	vendorID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	cardID := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"
	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	userID := "4dc93f16-93d8-4c2c-a0f2-27c48887617d"

	request := model.CreatePayoutRequest{
		VendorID:    vendorID,
		ReferenceID: referenceID,
		Amount: commonModel.AmountRequest{
			Currency: constant.CurrencyIDR,
			Value:    1_000_000.00,
		},
		Remarks:          "Test payout",
		SettlementMethod: constant.PaymentSettlementMethodStandard,
		CardID:           cardID,
		PayoutActionRequest: model.PayoutActionRequest{
			MerchantID: merchantID,
			UserID:     userID,
		},
	}

	validVendor := &vendor.Vendor{
		UUID:       vendorID,
		MerchantID: merchantID,
		Name:       "PT Test Vendor", // NOSONAR
		Status:     constant.StatusActive,
	}

	validCardDetail := &model.GetSavedCardResponse{
		ID:             cardID,
		CardName:       "Airlines", // NOSONAR
		CardOrigin:     "LOCAL",    // NOSONAR
		PaymentChannel: "VISA",     // NOSONAR
		IssuingBank:    "BNI",      // NOSONAR
	}

	validFeeDetail := &feeModel.FeeMetadataObject{
		Type:          "CARD_FUNDED_PAYOUT", // NOSONAR
		ReferenceType: constant.ReferencePaymentFundedPayout,
		Channel:       "LOCAL_VISA", // NOSONAR
		Method:        paymentConst.PAYMENT_METHOD_CREDIT_CARD,
		FinalAmount:   5000.00, // NOSONAR
	}

	tests := []struct {
		name      string
		request   model.CreatePayoutRequest
		setupMock func()
		wantError error
	}{
		{
			name:    "ERROR:Reference ID already exists",
			request: request,
			setupMock: func() {
				disbursementRepo.On(
					"CountByMerchantAndReference", mock.Anything, merchantID, referenceID,
				).Once().Return(1)
			},
			wantError: pkgErrs.New(response.HttpErrConflict, constant.ErrReferenceIdExist),
		},
		{
			name:    "ERROR:Get vendor detail returns error",
			request: request,
			setupMock: func() {
				disbursementRepo.On(
					"CountByMerchantAndReference", mock.Anything, merchantID, referenceID,
				).Return(0)
				vendorSvc.On(
					"Detail", mock.Anything, vendorID,
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Vendor not found (merchant ID mismatch)",
			request: request,
			setupMock: func() {
				vendorSvc.On(
					"Detail", mock.Anything, vendorID,
				).Once().Return(&vendor.Vendor{
					UUID:       vendorID,
					MerchantID: "different-merchant-id",
					Status:     constant.StatusActive,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, constant.ErrVendorNotFound),
		},
		{
			name:    "ERROR:Vendor not active",
			request: request,
			setupMock: func() {
				vendorSvc.On(
					"Detail", mock.Anything, vendorID,
				).Once().Return(&vendor.Vendor{
					UUID:       vendorID,
					MerchantID: merchantID,
					Status:     "INACTIVE",
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrForbidden, constant.ErrVendorNotActive),
		},
		{
			name:    "ERROR:Get card detail returns error",
			request: request,
			setupMock: func() {
				vendorSvc.On("Detail", mock.Anything, vendorID).Return(validVendor, nil)
				customerSvc.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, model.GetSavedCardDetailRequest{
						CardID:     cardID,
						MerchantID: merchantID,
					},
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Get fee calculation returns error",
			request: request,
			setupMock: func() {
				customerSvc.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, model.GetSavedCardDetailRequest{
						CardID:     cardID,
						MerchantID: merchantID,
					},
				).Return(validCardDetail, nil)
				feeSvc.On(
					"GetFeeCalculationAndDetail", mock.Anything, &feeModel.GetFeeRequest{
						MerchantID:       merchantID,
						Reference:        constant.ReferencePaymentFundedPayout,
						PaymentMethod:    paymentConst.PAYMENT_METHOD_CREDIT_CARD,
						Channel:          validCardDetail.CardOrigin + "_" + validCardDetail.PaymentChannel,
						ReferenceAmount:  request.Amount.Value,
						SettlementMethod: constant.PaymentSettlementMethodStandard,
					},
				).Once().Return(0.0, nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Insert disbursement returns error",
			request: request,
			setupMock: func() {
				feeSvc.On(
					"GetFeeCalculationAndDetail", mock.Anything, mock.Anything,
				).Return(5000.0, validFeeDetail, nil)
				disbursementRepo.On(
					"Insert", mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to create disbursement data", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "SUCCESS:Create payout",
			request: request,
			setupMock: func() {
				disbursementRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.CreatePayout(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if err == nil {
				require.NotNil(t, result)
				require.NotEmpty(t, result.ID)
				require.Equal(t, validVendor.UUID, result.VendorID)
				require.Equal(t, validVendor.Name, result.VendorName)
				require.Equal(t, request.ReferenceID, result.ReferenceID)
				require.Equal(t, 5000.0, result.FeeAmount)
				require.Equal(t, request.Amount, result.Amount)
				require.Equal(t, request.Remarks, result.Remarks)
				require.Equal(t, request.SettlementMethod, result.SettlementMethod)
				require.Equal(t, validCardDetail.ID, result.CardID)
				require.Equal(t, validCardDetail.CardName, result.CardName)
				require.NotNil(t, result.CreatedAt)
			}

			log.AssertExpectations(t)
			feeSvc.AssertExpectations(t)
			vendorSvc.AssertExpectations(t)
			customerSvc.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
		})
	}
}

func TestRejectPayout(t *testing.T) {
	log := logMock.NewILogger(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)

	service := New(nil, log,
		WithDisbursementRepository(disbursementRepo),
		WithStatusHistoriesRepository(statusHistoryRepo),
	)

	payoutID := "5fc93f16-93d8-4c2c-a0f2-27c48887617a"
	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	vendorID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	cardID := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"
	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	userID := "4dc93f16-93d8-4c2c-a0f2-27c48887617d"
	userName := "Test User" // NOSONAR
	reason := "Incorrect vendor details"

	request := model.RejectPayoutRequest{
		ID:     payoutID,
		Reason: reason,
		PayoutActionRequest: model.PayoutActionRequest{
			MerchantID: merchantID,
			UserID:     userID,
			UserName:   userName,
		},
	}

	validPayout := &disbursementModel.Disbursement{
		UUID:        payoutID,
		ReferenceID: referenceID,
		MerchantID:  merchantID,
		Currency:    constant.CurrencyIDR,
		Amount:      decimal.NewFromFloat(1_000_000.00),
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
		},
	}

	tests := []struct {
		name      string
		request   model.RejectPayoutRequest
		setupMock func()
		wantError error
	}{
		{
			name:    "ERROR:GetDetailForCardFundedPayoutByID returns error",
			request: request,
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to retrieve card-funded payout detail", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:Payout not found (nil)",
			request: request,
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Once().Return(nil, nil)
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
			name:    "ERROR:Reject returns error",
			request: request,
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Return(validPayout, nil)
				disbursementRepo.On(
					"Reject", mock.Anything, payoutID, constant.DisbursementReasonTypeOther, reason, userID,
				).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed when reject card-funded payout", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "SUCCESS:Reject payout",
			request: request,
			setupMock: func() {
				disbursementRepo.On("Reject", mock.Anything, payoutID, constant.DisbursementReasonTypeOther, reason, userID).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.RejectPayout(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if err == nil {
				require.NotNil(t, result)
				require.Equal(t, validPayout.UUID, result.ID)
				require.Equal(t, validPayout.MetadataObj.CardFundedDetail.VendorID, result.VendorID)
				require.Equal(t, validPayout.MetadataObj.CardFundedDetail.VendorName, result.VendorName)
				require.Equal(t, validPayout.ReferenceID, result.ReferenceID)
				require.Equal(t, validPayout.Fee.InexactFloat64(), result.FeeAmount)
				require.Equal(t, validPayout.Currency, result.Amount.Currency)
				require.Equal(t, validPayout.Amount.InexactFloat64(), result.Amount.Value)
				require.Equal(t, *validPayout.Remark, result.Remarks)
				require.Equal(t, validPayout.MetadataObj.CardFundedDetail.SettlementMethod, result.SettlementMethod)
				require.Equal(t, validPayout.MetadataObj.CardFundedDetail.Card.ID, result.CardID)
				require.Equal(t, validPayout.MetadataObj.CardFundedDetail.Card.CardName, result.CardName)
				require.Equal(t, &reason, result.RejectReason)
				require.NotNil(t, result.RejectedAt)
			}

			log.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
		})
	}
}
