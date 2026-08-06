package internalAccountInquiry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mock_service "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequestAccountInquiry(t *testing.T) {
	var accountInquirySvc *mock_service.IAccountInquiryService

	validPayload := requestAccountInquiries.RequestAccountInquiriesHttpRequest{
		ChannelCode: "008",
		ChannelInformation: requestAccountInquiries.ChannelInformation{
			AccountName:   "ardabilli",
			AccountNumber: "1234567890",
		},
	}
	testCases := []struct {
		desc           string
		expectedStatus int
		expectedBody   string
		claimMerchant  bool
		mockSetup      func()
		reqPayload     func() []byte
		setHeaders     func(req *http.Request)
	}{
		{
			desc: "got error 401 merchantInfo not found",
			reqPayload: func() []byte {
				res, _ := json.Marshal(validPayload)
				return res
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// Empty mock setup
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:           "got error 400 EOF",
			claimMerchant:  true,
			expectedStatus: http.StatusBadRequest,
			reqPayload: func() []byte {
				res, _ := json.Marshal("")
				return res
			},
			mockSetup: func() {
				// Empty mock setup
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:          "got error 400 when invalid payload",
			claimMerchant: true,
			reqPayload: func() []byte {
				invalidPayload := validPayload
				invalidPayload.ChannelCode = ""

				res, err := json.Marshal(invalidPayload)
				if err != nil {
					t.Log(err)
				}

				return res
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// Empty mock setup
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:          "got error 400 when failed to request account inquiry",
			claimMerchant: true,
			reqPayload: func() []byte {
				res, err := json.Marshal(validPayload)
				if err != nil {
					t.Log(err)
				}

				return res
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				accountInquirySvc.On("RequestAccountInquiry", mock.Anything, mock.Anything).Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("error business logic"))).Once()
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:          "success 200 request account inquiry",
			claimMerchant: true,
			reqPayload: func() []byte {
				res, err := json.Marshal(validPayload)
				if err != nil {
					t.Log(err)
				}

				return res
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				accountInquirySvc.On("RequestAccountInquiry", mock.Anything, mock.Anything).Return(&requestAccountInquiries.RequestAccountInquiriesHttpResponse{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					InquiryResult: requestAccountInquiries.InquiryResult{
						Status: constant.RequestAccountInquiryStatusValid,
						Detail: "success",
					},
				}, nil)
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:          "success 200 request account inquiry with header submerchant",
			claimMerchant: true,
			reqPayload: func() []byte {
				res, err := json.Marshal(validPayload)
				if err != nil {
					t.Log(err)
				}

				return res
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				accountInquirySvc.On("RequestAccountInquiry", mock.Anything, mock.Anything).Return(&requestAccountInquiries.RequestAccountInquiriesHttpResponse{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					InquiryResult: requestAccountInquiries.InquiryResult{
						Status: constant.RequestAccountInquiryStatusValid,
						Detail: "success",
					},
				}, nil)
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
		},
		{
			desc:          "success 200 request account inquiry with additional info",
			claimMerchant: true,
			reqPayload: func() []byte {
				res, err := json.Marshal(validPayload)
				if err != nil {
					t.Log(err)
				}

				return res
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"additionalInfo":{"isVirtualAccount":true}`,
			mockSetup: func() {
				accountInquirySvc.On("RequestAccountInquiry", mock.Anything, mock.Anything).Return(&requestAccountInquiries.RequestAccountInquiriesHttpResponse{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					InquiryResult: requestAccountInquiries.InquiryResult{
						Status: constant.RequestAccountInquiryStatusValid,
						Detail: "success",
					},
					AdditionalInfo: &requestAccountInquiries.AccountInquiryResultAdditionalInfo{
						IsVirtualAccount: true,
					},
				}, nil)
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:          "got error 400 when special characters exist in AccountNumber",
			claimMerchant: true,
			reqPayload: func() []byte {
				payload := validPayload
				payload.ChannelInformation.AccountNumber = "1234!@#"
				res, _ := json.Marshal(payload)
				return res
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No specific mock setup needed
			},
			setHeaders: func(req *http.Request) {},
		},
		{
			desc:          "got success 200 when spaces in AccountNumber are trimmed",
			claimMerchant: true,
			reqPayload: func() []byte {
				payload := validPayload
				payload.ChannelInformation.AccountNumber = "1234 5678"
				res, _ := json.Marshal(payload)
				return res
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				accountInquirySvc.On("RequestAccountInquiry", mock.Anything, mock.Anything).
					Return(&requestAccountInquiries.RequestAccountInquiriesHttpResponse{
						UUID:       uuid.NewString(),
						MerchantID: uuid.NewString(),
						InquiryResult: requestAccountInquiries.InquiryResult{
							Status: constant.RequestAccountInquiryStatusValid,
							Detail: "success",
						},
					}, nil)
			},
			setHeaders: func(req *http.Request) {},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			accountInquirySvc = mock_service.NewIAccountInquiryService(t)
			ctrl := New(accountInquirySvc)
			ctx := context.Background()
			tc.mockSetup()

			baseUrl := "/api/internal/v1/inquiry-account"
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewReader(tc.reqPayload()))
			chiRouterCtx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(ctx)

			if tc.claimMerchant {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
				req = req.WithContext(ctx)
			}

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.RequestAccountInquiry)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, httpRecorder.Body.String(), tc.expectedBody)
			}
			accountInquirySvc.AssertExpectations(t)
		})
	}
}
