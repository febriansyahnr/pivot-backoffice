package disbursementController

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	accountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/account"
	accountInquriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/accountInquiries"
	accounttransactionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/accountTransaction"
	beneficiaryAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/beneficiaryAccount"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	snapCoreRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/snapCore"
	userRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/user"
	accountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/account"
	beneficiaryAccountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/beneficiaryAccount"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	merchantService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	orchestratorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	jwtCore "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/test/schemas"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	data := make([]disbursementModel.Disbursement, 0)
	data = append(data, disbursementModel.Disbursement{
		UUID: uuid.NewString(),
	})
	expectedResponse := &commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}
	validUserClaims := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name            string
		mockSetup       func(mockService *serviceMocks.IDisbursementService)
		expectedStatus  int
		funcQueryParams func() *url.Values
		userClaims      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List with filter bulk id",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("bulkId", "uuid")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
				queryParams.Add("endCreatedAt", util.TimeNow.Format(util.UTCLayout))
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "2")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 500 on get list caused by service error",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("some-error"))
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid startCreatedAt",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid endCreatedAt",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("endCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid perPage",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("perPage", "invalid format perPage")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			userClaims: nil,
			funcQueryParams: func() *url.Values {
				return nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIDisbursementService(t)
			tc.mockSetup(mockService)

			cfg := &config.Config{}
			cfg.AppConfig.PaginationPerPage = 20
			mc := New(cfg, nil, nil, Services{DisbursementSvc: mockService}, nil, nil)
			baseUrl := "/api/v1/disbursements"

			// Append query parameters to the URL
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req.Header.Set("Time-Zone", "Asia/Jakarta")
			if tc.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaims))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mc.GetList)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestIntegrationGetList(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	ctx := context.Background()
	cfg := &config.Config{
		Environment: constant.EnvironmentStaging,
	}
	scr := &config.Secret{}

	// Statsd Monitoring
	monitor, err := monitoring.New("backend-portal", "0.0.0.0", "1234")
	require.NoErrorf(t, err, "logger.New")

	// Setup JWT
	jwtConfig, _ := jwtCore.New(cfg, scr, cacheClient)

	// Setup GCS
	gcsConfig := gcs.Config{
		ServiceBucketName:          cfg.GCSConfig.ServiceBucketName,
		ReportingBucketName:        cfg.GCSConfig.ReportingBucketName,
		BulkDisbursementBucketName: cfg.GCSConfig.BulkDisbursementBucketName,
		ProofOfTransferFolderName:  cfg.GCSConfig.ProofOfTransferFolderName,
	}
	gcs := gcs.NewGCSService(gcsConfig)
	defer gcs.Close()

	// Init service
	accTrxRepo := accounttransactionRepository.New(db, pdkLoggerMock)
	accRepo := accountRepository.New(db, pdkLoggerMock)
	accSvc := accountService.New(pdkLoggerMock, accTrxRepo, accRepo, nil)
	merchantSvc := merchantService.New(
		merchantRepository.New(db, pdkLoggerMock),
		pdkLoggerMock,
		userRepository.New(db, pdkLoggerMock),
		jwtConfig,
		publisher,
		nil,
		merchantService.WithAccountService(accSvc),
	)
	orchestratorSvc := orchestratorService.New(pdkLoggerMock, gcs, accTrxRepo, accRepo)
	beneficiaryAccountSvc := beneficiaryAccountService.New(pdkLoggerMock,
		beneficiaryAccountRepository.New(db, pdkLoggerMock),
		accountInquriesRepository.New(db, pdkLoggerMock),
		snapCoreRepository.New(cfg, scr, pdkLoggerMock, httpRequestExt.New()),
	)
	disbursementSvc := disbursementService.New(
		cfg,
		pdkLoggerMock,
		merchantRepository.New(db, pdkLoggerMock),
		disbursementRepository.New(db, pdkLoggerMock),
		snapCoreRepository.New(cfg, scr, pdkLoggerMock, httpRequestExt.New()),
		nil,
		disbursementService.WithOrchestratorService(orchestratorSvc),
		disbursementService.WithBeneficiaryAccService(beneficiaryAccountSvc),
		disbursementService.WithRabbitMQClient(publisher),
		disbursementService.WithGoogleCloudStorage(gcs),
	)

	// Prepare Data
	userID := uuid.NewString()
	merchantID := uuid.NewString()

	userClaims := &user.UserTokenClaims{
		UUID:       userID,
		MerchantId: merchantID,
	}

	schemaMigrations := []string{
		schemas.Disbursement(),
		schemas.AccountTransaction(),
		schemas.User(),
	}
	for _, raw := range schemaMigrations {
		_, err := db.ExecContext(ctx, raw)
		require.NoErrorf(t, err, "run.schema-migration")
	}

	bankName := "BANK RAKYAT INDONESIA"
	amount := decimal.NewFromInt(100000)
	fee := decimal.NewFromFloat(constant.DefaultMerchantFee)
	totalAmount := amount.Add(fee)
	createdFrom := constant.DisbursementCreatedFromMerchantPortal
	requestDisbursement := &disbursementModel.Disbursement{
		UUID:                   uuid.NewString(),
		ReferenceID:            fmt.Sprintf("ref-%d", time.Now().Unix()),
		MerchantID:             merchantID,
		BulkID:                 nil,
		PurposeID:              nil,
		SenderName:             "Sender Name",
		BeneficiaryBankCode:    "002",
		BeneficiaryBankName:    &bankName,
		BeneficiaryAccountNo:   "888801000157508",
		BeneficiaryAccountName: "Dummy",
		ProcessorReferenceID:   nil,
		BankReferenceNo:        nil,
		Currency:               "IDR",
		Amount:                 amount,
		Fee:                    &fee,
		TotalAmount:            totalAmount,
		Status:                 constant.DisbursementStatusApproved,
		ReasonType:             nil,
		ReasonDescription:      nil,
		Remark:                 nil,
		CreatedFrom:            &createdFrom,
		CreatedBy:              &userID,
		ApprovedBy:             &userID,
		ApprovedAt:             &util.TimeNow,
		CreatedAt:              util.TimeNow,
		UpdatedAt:              util.TimeNow,
		DeletedAt:              nil,
	}
	_, err = db.NamedExecContext(
		ctx,
		`INSERT INTO disbursements
		(
			uuid,
		 	reference_id,
			merchant_id,
			bulk_id,
			purpose_id,
			sender_name,
			beneficiary_bank_code,
			beneficiary_bank_name,
			beneficiary_account_no,
			beneficiary_account_name,
			processor_reference_id,
			currency,
			amount,
			fee,
			total_amount,
			status,
			reason_type,
			reason_description,
		 	remark,
			created_from,
			created_by,
			approved_by,
			approved_at,
			created_at,
			updated_at,
			deleted_at
		) VALUES (
		    :uuid,
		    :reference_id,
			:merchant_id,
			:bulk_id,
			:purpose_id,
			:sender_name,
			:beneficiary_bank_code,
			:beneficiary_bank_name,
			:beneficiary_account_no,
			:beneficiary_account_name,
			:processor_reference_id,
			:currency,
			:amount,
			:fee,
			:total_amount,
			:status,
			:reason_type,
			:reason_description,
		    :remark,
			:created_from,
			:created_by,
			:approved_by,
			:approved_at,
			:created_at,
			:updated_at,
			:deleted_at
		)`, requestDisbursement,
	)
	require.NoErrorf(t, err, "insert.bulk_disbursements")

	ctrl := New(
		cfg,
		nil,
		monitor,
		Services{
			MerchantSvc:              merchantSvc,
			DisbursementDashboardSvc: nil,
			DisbursementSvc:          disbursementSvc,
			BeneficiaryAccountSvc:    nil,
		},
		publisher,
		gcs,
	)

	testCase := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "SUCCESS",
			expectedStatus: 200,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/disbursements", nil)
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims))

			rr := httptest.NewRecorder()

			// get disbursement
			handler := http.HandlerFunc(ctrl.GetList)

			// Serve the request
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
