package cardFundedPayoutService_test

import (
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	encryptionMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessPendingSubsequentPayments(t *testing.T) {
	log := logMock.NewILogger(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)
	creditCardSvc := serviceMocks.NewICreditCardService(t)
	cryptoProvider := encryptionMock.NewCryptoProvider(t)

	cfg := &config.Config{}
	service := New(cfg, log,
		WithDisbursementRepository(disbursementRepo),
		WithPaymentRepository(paymentRepo),
		WithAccountTransactionRepository(accountTransactionRepo),
		WithStatusHistoriesRepository(statusHistoryRepo),
		WithCreditCardService(creditCardSvc),
		WithCryptoProvider(cryptoProvider),
	)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	vendorID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	paymentID := "payment-id-1" // NOSONAR
	chargeID := "charge-id-1"   // NOSONAR

	validInstantPayout := &disbursementModel.Disbursement{
		UUID:        referenceID,
		ReferenceID: referenceID,
		MerchantID:  merchantID,
		Currency:    constant.CurrencyIDR,
		Amount:      decimal.NewFromFloat(1_000_000.00),
		TotalAmount: decimal.NewFromFloat(1_005_000.00),
		Fee:         util.ValueToPtr(decimal.NewFromFloat(5000.00)),
		Status:      constant.DisbursementStatusApproved,
		MetadataObj: disbursementModel.Metadata{
			CardFundedDetail: &disbursementModel.CardFundedDetailMetadata{
				VendorID:         vendorID,
				VendorName:       "PT Test Vendor", // NOSONAR
				SettlementMethod: constant.PaymentSettlementMethodInstant,
			},
		},
	}

	validPayments := []model.CardFundedPayment{
		{
			ID:              paymentID,
			ChargeID:        chargeID,
			MerchantID:      merchantID,
			ReferenceID:     referenceID,
			Currency:        constant.CurrencyIDR,
			Fee:             5000.00,
			Amount:          500_000.00,
			Sequence:        2,
			FirstPaymentID:  "first-payment-id",
			CardFingerprint: "fingerprint-abc",
		},
	}

	certPEM := []byte("-----BEGIN CERTIFICATE-----\ntest-public-key\n-----END CERTIFICATE-----") // NOSONAR
	encryptedPayload := "encrypted-payload-string"                                               // NOSONAR

	tests := []struct {
		name        string
		merchantID  string
		referenceID string
		setupMock   func()
		wantError   error
	}{
		{
			name:        "ERROR:GetDetailForCardFundedPayoutByID returns error",
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
			name:        "SUCCESS:Instant settlement with no pending subsequent payments",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				disbursementRepo.On(
					"GetDetailForCardFundedPayoutByID", mock.Anything, referenceID,
				).Return(validInstantPayout, nil)
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Once().Return([]model.CardFundedPayment{}, nil)
				log.On("Info", mock.Anything, fmt.Sprintf("Transaction with reference ID %s has no subsequent transactions", referenceID)).Once().Return()
			},
			wantError: nil,
		},
		{
			name:        "ERROR:FindPendingSubsequentCardFundedPayout returns error",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed when finding pending subsequent card-funded payment transactions", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:        "ERROR:GetCardEncryptionPublicKey returns error",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				paymentRepo.On(
					"FindPendingSubsequentCardFundedPayout", mock.Anything, merchantID, referenceID,
				).Return(validPayments, nil)
				creditCardSvc.On(
					"GetCardEncryptionPublicKey", mock.Anything, merchantID,
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:        "SUCCESS:Instant settlement with successful authentication",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				creditCardSvc.On(
					"GetCardEncryptionPublicKey", mock.Anything, merchantID,
				).Return(certPEM, nil)
				cryptoProvider.On("EncryptPKCS7", certPEM, mock.Anything).Once().Return(encryptedPayload, nil)
				creditCardSvc.On(
					"Authentication", mock.Anything, mock.MatchedBy(func(req creditcardCoreProcessorModel.AuthenticationRequest) bool {
						return req.MerchantID == merchantID && req.PaymentID == paymentID
					}),
				).Once().Return(&creditcardCoreProcessorModel.AuthenticationResponse{}, nil)
				log.On("Info", mock.Anything, "Processing subsequent payment for payment ID "+validPayments[0].ID, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:        "SUCCESS:Instant settlement with authentication failure - should fail payment and charge",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				cryptoProvider.On("EncryptPKCS7", certPEM, mock.Anything).Once().Return(encryptedPayload, nil)
				creditCardSvc.On("Authentication", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				paymentRepo.On(
					"UpdatePaymentStatusWithReason", mock.Anything, paymentID, mock.Anything,
				).Once().Return(nil)
				accountTransactionRepo.On(
					"UpdateStatusAccountTransaction", mock.Anything, chargeID, constant.StatusFailed, mock.Anything, mock.Anything,
				).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Processing subsequent payment for payment ID "+validPayments[0].ID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Failed when execute card authentication request", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:        "SUCCESS:Instant settlement with encryption failure - should skip payment",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				cryptoProvider.On(
					"EncryptPKCS7", certPEM, mock.Anything,
				).Once().Return("", assert.AnError)
				log.On("Info", mock.Anything, "Processing subsequent payment for payment ID "+validPayments[0].ID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Failed when encrypt request payload for card authentication", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:        "SUCCESS:Instant settlement with failed payment status update - should warn",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				cryptoProvider.On(
					"EncryptPKCS7", certPEM, mock.Anything,
				).Once().Return(encryptedPayload, nil)
				creditCardSvc.On(
					"Authentication", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
				paymentRepo.On(
					"UpdatePaymentStatusWithReason", mock.Anything, paymentID, mock.Anything,
				).Once().Return(assert.AnError)
				statusHistoryRepo.On(
					"Insert", mock.Anything, mock.Anything,
				).Once().Return(nil)
				log.On("Info", mock.Anything, "Processing subsequent payment for payment ID "+validPayments[0].ID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Failed when execute card authentication request", mock.Anything).Once().Return()
				log.On("Warn", mock.Anything, "Failed to update payment status with failure reason", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:        "SUCCESS:Instant settlement with failed charge status update - should warn",
			merchantID:  merchantID,
			referenceID: referenceID,
			setupMock: func() {
				cryptoProvider.On(
					"EncryptPKCS7", certPEM, mock.Anything,
				).Once().Return(encryptedPayload, nil)
				creditCardSvc.On(
					"Authentication", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
				paymentRepo.On(
					"UpdatePaymentStatusWithReason", mock.Anything, paymentID, mock.Anything,
				).Once().Return(nil)
				accountTransactionRepo.On(
					"UpdateStatusAccountTransaction", mock.Anything, chargeID, constant.StatusFailed, mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
				statusHistoryRepo.On(
					"Insert", mock.Anything, mock.Anything,
				).Once().Return(nil)
				log.On("Info", mock.Anything, "Processing subsequent payment for payment ID "+validPayments[0].ID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Failed when execute card authentication request", mock.Anything).Once().Return()
				log.On("Warn", mock.Anything, "Failed to update charge status", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := service.ProcessPendingSubsequentPayments(t.Context(), test.merchantID, test.referenceID)

			require.Equal(t, test.wantError, err)

			log.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			accountTransactionRepo.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
			creditCardSvc.AssertExpectations(t)
			cryptoProvider.AssertExpectations(t)
		})
	}
}
