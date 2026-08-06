package unifiedPaymentService_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrepareRecurringPaymentRequest(t *testing.T) {
	log := logMock.NewILogger(t)
	redisExt := redisMock.NewIRedisExt(t)
	recurringContractRepo := repoMocks.NewIRecurringContractRepository(t)

	service := New(nil, log, nil, nil, nil, WithRecurringContractRepository(recurringContractRepo), WithRedisClient(redisExt))

	customerID := "customer-123"                          // NOSONAR
	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"  // NOSONAR
	recurringID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b" // NOSONAR
	paymentTokenID := "af02d570-50de-4a1f-b541-77bf31642abf"
	processFirstAuthKey := fmt.Sprintf(constant.RecurringPaymentMutualExclusionKey, constant.RecurringPaymentTypeFirstAuthorization, recurringID)
	processSubsequentPaymentKey := fmt.Sprintf(constant.RecurringPaymentMutualExclusionKey, constant.RecurringPaymentTypeSubsequentPayment, recurringID)

	redisSuccessResult := &redis.BoolCmd{}
	redisSuccessResult.SetVal(true)

	tests := []struct {
		name            string
		request         *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest
		setupMock       func()
		wantError       error
		validateRequest func(t *testing.T, req *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest)
	}{
		{
			name:      "SUCCESS:Empty recurring ID",
			request:   &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{},
			wantError: nil,
		},
		{
			name: "ERROR:Database error when retrieving recurring contract",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			setupMock: func() {
				recurringContractRepo.On("GetDetailByID", mock.Anything, merchantID, recurringID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed when retrieving recurring payment contract", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Recurring contract not found",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			setupMock: func() {
				recurringContractRepo.On("GetDetailByID", mock.Anything, merchantID, recurringID).Once().Return(nil, nil)
			},
			wantError: constant.NewErrResourceNotFound("recurring payment contract", recurringID),
		},
		{
			name: "ERROR:First authorization must be completed",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:                 merchantID,
				RecurringID:                recurringID,
				InitiateFirstAuthorization: false,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			setupMock: func() {
				recurringContractRepo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(&recurringContractModel.RecurringContractDetail{
					Status: constant.RecurringContractStatusCreated,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrRequest, fmt.Errorf("%s", "The first authorization must be completed before the next subsequent transaction")),
		},
		{
			name: "ERROR:Recurring contract inactive",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			setupMock: func() {
				recurringContractRepo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(&recurringContractModel.RecurringContractDetail{
					Status: constant.RecurringContractStatusInactive,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrRequest, fmt.Errorf("%s", "The recurring payment contract is no longer active")),
		},
		{
			name: "ERROR:Amount mismatch",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				Amount: unifiedPaymentModel.Amount{
					Value:    50000,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			setupMock: func() {
				recurringContractRepo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Return(&recurringContractModel.RecurringContractDetail{
					UUID:              recurringID,
					MerchantID:        merchantID,
					CustomerID:        customerID,
					AuthMethod:        constant.RecurringContractAuthMethodFirstPayment,
					PaymentTokenID:    &paymentTokenID,
					Status:            constant.RecurringContractStatusActive,
					PaymentMethodType: util.ValueToPtr("CREDIT_CARD"),
					Billing: recurringContractModel.Billing{
						Interval:     1,
						IntervalUnit: "MONTH",
						Count:        1,
					},
					Amount: 100000,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrRequest, fmt.Errorf("%s", "The transaction amount does not match the billing cycle calculation")),
		},
		{
			name: "ERROR:Acquire recurring payment lock",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				Amount: unifiedPaymentModel.Amount{
					Value:    0,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				InitiateFirstAuthorization: true,
				ExpiryAt:                   time.Now().UTC().Add(time.Minute),
			},
			setupMock: func() {
				redisResult := &redis.BoolCmd{}
				redisResult.SetErr(assert.AnError)
				redisExt.On("SetNX", mock.Anything, processFirstAuthKey, true, mock.Anything).Once().Return(redisResult)
				log.On("Error", mock.Anything, "Failed to acquire recurring payment mutex lock", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Recurring payment is currently being processed",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				Amount: unifiedPaymentModel.Amount{
					Value:    0,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				InitiateFirstAuthorization: true,
				ExpiryAt:                   time.Now().UTC().Add(time.Minute),
			},
			setupMock: func() {
				redisResult := &redis.BoolCmd{}
				redisExt.On("SetNX", mock.Anything, processFirstAuthKey, true, mock.Anything).Once().Return(redisResult)
			},
			wantError: constant.NewErrStringRequest(response.HttpErrConflict, constant.ErrCodeDuplicateError, "The recurring payment is currently being processed"),
		},
		{
			name: "SUCCESS:Initiate for change of payment method",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				Amount: unifiedPaymentModel.Amount{
					Value:    0,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				InitiateFirstAuthorization: true,
				ExpiryAt:                   time.Now().UTC().Add(time.Minute),
			},
			setupMock: func() {
				log.On(
					"Info", mock.Anything, "Exclusive lock acquired for recurring ID "+recurringID+" type "+constant.RecurringPaymentTypeFirstAuthorization,
				).Once().Return()
				redisExt.On("SetNX", mock.Anything, processFirstAuthKey, true, mock.Anything).Once().Return(redisSuccessResult)
			},
			wantError: nil,
			validateRequest: func(t *testing.T, req *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
				require.True(t, *req.SaveForFutureUse)
				require.Equal(t, 10_000.0, req.Amount.Value)
				require.Equal(t, customerID, req.CustomerID)
				require.Equal(t, uint16(0), req.RecurringBillingCycle.Count)
				require.Equal(t, constant.RecurringContractAuthMethodOneDollar, req.FirstAuthorizationMethod)
				require.Equal(t, constant.CardThreeDsMethodChallenge, req.PaymentMethodOptions.Card.ThreeDsMethod)
				require.Equal(t, constant.UnifiedPaymentCardCaptureMethodManual, req.PaymentMethodOptions.Card.CaptureMethod)
				require.NotNil(t, req.CleanupPreparedRecurringPaymentLock)

				// Testing when cleanup fails
				log.On(
					"Info", mock.Anything, "Cleanup prepared recurring payment for recurring ID "+recurringID+" type "+constant.RecurringPaymentTypeFirstAuthorization,
				).Once().Return()

				redisResult := &redis.IntCmd{}
				redisResult.SetErr(assert.AnError)
				redisExt.On("Del", mock.Anything, processFirstAuthKey).Once().Return(redisResult)
				log.On("Error", mock.Anything, "Failed to cleanup prepared recurring payment lock", mock.Anything).Once().Return()

				req.CleanupPreparedRecurringPaymentLock(t.Context())
			},
		},
		{
			name: "SUCCESS:With correct amount",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:  merchantID,
				RecurringID: recurringID,
				Amount: unifiedPaymentModel.Amount{
					Value:    100000,
					Currency: constant.CurrencyIDR,
				},
			},
			setupMock: func() {
				log.On(
					"Info", mock.Anything, "Exclusive lock acquired for recurring ID "+recurringID+" type "+constant.RecurringPaymentTypeSubsequentPayment,
				).Once().Return()
				redisExt.On("SetNX", mock.Anything, processSubsequentPaymentKey, true, mock.Anything).Once().Return(redisSuccessResult)
			},
			wantError: nil,
			validateRequest: func(t *testing.T, req *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
				require.Equal(t, 100000.0, req.Amount.Value)
				require.Equal(t, customerID, req.CustomerID)
				require.Empty(t, req.FirstAuthorizationMethod)
				require.Equal(t, constant.UnifiedPaymentMethodCard, req.PaymentMethod.Type)
				require.Equal(t, paymentTokenID, req.PaymentMethod.CardPaymentMethodDetail.Token)
				require.Equal(t, uint16(2), req.RecurringBillingCycle.Count)
				require.Equal(t, constant.CardThreeDsMethodNever, req.PaymentMethodOptions.Card.ThreeDsMethod)
				require.Empty(t, req.PaymentMethodOptions.Card.CaptureMethod)
				require.NotNil(t, req.CleanupPreparedRecurringPaymentLock)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			err := service.(port.IInternalUnifiedPaymentService).PrepareRecurringPaymentRequest(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if err == nil && test.validateRequest != nil {
				test.validateRequest(t, test.request)
			}

			log.AssertExpectations(t)
			redisExt.AssertExpectations(t)
			recurringContractRepo.AssertExpectations(t)
		})
	}
}
