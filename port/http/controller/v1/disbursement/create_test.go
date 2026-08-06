package disbursementController

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestControllerCreateSingle(t *testing.T) {
	merchantId := "b60b4c2e-d9a3-4fed-8b26-04050023ec59"
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: merchantId,
		Role:       constant.RoleMaker,
	}

	validPayload := disbursementModel.CreateSingleRequest{
		ReferenceID:            "client-ref",
		BeneficiaryBankCode:    "008",
		BeneficiaryBankName:    "Bank Permata",
		BeneficiaryAccountNo:   "8000800808",
		BeneficiaryAccountName: "Yories Yolanda",
		Amount:                 decimal.NewFromInt(100000),
		Remark:                 strings.Repeat("A", 40),
	}
	rawValidPayload, err := json.Marshal(validPayload)
	require.NoError(t, err)

	merchantData := &merchantModel.Merchant{UUID: merchantId}
	merchantConfigs := &merchantModel.MerchantIdForConfigs{
		MerchantTransactionConfig: merchantId,
	}
	ctx := context.WithValue(context.Background(), constant.CtxMerchantData, merchantData)

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10_000,
		MaxAmount: 100_000,
	}

	testCases := []struct {
		name           string
		mockSetup      func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, beneficiarySvc *mocks.IBeneficiaryAccountService)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				disbursementSvc.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(trxConfig, nil)
				merchantSvc.On(
					"GetMerchantIdForConfigs", constant.ValueCtxMockType(), validUserClaims.MerchantId, true,
				).Return(ctx, merchantConfigs, nil)

				disbursementSvc.
					On(
						"CreateSingle",
						constant.ValueCtxMockType(),
						mock.AnythingOfType("*disbursementModel.CreateSingleRequest")).
					Return(&disbursementModel.Disbursement{}, nil)

				disbursementSvc.
					On(
						"IsExistReferenceID",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(false)

				disbursementSvc.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "008", constant.StringMockType()).
					Return(false)
			},
			setupBody:      func(*testing.T) []byte { return rawValidPayload },
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:User not in Context",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				// Empty function body
			},
			setupBody: func(t *testing.T) []byte {
				return []byte{}
			},
			expectedStatus: http.StatusUnauthorized,
			userClaim:      nil,
		},
		{
			name: "ERROR:Invalid payload",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				// Empty function body
			},
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:Remark is too long",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				// Empty function body
			},
			setupBody: func(t *testing.T) []byte {
				return []byte(`{"beneficiaryAccountNo":"00000001","beneficiaryBankCode":"001","referenceId":"ABC124","beneficiaryBankName":"BANK TEST","beneficiaryAccountName":"DUMMY SIMULATION","amount":10000,"remark":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}`)
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:Missing required payload",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				// Empty function body
			},
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(disbursementModel.CreateSingleRequest{})
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:Using magic number",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				// Empty function body
			},
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.CreateSingleRequest{
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Permata",
					BeneficiaryAccountNo:   "999966660004",
					BeneficiaryAccountName: "Yories Yolanda",
					ReferenceID:            "1234567890",
					Amount:                 decimal.NewFromInt(100000),
					Remark:                 "This is remark",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:Get merchant id for configs",
			mockSetup: func(_ *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, true).Return(ctx, nil, constant.ErrSomeErrorForUnitTest)
			},
			setupBody:      func(*testing.T) []byte { return rawValidPayload },
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:Get transaction config",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, true).Return(ctx, merchantConfigs, nil)
				disbursementSvc.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			setupBody:      func(*testing.T) []byte { return rawValidPayload },
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:Amount less than 10.000",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, beneficiarySvc *mocks.IBeneficiaryAccountService) {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, true).Return(ctx, merchantConfigs, nil)
				disbursementSvc.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(trxConfig, nil)
				disbursementSvc.On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "008", constant.StringMockType()).Return(false)
			},
			setupBody: func(t *testing.T) []byte {
				buf, err := json.Marshal(disbursementModel.CreateSingleRequest{
					ReferenceID:            "client-ref",
					BeneficiaryBankCode:    "008",
					BeneficiaryBankName:    "Bank Permata",
					BeneficiaryAccountNo:   "8000800808",
					BeneficiaryAccountName: "Yories Yolanda",
					Amount:                 decimal.NewFromInt(9999),
					Remark:                 "This is remark",
				})
				assert.NoError(t, err)
				return buf
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:Amount more than maximum amount for custom rule",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, beneficiarySvc *mocks.IBeneficiaryAccountService) {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, true).Return(ctx, merchantConfigs, nil)
				disbursementSvc.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(trxConfig, nil)

				disbursementSvc.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "002", constant.StringMockType()).
					Return(true)

				beneficiarySvc.On("FindByBankCodeAndAccountNo",
					constant.ValueCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest")).
					Return(&beneficiaryAccountModel.Account{
						MetadataObj: beneficiaryAccountModel.Metadata{
							BeneficiaryPayoutLimitRule: &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
								Velocity:        10,
								AmountThreshold: 1_000_000_000.00,
							},
						},
					}, nil)

				disbursementSvc.On(
					"IsMerchantAllowedExcludeBeneficiaryRules", constant.ValueCtxMockType(), constant.StringMockType(), constant.Float64MockType(),
				).Return(0.0, false)
				disbursementSvc.On(
					"IsMerchantAllowedToUseBeneficiaryCustomRule", constant.ValueCtxMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(true)
			},
			setupBody: func(t *testing.T) []byte {
				buf, err := json.Marshal(disbursementModel.CreateSingleRequest{
					ReferenceID:            "client-ref2",
					BeneficiaryBankCode:    "002",
					BeneficiaryBankName:    "Bank Permata",
					BeneficiaryAccountNo:   "8000800809",
					BeneficiaryAccountName: "John Wick",
					Amount:                 decimal.NewFromInt(100_500),
					Remark:                 "Empty",
				})
				assert.NoError(t, err)
				return buf
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:CreateSingle validation error",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				disbursementSvc.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(trxConfig, nil)
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, true).Return(ctx, merchantConfigs, nil)

				disbursementSvc.
					On(
						"IsExistReferenceID",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(false)

				disbursementSvc.
					On(
						"CreateSingle",
						constant.ValueCtxMockType(),
						mock.AnythingOfType("*disbursementModel.CreateSingleRequest")).
					Return(nil, pkgErrs.New(response.HttpErrRequest, constant.ErrSomeErrorForUnitTest))

				disbursementSvc.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "008", constant.StringMockType()).
					Return(false)
			},
			setupBody:      func(*testing.T) []byte { return rawValidPayload },
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR:CreateSingle internal error",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService, merchantSvc *mocks.IMerchantService, _ *mocks.IBeneficiaryAccountService) {
				disbursementSvc.On(
					"GetTransactionConfig", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(trxConfig, nil)
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, true).Return(ctx, merchantConfigs, nil)

				disbursementSvc.
					On(
						"IsExistReferenceID",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(false)

				disbursementSvc.
					On(
						"CreateSingle",
						constant.ValueCtxMockType(),
						mock.AnythingOfType("*disbursementModel.CreateSingleRequest")).
					Return(nil, constant.ErrSomeErrorForUnitTest)

				disbursementSvc.
					On("IsBankcodeOverbookingChannelAllowed", mock.Anything, "008", constant.StringMockType()).
					Return(false)
			},
			setupBody:      func(*testing.T) []byte { return rawValidPayload },
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
				Environment: constant.EnvironmentStaging,
			}

			mockValidator := validator.New()
			disbursementSvc := mocks.NewIDisbursementService(t)
			merchantSvc := mocks.NewIMerchantService(t)
			beneficiaryAccountSvc := mocks.NewIBeneficiaryAccountService(t)

			tt.mockSetup(disbursementSvc, merchantSvc, beneficiaryAccountSvc)

			// Statsd Monitoring
			monitor, err := monitoring.New("backend-portal", "0.0.0.0", "1234")
			if err != nil {
				fmt.Printf("Unable to init monitoring, %v", err)
				panic(err)
			}

			mc := New(cfg, mockValidator, monitor, Services{MerchantSvc: merchantSvc, DisbursementSvc: disbursementSvc, BeneficiaryAccountSvc: beneficiaryAccountSvc}, nil, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/api/v1/disbursements/single/create", bytes.NewBuffer(tt.setupBody(t)))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.CreateSingle)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			disbursementSvc.AssertExpectations(t)
			merchantSvc.AssertExpectations(t)
		})
	}
}
