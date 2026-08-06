package internalAccountInquiry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mock_service "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckStatusRequestInquiry(t *testing.T) {
	testCases := []struct {
		desc           string
		expectedStatus int
		expectedBody   string
		mockSetup      func(svc *mock_service.IAccountInquiryService, httpReq *http.Request) *http.Request
	}{
		{
			desc:           "error when inquiry id is not valid uuid",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(svc *mock_service.IAccountInquiryService, httpReq *http.Request) *http.Request {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("inquiryId", "invalid-uuid")
				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, chiCtx))

				return httpReq
			},
		},
		{
			desc:           "error when merchant info not found",
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func(svc *mock_service.IAccountInquiryService, httpReq *http.Request) *http.Request {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("inquiryId", uuid.NewString())
				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, chiCtx))

				return httpReq
			},
		},
		{
			desc:           "error when CheckStatusRequestInquiry",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func(svc *mock_service.IAccountInquiryService, httpReq *http.Request) *http.Request {

				svc.On("CheckStatusRequestInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, assert.AnError)

				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("inquiryId", uuid.NewString())
				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, chiCtx))
				merchantInfo := merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				}

				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), constant.CtxMerchantInfo, &merchantInfo))
				return httpReq
			},
		},
		{
			desc:           "success CheckStatusRequestInquiry",
			expectedStatus: http.StatusOK,
			mockSetup: func(svc *mock_service.IAccountInquiryService, httpReq *http.Request) *http.Request {

				svc.On("CheckStatusRequestInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiriesHttpResponse{}, nil)

				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("inquiryId", uuid.NewString())
				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, chiCtx))
				merchantInfo := merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				}

				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), constant.CtxMerchantInfo, &merchantInfo))
				return httpReq
			},
		},
		{
			desc:           "success CheckStatusRequestInquiry with additional info",
			expectedStatus: http.StatusOK,
			expectedBody:   `"additionalInfo":{"isVirtualAccount":true}`,
			mockSetup: func(svc *mock_service.IAccountInquiryService, httpReq *http.Request) *http.Request {

				svc.On("CheckStatusRequestInquiry",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiriesHttpResponse{
					AdditionalInfo: &requestAccountInquiries.AccountInquiryResultAdditionalInfo{
						IsVirtualAccount: true,
					},
				}, nil)

				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("inquiryId", uuid.NewString())
				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, chiCtx))
				merchantInfo := merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				}

				httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), constant.CtxMerchantInfo, &merchantInfo))
				return httpReq
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			svc := mock_service.NewIAccountInquiryService(t)

			baseUrl := "/api/internal/v1/inquiry-account/{inquiryId}"
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)

			req = tc.mockSetup(svc, req)

			ctrl := New(svc)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.CheckStatusRequestInquiry)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, httpRecorder.Body.String(), tc.expectedBody)
			}
			svc.AssertExpectations(t)
		})
	}
}
