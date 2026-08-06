package cardFundedPayoutService_test

import (
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
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

func TestProcessInitialCardFundedPayoutAuthFailure(t *testing.T) {
	// Create fresh mock instances for each test case
	log := logMock.NewILogger(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	referenceID := "5fc93f16-93d8-4c2c-a0f2-27c48887617a"
	vendorID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	cardID := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"
	paymentID := "payment-uuid-123"
	chargeID := "charge-uuid-123"

	validPayout := &disbursementModel.Disbursement{
		UUID:        referenceID,
		ReferenceID: "REF/PAYOUT/202603/0001",
		MerchantID:  merchantID,
		Currency:    constant.CurrencyIDR,
		Amount:      decimal.NewFromFloat(1_000_000.00),
		Fee:         util.ValueToPtr(decimal.NewFromFloat(5000.00)),
		Status:      constant.DisbursementStatusWaiting,
		Remark:      util.ValueToPtr("Test payout"),
		MetadataObj: disbursementModel.Metadata{
			CardFundedDetail: &disbursementModel.CardFundedDetailMetadata{
				VendorID:         vendorID,
				VendorName:       "PT Test Vendor",
				SettlementMethod: constant.PaymentSettlementMethodStandard,
				Card: &disbursementModel.CardFundedDetailMetadataCard{
					ID:             cardID,
					CardName:       "Airlines",
					PaymentChannel: "VISA",
					IssuingBank:    "BNI",
				},
			},
		},
	}

	validPendingPayments := []model.CardFundedPayment{
		{
			ID:              paymentID,
			ChargeID:        chargeID,
			MerchantID:      merchantID,
			ReferenceID:     referenceID,
			Currency:        constant.CurrencyIDR,
			Fee:             5000.00,
			Amount:          100_000.00,
			Sequence:        1,
			FirstPaymentID:  "first-payment-id",
			CardFingerprint: "card-fingerprint-123",
		},
	}

	tests := []struct {
		name        string
		merchantID  string
		referenceID string
		setupMock   func()
		wantError   error
	}{
		{
			name:        "ERROR: GetDetailForCardFundedPayoutByID returns error",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed when get card-funded payout detail", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:        "ERROR: Payout not found (nil)",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Once().Return(nil, nil)
				log.On("Warn", mock.Anything, "No data found when get card-funded payout detail", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound),
		},
		{
			name:        "ERROR: FindPendingSubsequentCardFundedPayout returns error",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Once().Return(validPayout, nil)
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed when finding pending subsequent card-funded payment transactions", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:        "ERROR: PostAccountTransaction returns error",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Once().Return(validPayout, nil)
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Once().Return([]model.CardFundedPayment{}, nil)
				orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					mock.Anything,
				).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("failed post account transaction: %w", assert.AnError),
		},
		{
			name:        "SUCCESS: Without payout ledger and no subsequent payments",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Once().Return(validPayout, nil)
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Once().Return([]model.CardFundedPayment{}, nil)
				orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name:        "SUCCESS: Without payout ledger and with subsequent payments",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Once().Return(validPayout, nil)
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Once().Return(validPendingPayments, nil)
				orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, fmt.Sprintf("Transaction with reference ID %s has subsequent transactions, all transactions will be canceled", referenceID)).Once().Return()
				// Mock for failPaymentAndChargeOnError
				paymentRepo.On(
					"UpdatePaymentStatusWithReason", mock.Anything, paymentID, mock.Anything,
				).Once().Return(nil)
				accountTransactionRepo.On(
					"UpdateStatusAccountTransaction", mock.Anything, chargeID, constant.StatusFailed, mock.Anything, mock.Anything,
				).Once().Return(nil)
				// Mock for recordPaymentFailedStatus (defer in failPaymentAndChargeOnError)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create service with fresh mocks
			service := New(nil, log,
				WithDisbursementRepository(disbursementRepo),
				WithAccountTransactionRepository(accountTransactionRepo),
				WithPaymentRepository(paymentRepo),
				WithOrchestratorService(orchestratorSvc),
				WithStatusHistoriesRepository(statusHistoryRepo),
			)

			// Setup mocks for this test
			test.setupMock()

			err := service.ProcessInitialCardFundedPayoutAuthFailure(t.Context(), test.merchantID, test.referenceID)
			require.Equal(t, test.wantError, err)

			log.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			accountTransactionRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
		})
	}
}
