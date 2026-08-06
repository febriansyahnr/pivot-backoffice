package beneficiaryAccountController

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestControllerCheckBeneficiary(t *testing.T) {

	feeSvc := mocks.NewIFeeService(t)
	merchantSvc := mocks.NewIMerchantService(t)

	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	validAccount := &beneficiaryAccountModel.Account{
		UUID:                   "uuid-uuid-uuid",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "12345678",
		BeneficiaryAccountName: "test",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "testing",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	payloadRequest := &beneficiaryAccountModel.CheckAccountRequest{
		BeneficiaryAccountNo: "12345678",
		BeneficiaryBankCode:  "1234",
		MerchantID:           "merchant-id",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)
	payloadRequestBankCode008 := &beneficiaryAccountModel.CheckAccountRequest{
		BeneficiaryAccountNo: "12345678",
		BeneficiaryBankCode:  "008",
		MerchantID:           "merchant-id",
	}
	payloadRequestBankCode008Byte, err := json.Marshal(payloadRequestBankCode008)
	assert.NoError(t, err)

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10_000, MaxAmount: 110_000,
	}
	ctxValue := context.WithValue(context.Background(), constant.CtxTraceIdKey, uuid.NewString())
	showAdditionalInfo := true
	hideAdditionalInfo := false

	testCases := []struct {
		name                   string
		requestBody            []byte
		mockSetup              func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus         int
		userClaim              *user.UserTokenClaims
		expectAdditionalInfo   *bool
		expectedVirtualAccount bool
	}{
		{
			name: "ERROR: User not in Context",
			mockSetup: func(_ *mocks.IBeneficiaryAccountService, _ *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				// No Body Function
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(_ *mocks.IBeneficiaryAccountService, _ *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				// No Body Function
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(_ *mocks.IBeneficiaryAccountService, _ *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				// No Body Function
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "INVALID: Special characters",
			requestBody: []byte(`{"BeneficiaryAccountNo":"12345678!.,.","BeneficiaryBankCode":"1234"}`),
			mockSetup: func(_ *mocks.IBeneficiaryAccountService, _ *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				// No Body Function
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Find by bank code and account no",
			requestBody: payloadRequestByte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, _ *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				ben.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Get transaction config",
			requestBody: payloadRequestByte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				ben.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(validAccount, nil)
				dis.On("GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Get merchant id for configs",
			requestBody: payloadRequestByte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				ben.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(validAccount, nil)
				dis.On("GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything).Return(trxConfig, nil)
				dis.On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "1234", constant.StringMockType()).Return(false, nil)
				merchantSvc.On(
					"GetMerchantIdForConfigs", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(ctxValue, nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Get payout transaction fee",
			requestBody: payloadRequestByte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, _ *mockRabbitMq.RabbitMQExt) {
				ben.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(validAccount, nil)
				dis.On("GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything).Return(trxConfig, nil)
				dis.On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "1234", constant.StringMockType()).Return(false, nil)
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, mock.Anything, mock.Anything).Return(ctxValue, &merchant.MerchantIdForConfigs{}, nil)
				feeSvc.On(
					"GetPayoutTransactionFee", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "VALID: Numbers only",
			requestBody: []byte(`{"BeneficiaryAccountNo":"12345678","BeneficiaryBankCode":"1234"}`),
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				feeSvc.On("GetPayoutTransactionFee", mock.Anything, mock.Anything).Return(&feeModel.FeeMetadataObject{}, nil)
				ben.
					On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).
					Return(validAccount, nil)
				dis.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything,
				).Return(trxConfig, nil)
				dis.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "1234", constant.StringMockType()).
					Return(true, nil)
				dis.
					On("IsMerchantAllowedToUseBeneficiaryCustomRule", constant.ValueCtxMockType(), constant.StringMockType(), constant.BoolMockType()).
					Return(true)
				rabbitMqMock.
					On("PublishActivity", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything,
						mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				dis.
					On("IsMerchantAllowedExcludeBeneficiaryRules", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("float64")).
					Return(float64(math.MaxInt), false)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name:        "VALID: Space is allowed",
			requestBody: []byte(`{"BeneficiaryAccountNo":"1234 5678","BeneficiaryBankCode":"1234"}`),
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				ben.
					On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).
					Return(validAccount, nil)
				dis.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything,
				).Return(trxConfig, nil)
				dis.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "1234", constant.StringMockType()).
					Return(true, nil)
				dis.
					On("IsMerchantAllowedToUseBeneficiaryCustomRule", constant.ValueCtxMockType(), constant.StringMockType(), constant.BoolMockType()).
					Return(true)
				dis.
					On("IsMerchantAllowedExcludeBeneficiaryRules", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("float64")).
					Return(float64(math.MaxInt), false)
				rabbitMqMock.
					On("PublishActivity", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything,
						mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				ben.
					On(
						"FindByBankCodeAndAccountNo",
						mock.Anything,
						mock.Anything).
					Return(validAccount, nil)
				dis.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything,
				).Return(trxConfig, nil)

				dis.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "1234", constant.StringMockType()).
					Return(true, nil)

				dis.
					On("IsMerchantAllowedToUseBeneficiaryCustomRule", constant.ValueCtxMockType(), constant.StringMockType(), constant.BoolMockType()).
					Return(true)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				dis.
					On("IsMerchantAllowedExcludeBeneficiaryRules", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("float64")).
					Return(float64(math.MaxInt), false)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:        "SUCCESS with empty account name (inquiry-in-progress)",
			requestBody: payloadRequestByte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				pendingAccount := *validAccount
				pendingAccount.BeneficiaryAccountName = ""
				ben.
					On(
						"FindByBankCodeAndAccountNo",
						mock.Anything,
						mock.Anything).
					Return(&pendingAccount, nil)
				dis.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), mock.Anything,
				).Return(trxConfig, nil)

				dis.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "1234", constant.StringMockType()).
					Return(true, nil)

				dis.
					On("IsMerchantAllowedToUseBeneficiaryCustomRule", constant.ValueCtxMockType(), constant.StringMockType(), constant.BoolMockType()).
					Return(true)

				dis.
					On("IsMerchantAllowedExcludeBeneficiaryRules", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("float64")).
					Return(float64(math.MaxInt), false)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 202,
			userClaim:      validUserClaims,
		},
		{
			name:        "SUCCESS with additional info when feature flag enabled",
			requestBody: payloadRequestBankCode008Byte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				ben.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(&beneficiaryAccountModel.Account{
					UUID:                   "beneficiary-id",
					MerchantID:             "merchant-flag-on",
					BeneficiaryAccountNo:   "12345678",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Name",
					MetadataObj: beneficiaryAccountModel.Metadata{
						IsVirtualAccount: true,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
				dis.On("GetTransactionConfig", mock.Anything, "merchant-flag-on").Return(trxConfig, nil)
				dis.On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "008", "merchant-flag-on").Return(false, nil)
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, "merchant-flag-on", false).Return(context.Background(), &merchant.MerchantIdForConfigs{}, nil)
				feeSvc.On("GetPayoutTransactionFee", mock.Anything, mock.Anything).Return(&feeModel.FeeMetadataObject{}, nil)
				rabbitMqMock.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedStatus:         http.StatusOK,
			userClaim:              &user.UserTokenClaims{UUID: "user-id", MerchantId: "merchant-flag-on"},
			expectAdditionalInfo:   &showAdditionalInfo,
			expectedVirtualAccount: true,
		},
		{
			name:        "SUCCESS without additional info when feature flag disabled",
			requestBody: payloadRequestBankCode008Byte,
			mockSetup: func(ben *mocks.IBeneficiaryAccountService, dis *mocks.IDisbursementService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				ben.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(&beneficiaryAccountModel.Account{
					UUID:                   "beneficiary-id",
					MerchantID:             "merchant-flag-off",
					BeneficiaryAccountNo:   "12345678",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Name",
					MetadataObj: beneficiaryAccountModel.Metadata{
						IsVirtualAccount: true,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
				dis.On("GetTransactionConfig", mock.Anything, "merchant-flag-off").Return(trxConfig, nil)
				dis.On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "008", "merchant-flag-off").Return(false, nil)
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, "merchant-flag-off", false).Return(context.Background(), &merchant.MerchantIdForConfigs{}, nil)
				feeSvc.On("GetPayoutTransactionFee", mock.Anything, mock.Anything).Return(&feeModel.FeeMetadataObject{}, nil)
				rabbitMqMock.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedStatus:       http.StatusOK,
			userClaim:            &user.UserTokenClaims{UUID: "user-id", MerchantId: "merchant-flag-off"},
			expectAdditionalInfo: &hideAdditionalInfo,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockBen := mocks.NewIBeneficiaryAccountService(t)
			mockDis := mocks.NewIDisbursementService(t)
			validate := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			cfg := &config.Config{
				ServiceName: "testing",
			}

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
			f, ffErr := os.CreateTemp(os.TempDir(), "account-inquiry-display-virtual-account-flag-for-whitelisted-merchant-*.yaml")
			require.NoError(t, ffErr)
			defer func() { require.NoError(t, os.Remove(f.Name())) }()
			defer func() { require.NoError(t, f.Close()) }()

			_, ffErr = f.WriteString(ffContentConfig)
			require.NoError(t, ffErr)

			ffErr = ffclient.Init(ffclient.Config{
				FileFormat: "YAML",
				Retriever: &fileretriever.Retriever{
					Path: f.Name(),
				},
			})
			require.NoError(t, ffErr)
			defer ffclient.Close()

			tt.mockSetup(mockBen, mockDis, mockRmq)

			mc := New(cfg, validate, mockRmq, mockBen, mockDis, merchantSvc, feeSvc)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/beneficiary-accounts/inquiry", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.CheckBeneficiary)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectAdditionalInfo != nil {
				if *tt.expectAdditionalInfo {
					if tt.expectedVirtualAccount {
						assert.Contains(t, rr.Body.String(), `"additionalInfo":{"isVirtualAccount":true}`)
					} else {
						assert.Contains(t, rr.Body.String(), `"additionalInfo":{"isVirtualAccount":false}`)
					}
				} else {
					assert.NotContains(t, rr.Body.String(), `"additionalInfo"`)
				}
			}
			mockBen.AssertExpectations(t)
		})
	}
}
