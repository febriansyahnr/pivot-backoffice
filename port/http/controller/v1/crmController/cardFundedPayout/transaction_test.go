package crmCardFundedPayoutController_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/cardFundedPayout"
	"github.com/shopspring/decimal"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutTransactionList(t *testing.T) {
	validator := validatorExt.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	merchantID := uuid.NewString()
	startDate := "2026-04-01T00:00:00Z"
	endDate := "2026-04-02T23:59:59Z"

	handler := New(validator, service)

	router := chi.NewRouter()
	router.Get("/", handler.GetPayoutTransactionList)

	testCases := []struct {
		name           string
		queryParams    string
		setupMock      func()
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "ERROR: Missing required merchantId",
			queryParams:    "?startDate=" + startDate + "&endDate=" + endDate,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"MerchantID":"Key: 'GetPayoutTransactionListRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: Missing required startDate",
			queryParams:    "?merchantId=" + merchantID + "&endDate=" + endDate,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"StrStartDate":"Key: 'GetPayoutTransactionListRequest.StrStartDate' Error:Field validation for 'StrStartDate' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: Missing required endDate",
			queryParams:    "?merchantId=" + merchantID + "&startDate=" + startDate,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"StrEndDate":"Key: 'GetPayoutTransactionListRequest.StrEndDate' Error:Field validation for 'StrEndDate' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: Invalid merchantId format",
			queryParams:    "?merchantId=invalid-uuid&startDate=" + startDate + "&endDate=" + endDate,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"MerchantID":"Key: 'GetPayoutTransactionListRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'uuid' tag"}}`,
		},
		{
			name:           "ERROR: Invalid startDate format",
			queryParams:    "?merchantId=" + merchantID + "&startDate=invalid-date&endDate=" + endDate,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"StrStartDate":"Key: 'GetPayoutTransactionListRequest.StrStartDate' Error:Field validation for 'StrStartDate' failed on the 'datetime' tag"}}`,
		},
		{
			name:           "ERROR: Invalid endDate format",
			queryParams:    "?merchantId=" + merchantID + "&startDate=" + startDate + "&endDate=invalid-date",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"StrEndDate":"Key: 'GetPayoutTransactionListRequest.StrEndDate' Error:Field validation for 'StrEndDate' failed on the 'datetime' tag"}}`,
		},
		{
			name:           "ERROR: Invalid date range",
			queryParams:    "?merchantId=" + merchantID + "&startDate=" + endDate + "&endDate=" + startDate,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":"invalid date range"}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "?merchantId=" + merchantID + "&startDate=" + startDate + "&endDate=" + endDate,
			setupMock: func() {
				service.On("GetPayoutTransactionList", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"code":"99","errors":"assert.AnError general error for testing"}`,
		},
		{
			name:        "SUCCESS: Without optional parameters",
			queryParams: "?merchantId=" + merchantID + "&startDate=" + startDate + "&endDate=" + endDate,
			setupMock: func() {
				service.On("GetPayoutTransactionList", mock.Anything, mock.Anything).Once().Return([]cardFundedPayoutModel.GetPayoutTransactionListResponse{
					{
						ID:              "test-id-1",
						VendorID:        "vendor-1",
						VendorName:      "Vendor Name",
						BankCode:        "BCA",
						BankName:        "Bank BCA",
						AccountNumber:   "1234567890",
						AccountName:     "Account Name",
						Remarks:         "Test payout",
						TrxAmount:       decimal.NewFromFloat(100000),
						TrxStatus:       "SCHEDULED",
						CreatedAt:       time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
						ApprovedAt:      time.Date(2026, 4, 1, 10, 5, 0, 0, time.UTC),
						ScheduledAt:     util.ValueToPtr(time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)),
						MerchantID:      merchantID,
						InitAmount:      decimal.NewFromFloat(100000),
						InitFee:         decimal.NewFromFloat(5000),
						InitTotalAmount: decimal.NewFromFloat(105000),
						ExecAmount:      decimal.NewFromFloat(100000),
						ExecFee:         decimal.NewFromFloat(5000),
						ExecTotalAmount: decimal.NewFromFloat(105000),
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"code":"00","data":[{"id":"test-id-1","trxId":"","clientReferenceId":"","vendorId":"vendor-1","vendorName":"Vendor Name","bankCode":"BCA","bankName":"Bank BCA","accountNumber":"1234567890","accountName":"Account Name","remarks":"Test payout","trxAmount":"100000","trxStatus":"SCHEDULED","trxReasonType":null,"trxReasonDesc":null,"createdAt":"2026-04-01T10:00:00Z","approvedAt":"2026-04-01T10:05:00Z","scheduledAt":"2026-04-02T10:00:00Z","trxCreatedAt":null,"trxUpdatedAt":null,"merchantId":"` + merchantID + `","initAmount":"100000","initFee":"5000","initTotalAmount":"105000","execAmount":"100000","execFee":"5000","execTotalAmount":"105000"}]}`,
		},
		{
			name:        "SUCCESS: With all optional parameters",
			queryParams: "?merchantId=" + merchantID + "&startDate=" + startDate + "&endDate=" + endDate + "&trxStatus=PENDING&trxReasonType=INSUFFICIENT_BALANCE",
			setupMock: func() {
				service.On("GetPayoutTransactionList", mock.Anything, mock.Anything).Once().Return([]cardFundedPayoutModel.GetPayoutTransactionListResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"code":"00","data":[]}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+tt.queryParams, nil)

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}

func TestPatchPayoutTransactionStatus(t *testing.T) {
	validator := validatorExt.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	payoutID := uuid.NewString()

	handler := New(validator, service)

	router := chi.NewRouter()
	router.Patch("/{payoutId}/status", handler.PatchPayoutTransactionStatus)

	testCases := []struct {
		name           string
		payoutID       string
		body           string
		setupMock      func()
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "ERROR: Invalid JSON body",
			payoutID:       payoutID,
			body:           "invalid-json",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name:           "ERROR: Missing required status",
			payoutID:       payoutID,
			body:           `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"Status":"Key: 'PatchPayoutTransactionStatusRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: Invalid status value",
			payoutID:       payoutID,
			body:           `{"status":"INVALID"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":"40","errors":{"Status":"Key: 'PatchPayoutTransactionStatusRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"}}`,
		},
		{
			name:     "ERROR: Service error",
			payoutID: payoutID,
			body:     `{"status":"SUCCESS","bankReferenceNo":"REF001","reconReferenceNo":"RECON001"}`,
			setupMock: func() {
				service.On("UpdatePayoutTransactionStatus", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"code":"99","errors":"assert.AnError general error for testing"}`,
		},
		{
			name:     "SUCCESS: Update status to SUCCESS",
			payoutID: payoutID,
			body:     `{"status":"SUCCESS","bankReferenceNo":"REF001","reconReferenceNo":"RECON001"}`,
			setupMock: func() {
				service.On("UpdatePayoutTransactionStatus", mock.Anything, mock.MatchedBy(func(req cardFundedPayoutModel.PatchPayoutTransactionStatusRequest) bool {
					return req.PayoutID == payoutID &&
						req.Status == "SUCCESS" &&
						req.BankReferenceNo == "REF001" &&
						req.ReconReferenceNo == "RECON001"
				})).Once().Return(&cardFundedPayoutModel.PayoutActionResponse{
					ID:               payoutID,
					Status:           "SUCCESS",  // NOSONAR
					BankReferenceNo:  "REF001",   // NOSONAR
					ReconReferenceNo: "RECON001", // NOSONAR
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       fmt.Sprintf(`{"code":"00","data":{"id":"%s","vendorId":"","vendorName":"","referenceId":"","feeAmount":0,"amount":{"currency":"","value":0},"remarks":"","settlementMethod":"","cardId":"","cardName":"","status":"SUCCESS","bankReferenceNo":"REF001","reconReferenceNo":"RECON001"}}`, payoutID),
		},
		{
			name:     "SUCCESS: Update status to FAILED",
			payoutID: payoutID,
			body:     `{"status":"FAILED","reasonType":"INSUFFICIENT_BALANCE","reasonDescription":"Insufficient card balance"}`,
			setupMock: func() {
				reasonType := "INSUFFICIENT_BALANCE"
				reasonDesc := "Insufficient card balance"
				service.On("UpdatePayoutTransactionStatus", mock.Anything, mock.MatchedBy(func(req cardFundedPayoutModel.PatchPayoutTransactionStatusRequest) bool {
					return req.PayoutID == payoutID &&
						req.Status == "FAILED" &&
						req.ReasonType != nil && *req.ReasonType == reasonType &&
						req.ReasonDescription != nil && *req.ReasonDescription == reasonDesc
				})).Once().Return(&cardFundedPayoutModel.PayoutActionResponse{
					ID:                payoutID,
					Status:            "FAILED",
					ReasonType:        &reasonType,
					ReasonDescription: &reasonDesc,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       fmt.Sprintf(`{"code":"00","data":{"id":"%s","vendorId":"","vendorName":"","referenceId":"","feeAmount":0,"amount":{"currency":"","value":0},"remarks":"","settlementMethod":"","cardId":"","cardName":"","status":"FAILED","reasonType":"INSUFFICIENT_BALANCE","reasonDescription":"Insufficient card balance"}}`, payoutID),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/"+tt.payoutID+"/status", strings.NewReader(tt.body))

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
