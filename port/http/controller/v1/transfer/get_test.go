package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	transferModel "github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
)

func TestParseFilterParam(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		expected      transferModel.GetTransferListRequest
		expectedError bool
	}{
		{
			name:        "default values",
			queryParams: "",
			expected: transferModel.GetTransferListRequest{
				SortOrder: "ASC",
				SortBy:    "createdAt",
				Page:      1,
				PerPage:   10,
			},
			expectedError: false,
		},
		{
			name:        "valid page and perPage",
			queryParams: "page=2&perPage=20",
			expected: transferModel.GetTransferListRequest{
				SortOrder: "ASC",
				SortBy:    "createdAt",
				Page:      2,
				PerPage:   20,
			},
			expectedError: false,
		},
		{
			name:          "invalid page format",
			queryParams:   "page=abc",
			expected:      transferModel.GetTransferListRequest{},
			expectedError: true,
		},
		{
			name:          "invalid perPage format",
			queryParams:   "perPage=abc",
			expected:      transferModel.GetTransferListRequest{},
			expectedError: true,
		},
		{
			name:        "valid startDate and endDate",
			queryParams: "startDate=2023-01-01T00:00:00Z&endDate=2023-01-31T23:59:59Z",
			expected: transferModel.GetTransferListRequest{
				SortOrder: "ASC",
				SortBy:    "createdAt",
				Page:      1,
				PerPage:   10,
				StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
			},
			expectedError: false,
		},
		{
			name:          "invalid startDate format",
			queryParams:   "startDate=invalid-date",
			expected:      transferModel.GetTransferListRequest{},
			expectedError: true,
		},
		{
			name:          "invalid endDate format",
			queryParams:   "endDate=invalid-date",
			expected:      transferModel.GetTransferListRequest{},
			expectedError: true,
		},
		{
			name:        "valid sort and sortBy",
			queryParams: "sort=DESC&sortBy=amount",
			expected: transferModel.GetTransferListRequest{
				SortOrder: "DESC",
				SortBy:    "amount",
				Page:      1,
				PerPage:   10,
			},
			expectedError: false,
		},
		{
			name:        "valid status and uuid",
			queryParams: "status=completed&uuid=123e4567-e89b-12d3-a456-426614174000",
			expected: transferModel.GetTransferListRequest{
				SortOrder:          "ASC",
				SortBy:             "createdAt",
				Page:               1,
				PerPage:            10,
				Status:             "completed",
				UUID:               "123e4567-e89b-12d3-a456-426614174000",
				PaymentReferenceID: "123e4567-e89b-12d3-a456-426614174000",
				PaymentID:          "123e4567-e89b-12d3-a456-426614174000",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/transfers?%s", tt.queryParams), nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			controller := &TransferController{}
			result, err := controller.ParseFilterParam(req)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if !tt.expectedError && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestGetById(t *testing.T) {
	mockTransferSvc := serviceMocks.NewITransferService(t)
	mockMerchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(nil, nil, nil, mockTransferSvc, WithMerchantService(mockMerchantSvc))

	router := chi.NewRouter()
	router.Get("/transfers/{id}", handler.GetTransferByID)

	transferID := uuid.NewString()

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "valid-user-id",
		MerchantId: "valid-merchant-id",
	}

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		id              string
		param           string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR: when the user token not exist, then should return error 41",
			id:             transferID,
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR: when failed to get transfer detail then should return error",
			id:              transferID,
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusInternalServerError,
			setupMock: func() {
				mockTransferSvc.On("GetTransferTransaction", constant.ValueCtxMockType(), transferModel.GetTransferTransactionRequest{
					MerchantID:    userTokenClaims.MerchantId,
					TransactionID: transferID,
				}).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			wantRespBody: constant.WrapErrApiRespForTest(99, response.ErrTypeUnknown, "some error"),
		},
		{
			name:            "ERROR: when invalid transfer ID format, then should return error",
			id:              "invalid-uuid",
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "invalid transfer ID format"),
		},
		{
			name:            "SUCCESS: when success to get transfer detail, then should return the data",
			id:              transferID,
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusOK,
			setupMock: func() {
				mockTransferSvc.On("GetTransferTransaction", constant.ValueCtxMockType(), transferModel.GetTransferTransactionRequest{
					MerchantID:    userTokenClaims.MerchantId,
					TransactionID: transferID,
				}).Return(&transferModel.TransferTransactionDetail{UUID: "valid-transfer-id"}, nil).Once()
			},
			wantRespBody: `{"code":"00","message":"OK","data":{"paymentReferenceId":null,"amount":0,"createdAt":"0001-01-01T00:00:00Z","currency":"","feeAmount":0,"feeCurrency":null,"recipientId":"","recipientName":"","referenceId":"","remarks":"","senderId":"","senderName":"","status":"","paymentId":null,"type":"","uuid":"valid-transfer-id"}}`,
		},
		{
			name:            "SUCCESS: when success to get transfer detail by parent merchant, then should return the data",
			id:              transferID,
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusOK,
			param:           "subMerchantId=valid-sub-merchant-id",
			setupMock: func() {
				mockMerchantSvc.On("ValidateSubMerchantParent", constant.ValueCtxMockType(), userTokenClaims.MerchantId, "valid-sub-merchant-id").Return(nil).Once()
				mockTransferSvc.On("GetTransferTransaction", constant.ValueCtxMockType(), transferModel.GetTransferTransactionRequest{
					MerchantID:    "valid-sub-merchant-id",
					TransactionID: transferID,
					ParentID:      userTokenClaims.MerchantId,
				}).Return(&transferModel.TransferTransactionDetail{UUID: "valid-transfer-id"}, nil).Once()
			},
			wantRespBody: `{"code":"00","message":"OK","data":{"paymentReferenceId":null,"amount":0,"createdAt":"0001-01-01T00:00:00Z","currency":"","feeAmount":0,"feeCurrency":null,"recipientId":"","recipientName":"","referenceId":"","remarks":"","senderId":"","senderName":"","status":"","paymentId":null,"type":"","uuid":"valid-transfer-id"}}`,
		},
		{
			name:            "ERROR: when invalid parent merchant id, then should return error",
			id:              transferID,
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusInternalServerError,
			param:           "subMerchantId=invalid-sub-merchant-id",
			setupMock: func() {
				mockMerchantSvc.On("ValidateSubMerchantParent", constant.ValueCtxMockType(), userTokenClaims.MerchantId, "invalid-sub-merchant-id").Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			wantRespBody: `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/transfers/%s?%s", test.id, test.param), nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userTokenClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
func TestFilterTransferHistory(t *testing.T) {
	mockTransferSvc := serviceMocks.NewITransferService(t)

	handler := New(nil, nil, nil, mockTransferSvc)

	router := chi.NewRouter()
	router.Get("/transfers", handler.FilterTransferHistory)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "valid-user-id",
		MerchantId: "valid-merchant-id",
	}

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		queryParams     string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR: when the user token not exist, then should return error 41",
			queryParams:    "",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR: When the date is more than 6 months from now",
			queryParams:     "page=1&perPage=10&startDate=2025-01-01T17:53:00Z&endDate=2025-01-02T17:53:00Z",
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusBadRequest,
			setupMock:       func() {},
			wantRespBody:    constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months."),
		},
		{
			name:            "ERROR: when failed to get transfer list then should return error",
			queryParams:     "page=1&perPage=10",
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusInternalServerError,
			setupMock: func() {
				mockTransferSvc.On("GetList", constant.ValueCtxMockType(), &transferModel.GetTransferListRequest{
					SortOrder:  "ASC",
					SortBy:     "createdAt",
					Page:       1,
					PerPage:    10,
					MerchantID: userTokenClaims.MerchantId,
				}, int64(1), int64(10)).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			wantRespBody: constant.WrapErrApiRespForTest(99, response.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS: when success to get transfer list, then should return the data",
			queryParams:     "page=1&perPage=10",
			userTokenClaims: userTokenClaims,
			wantStatusCode:  http.StatusOK,
			setupMock: func() {
				mockTransferSvc.On("GetList", constant.ValueCtxMockType(), &transferModel.GetTransferListRequest{
					SortOrder:  "ASC",
					SortBy:     "createdAt",
					Page:       1,
					PerPage:    10,
					MerchantID: userTokenClaims.MerchantId,
				}, int64(1), int64(10)).Return(&commonModel.PaginationResponse{
					Data: []transferModel.TransferTransactionDetail{},
					Meta: commonModel.Meta{
						TotalItems: 0,
						TotalPages: 0,
						Page:       1,
						PerPage:    10,
					},
				}, nil).Once()
			},
			wantRespBody: `{"code":"00","message":"OK","data":[],"pagination":{"totalItems":0,"totalPages":0,"page":1,"perPage":10}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/transfers?%s", test.queryParams), nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userTokenClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
