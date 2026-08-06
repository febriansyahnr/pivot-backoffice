package paymentService_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	cryptoProviderMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	rdbMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/redis/go-redis/v9"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVCCTerminalBatchCharge(t *testing.T) {
	rmq := rmqMock.NewRabbitMQExt(t)
	logger := loggerMock.NewILogger(t)
	cardSvc := serviceMocks.NewICreditCardService(t)
	cryptoProvider := cryptoProviderMock.NewCryptoProvider(t)
	internal := serviceMocks.NewIPaymentInternalDirectFunc(t)
	paymentMethodRepo := repoMocks.NewIPaymentMethodRepository(t)
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)

	conf := &config.Config{
		VccTerminal: config.VccTerminalConfig{
			DefaultConfig: config.VCCTerminalDefaultConfig{
				AcquirerMerchantID: "TESTING123456",             // NOSONAR
				CardTypes:          []string{"DEBIT", "CREDIT"}, // NOSONAR
				PrincipalAvailable: []string{"VISA", "JCB"},     // NOSONAR
			},
			TravelAgents: config.MStrStr{
				"ABCD": "Abcd.id",  // NOSONAR
				"TEST": "Test.com", // NOSONAR
			},
			WorkerCount: 1, // NOSONAR
		},
	}
	service := New(
		nil, logger, nil, nil, nil, paymentMethodRepo, nil,
		WithConfig(conf),
		WithRabbitMQClient(rmq),
		WithCreditCardService(cardSvc),
		WithInternalDirectFunc(internal),
		WithValidator(validatorExt.New()),
		WithCryptoProvider(cryptoProvider),
	)
	WithUnifiedPaymentService(service, unifiedPaymentSvc)

	cardPaymentMethod := &model.PaymentMethodWithPivot{
		MerchantConfigObj: &model.PaymentMethodMerchantConfigObject{
			PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
					Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
						{
							IsActive:        true,
							TravelAgentCode: "TEST", // NOSONAR
							SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
								AllowedBinNumbers: []string{"444000"},
							},
							AcquirerMerchantID: "000000001", // NOSONAR
						},
					},
				},
			},
		},
	}

	certPEM := []byte(`-----BEGIN CERTIFICATE-----`)
	merchantID := "0f382eb3-7fdc-4e1e-8212-951184743b0a"
	createdBy := "2e2b5e5b-9b97-4c1b-9a40-11c53629f09e"
	booking := model.BookingPayload{
		BookingID:   "BOOK0001", // NOSONAR
		ReferenceID: "REFF0002", // NOSONAR
		Amount: model.Amount{
			Value:    decimal.NewFromInt(150_000),
			Currency: constant.CurrencyIDR,
		},
		TravelAgentCode: "TEST", // NOSONAR
		Card: model.Card{
			Number:       "4440000042200014", // NOSONAR
			SecurityCode: "123",              // NOSONAR
			Expiry: model.Expiry{
				Month: "12", // NOSONAR
				Year:  "30", // NOSONAR
			},
		},
		BankMerchantID: "000000001", // NOSONAR
		MerchantID:     merchantID,
		CreatedBy:      createdBy,
		ChargeID:       "a385fa65-e339-406b-81e7-cc01e51dc638",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *model.VCCTerminalBatchChargeResponse
	}{
		{
			name: "ERROR:Get active payment method virtual terminal",
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Payment method virtual terminal not found",
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrFeatureIsNotYetEnable),
		},
		{
			name: "ERROR:Decrypt request",
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(cardPaymentMethod, nil)
				internal.On("DecryptRequest", mock.Anything, mock.Anything, mock.Anything).Once().Return(pkgErrors.New(response.HttpErrInternal, assert.AnError))
			},
			wantError: pkgErrors.New(response.HttpErrInternal, assert.AnError),
		},
		{
			name: "ERROR:Empty booking transactions",
			setupMock: func() {
				internal.On("DecryptRequest", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("empty bookings transaction")),
		},
		{
			name: "ERROR:Required fields",
			setupMock: func() {
				internal.On("DecryptRequest", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					*args.Get(2).(*[]model.BookingPayload) = []model.BookingPayload{{}}
				}).Once().Return(nil)
			},
			wantError: pkgErrors.New(response.HttpErrRequest, func() error {
				return validatorExt.New().Struct(model.BookingPayload{})
			}()),
		},
		{
			name: "ERROR:Travel agent code not found",
			setupMock: func() {
				internal.On("DecryptRequest", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					result := booking
					result.TravelAgentCode = "XXXX"
					*args.Get(2).(*[]model.BookingPayload) = []model.BookingPayload{result}
				}).Once().Return(nil)
			},
			wantError: pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("travel agent code XXXX not found")),
		},
		{
			name: "ERROR:Card BIN is not allowed",
			setupMock: func() {
				cardPaymentMethod.MerchantConfigObj.PartnerConfig.Card.Items[0].SupportedUseCase.AllowedBinNumbers = []string{"555000"}

				internal.On("DecryptRequest", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					*args.Get(2).(*[]model.BookingPayload) = []model.BookingPayload{booking}
				}).Return(nil)
			},
			wantError: pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("card BIN is not allowed for travel agent code TEST or global config, masked card number 444000xxxxxx0014")),
		},
		{
			name: "ERROR:Get card encryption key",
			setupMock: func() {
				cardPaymentMethod.MerchantConfigObj.PartnerConfig.Card.Items[0].SupportedUseCase.AllowedBinNumbers = []string{"444000"}

				cardSvc.On("GetCardEncryptionPublicKey", mock.Anything, merchantID).Once().Return(nil, pkgErrors.New(response.HttpErrInternal, assert.AnError))
			},
			wantError: pkgErrors.New(response.HttpErrInternal, assert.AnError),
		},
		{
			name: "ERROR:Encrypt card charge",
			setupMock: func() {
				cardSvc.On("GetCardEncryptionPublicKey", mock.Anything, merchantID).Return(certPEM, nil)
				logger.On("Info", mock.Anything, "Processing VCC terminal transaction", mock.Anything).Once().Return()
				cryptoProvider.On("EncryptPKCS7", certPEM, mock.Anything).Once().Return("", assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to encrypt VCC terminal charge using PKCS#7", mock.Anything).Once().Return()
			},
			wantResult: &model.VCCTerminalBatchChargeResponse{
				FailedCount:   1,
				FailedTotal:   150_000,
				FailedCharges: []model.BookingPayload{booking.ToResponse()},
			},
		},
		{
			name: "ERROR:Create payment session",
			setupMock: func() {
				logger.On("Info", mock.Anything, "Processing VCC terminal transaction", mock.Anything).Once().Return()
				cryptoProvider.On("EncryptPKCS7", certPEM, mock.Anything).Return("encrypted", nil)
				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.MatchedBy(func(r *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) bool {
					return r.ClientReferenceID == booking.ReferenceID &&
						r.MerchantID == booking.MerchantID &&
						uuid.Validate(r.PaymentID) == nil &&
						r.Mode == constant.UnifiedPaymentModeRedirect &&
						r.VirtualTerminal != nil
				})).Once().Return(nil, assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to create payment session for booking transaction", mock.Anything).Once().Return()
			},
			wantResult: &model.VCCTerminalBatchChargeResponse{
				FailedCount:   1,
				FailedTotal:   150_000,
				FailedCharges: []model.BookingPayload{booking.ToResponse()},
			},
		},
		{
			name: "ERROR:Publish charge",
			setupMock: func() {
				logger.On("Info", mock.Anything, "Processing VCC terminal transaction", mock.Anything).Once().Return()
				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.MatchedBy(func(r *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) bool {
					return r.ClientReferenceID == booking.ReferenceID &&
						r.MerchantID == booking.MerchantID &&
						uuid.Validate(r.PaymentID) == nil &&
						r.Mode == constant.UnifiedPaymentModeRedirect &&
						r.VirtualTerminal != nil
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{ChargeDetails: []*unifiedPaymentModel.ChargeResponse{{ID: booking.ChargeID}}}, nil)
				rmq.On("Publish", mock.Anything, rabbitMqExt.VccTerminalChargeRoutingKey, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to publish booking transaction", mock.Anything).Once().Return()
			},
			wantResult: &model.VCCTerminalBatchChargeResponse{
				FailedCount:   1,
				FailedTotal:   150_000,
				FailedCharges: []model.BookingPayload{booking.ToResponse()},
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				logger.On("Info", mock.Anything, "Processing VCC terminal transaction", mock.Anything).Once().Return()
				rmq.On("Publish", mock.Anything, rabbitMqExt.VccTerminalChargeRoutingKey, mock.Anything, mock.MatchedBy(func(payload model.VCCTerminalChargeMessage) bool {
					return payload.MerchantID == merchantID &&
						uuid.Validate(payload.PaymentID) == nil &&
						uuid.Validate(payload.BatchID) == nil &&
						payload.ChargeID == booking.ChargeID
				})).Once().Return(nil)
			},
			wantResult: &model.VCCTerminalBatchChargeResponse{
				SuccessCount: 1,
				SuccessTotal: 150_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.VCCTerminalBatchCharge(t.Context(), &model.VCCTerminalChargeRequest{
				MerchantID: merchantID,
				UserID:     createdBy,
			})
			assert.Equal(t, test.wantError, err)

			if test.wantResult != nil && result != nil {

				test.wantResult.BatchID = result.BatchID

				if len(test.wantResult.FailedCharges) > 0 && len(result.FailedCharges) > 0 {
					test.wantResult.FailedCharges[0].BatchID = result.FailedCharges[0].BatchID
					test.wantResult.FailedCharges[0].PaymentID = result.FailedCharges[0].PaymentID
				}
			}
			resultJson, _ := json.Marshal(result)
			wantResultJson, _ := json.Marshal(test.wantResult)
			assert.JSONEq(t, string(wantResultJson), string(resultJson))

			rmq.AssertExpectations(t)
			logger.AssertExpectations(t)
			cardSvc.AssertExpectations(t)
			internal.AssertExpectations(t)
			cryptoProvider.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
			unifiedPaymentSvc.AssertExpectations(t)
		})
	}
}

func TestVCCTerminalSubmitCharge(t *testing.T) {
	rdb := rdbMock.NewIRedisExt(t)
	logger := loggerMock.NewILogger(t)
	cardSvc := serviceMocks.NewICreditCardService(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	statusHistoriesRepo := repoMocks.NewIStatusHistoriesRepository(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)

	service := New(
		paymentRepo, logger, nil, nil, nil, nil, nil,
		WithRedisClient(rdb),
		WithCreditCardService(cardSvc),
		WithStatusHistoriesRepository(statusHistoriesRepo),
		WithAccountTransactionRepository(accountTransactionRepo),
	)

	request := paymentModel.VCCTerminalChargeMessage{
		PaymentID: "19e6d3f6-5f59-4c25-9c9c-019b0463ac2e",
	}
	chargeKey := fmt.Sprintf(constant.VCCTerminalSubmitChargeCacheKey, request.PaymentID)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: Set transaction lock", // NOSONAR
			setupMock: func() {
				result := &redis.BoolCmd{}
				result.SetErr(assert.AnError)
				rdb.On("SetNX", mock.Anything, chargeKey, true, constant.VCCTerminalSubmitChargeTTL).Once().Return(result)
			},
			wantError: fmt.Errorf("set vcc terminal lock: %w", assert.AnError),
		},
		{
			name: "ERROR: Transaction has been processed", // NOSONAR
			setupMock: func() {
				result := &redis.BoolCmd{}
				result.SetVal(false)
				rdb.On("SetNX", mock.Anything, chargeKey, true, constant.VCCTerminalSubmitChargeTTL).Once().Return(result)
			},
			wantError: pkgErrors.NewNonRetryableError(errors.New("vcc terminal transaction has been processed")),
		},
		{
			name: "ERROR: Get payment by id", // NOSONAR
			setupMock: func() {
				result := &redis.BoolCmd{}
				result.SetVal(true)
				rdb.On("SetNX", mock.Anything, chargeKey, true, constant.VCCTerminalSubmitChargeTTL).Return(result)

				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(nil, assert.AnError)
				resultDel := &redis.IntCmd{}
				resultDel.SetErr(assert.AnError)
				rdb.On("Del", mock.Anything, chargeKey).Once().Return(resultDel)
				logger.On("Error", mock.Anything, "Failed to delete VCC terminal lock for payment ID "+request.PaymentID, mock.Anything).Once().Return()
			},
			wantError: fmt.Errorf("get payment by id: %w", assert.AnError),
		},
		{
			name: "ERROR: Transaction not found", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(nil, nil)
				result := &redis.IntCmd{}
				result.SetErr(nil)
				rdb.On("Del", mock.Anything, chargeKey).Once().Return(result)
			},
			wantError: pkgErrors.NewNonRetryableError(errors.New("vcc terminal transaction not found")),
		},
		{
			name: "ERROR: Transaction is already in final status", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(&model.Payment{
					Status: constant.UnifiedPaymentSessionStatusPaid,
				}, nil)
				result := &redis.IntCmd{}
				result.SetErr(nil)
				rdb.On("Del", mock.Anything, chargeKey).Once().Return(result)
			},
			wantError: pkgErrors.NewNonRetryableError(errors.New("vcc terminal transaction is already in final status")),
		},
		{
			name: "ERROR: Authentication fail with error update payment status", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Return(&model.Payment{
					Status: constant.UnifiedPaymentSessionStatusProcessing,
				}, nil)
				cardSvc.On("Authentication", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				paymentRepo.On("UpdatePaymentStatusWithReason", mock.Anything, request.PaymentID, mock.MatchedBy(func(p model.UpdatePaymentStatusWithReasonRequest) bool {
					return p.Status == constant.UnifiedPaymentSessionStatusCancelled &&
						*p.ReasonType == constant.CreditCardAuthorizationFailed &&
						*p.ReasonDescription == "Failed to process authorization request" // NOSONAR
				})).Once().Return(assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to update payment status with reason", mock.Anything).Once().Return()
				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				result := &redis.IntCmd{}
				result.SetErr(nil)
				rdb.On("Del", mock.Anything, chargeKey).Once().Return(result)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: Authentication fail with error update charge status", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Return(&model.Payment{
					Status: constant.UnifiedPaymentSessionStatusProcessing,
				}, nil)
				cardSvc.On("Authentication", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				paymentRepo.On("UpdatePaymentStatusWithReason", mock.Anything, request.PaymentID, mock.MatchedBy(func(p model.UpdatePaymentStatusWithReasonRequest) bool {
					return p.Status == constant.UnifiedPaymentSessionStatusCancelled &&
						*p.ReasonType == constant.CreditCardAuthorizationFailed &&
						*p.ReasonDescription == "Failed to process authorization request" // NOSONAR
				})).Once().Return(nil)
				accountTransactionRepo.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to update charge status", mock.Anything).Once().Return()
				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				result := &redis.IntCmd{}
				result.SetErr(nil)
				rdb.On("Del", mock.Anything, chargeKey).Once().Return(result)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				cardSvc.On("Authentication", mock.Anything, mock.Anything).Once().Return(&creditcardCoreProcessorModel.AuthenticationResponse{}, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, service.VCCTerminalSubmitCharge(t.Context(), request))

			rdb.AssertExpectations(t)
			logger.AssertExpectations(t)
			cardSvc.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetVCCTerminalList(t *testing.T) {
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	logger := loggerMock.NewILogger(t)

	service := New(paymentRepo, logger, nil, nil, nil, nil, nil)

	tests := []struct {
		name      string
		request   *paymentModel.GetVCCTerminalListFilterRequest
		setupMock func()
		wantErr   bool
	}{
		{
			name:    "ERROR: repository returns error",
			request: &paymentModel.GetVCCTerminalListFilterRequest{},
			setupMock: func() {
				paymentRepo.On("GetVCCTerminalList", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "SUCCESS: repository returns valid response",
			request: &paymentModel.GetVCCTerminalListFilterRequest{},
			setupMock: func() {
				expectedResponse := &commonModel.PaginationResponse{
					Data: []paymentModel.VccTerminalItem{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 0,
						TotalPages: 0,
					},
				}
				paymentRepo.On("GetVCCTerminalList", mock.Anything, mock.Anything).Once().Return(expectedResponse, nil)
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			response, err := service.GetVCCTerminalList(t.Context(), test.request)

			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
