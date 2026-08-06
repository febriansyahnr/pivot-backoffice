package payment_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChannelList(t *testing.T) {
	paymentMethodService := serviceMocks.NewIPaymentMethodService(t)

	handler := New(nil, validator.New(), nil, WithPaymentMethodService(paymentMethodService))

	router := chi.NewRouter()
	router.Get("/payments/channels", handler.GetChannelList)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "000c4096-e92e-4f59-a0fe-bab1fd53b1c9",
		MerchantId: "23bd62a9-239f-4d90-a6b0-6bf22fdec793",
	}

	tests := []struct {
		name                  string
		userTokenClaims       *user.UserTokenClaims
		withDerivedMerchantID bool
		setupMock             func()
		wantStatusCode        int
		wantRespBody          string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR:Some error",
			userTokenClaims: userTokenClaims,
			setupMock: func() {
				paymentMethodService.On(
					"GetPaymentMethodByMerchant", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS",
			userTokenClaims: userTokenClaims,
			setupMock: func() {
				paymentMethodService.On(
					"GetPaymentMethodByMerchant", c.ValueCtxMockType(), mock.Anything,
				).Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID: "method-1",
							Name: "Credit Card",
						},
					},
				}, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"uuid":"method-1","type":"","category":"","name":"Credit Card","description":null,"logo":null,"acquirer":"","bankName":null,"instructions":null,"processor":"","activationMethod":"","countryOfOperation":"","supportedCurrency":"","requiredDocuments":null,"subtype":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","deletedAt":null,"merchantId":"","isActive":false,"activationStatus":"","channelType":"","merchantConfig":null}]}`,
		},
		{
			name:                  "SUCCESS With Derived Merchant ID",
			userTokenClaims:       userTokenClaims,
			withDerivedMerchantID: true,
			setupMock: func() {
				paymentMethodService.On(
					"GetPaymentMethodByMerchant", c.ValueCtxMockType(), mock.Anything,
				).Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID: "method-1",
							Name: "Credit Card",
						},
						IsDerivedMerchant: util.ValueToPtr(true),
					},
				}, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"uuid":"method-1","type":"","category":"","name":"Credit Card","description":null,"logo":null,"acquirer":"","bankName":null,"instructions":null,"processor":"","activationMethod":"","countryOfOperation":"","supportedCurrency":"","requiredDocuments":null,"subtype":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","deletedAt":null,"merchantId":"","isActive":false,"activationStatus":"","channelType":"","merchantConfig":null,"isDerivedMerchant":true}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments/channels", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userTokenClaims))
			}

			if test.withDerivedMerchantID {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxDerivedMerchantID, c.EmptyUUID))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestGetChannelDocuments(t *testing.T) {
	paymentMethodService := serviceMocks.NewIPaymentMethodService(t)

	handler := New(nil, validator.New(), nil, WithPaymentMethodService(paymentMethodService))

	router := chi.NewRouter()
	router.Get("/payments/channels/{paymentMethodId}/documents", handler.GetChannelDocuments)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "000c4096-e92e-4f59-a0fe-bab1fd53b1c9",
		MerchantId: "23bd62a9-239f-4d90-a6b0-6bf22fdec793",
	}

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		paymentMethodId string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR:Some error",
			userTokenClaims: userTokenClaims,
			paymentMethodId: "method-1",
			setupMock: func() {
				paymentMethodService.On(
					"GetRequiredMerchantDocuments", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS",
			userTokenClaims: userTokenClaims,
			paymentMethodId: "method-1",
			setupMock: func() {
				paymentMethodService.On(
					"GetRequiredMerchantDocuments", c.ValueCtxMockType(), mock.Anything,
				).Return(&[]paymentMethodModel.MerchantRequiredDocumentsResponse{
					{
						Name:                   "document1",
						Format:                 "pdf",
						MerchantDocumentID:     "doc-1",
						MerchantDocumentStatus: "pending",
					},
					{
						Name:                   "document2",
						Format:                 "jpg",
						MerchantDocumentID:     "doc-2",
						MerchantDocumentStatus: "approved",
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"name":"document1","format":"pdf","merchantDocumentID":"doc-1","merchantDocumentStatus":"pending"},{"name":"document2","format":"jpg","merchantDocumentID":"doc-2","merchantDocumentStatus":"approved"}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			url := "/payments/channels/" + test.paymentMethodId + "/documents"
			req := httptest.NewRequest(http.MethodGet, url, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userTokenClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestGetChannelListWithPaymentToken(t *testing.T) {
	paymentService := serviceMocks.NewIPaymentService(t)
	merchantService := serviceMocks.NewIMerchantService(t)
	paymentMethodService := serviceMocks.NewIPaymentMethodService(t)

	handler := New(nil, validator.New(), nil,
		WithPaymentService(paymentService),
		WithMerchantService(merchantService),
		WithPaymentMethodService(paymentMethodService))

	router := chi.NewRouter()
	router.Get("/payments/channels", handler.GetChannelListWithPaymentToken)

	tests := []struct {
		name           string
		paymentID      string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Payment ID not found in context",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "invalid token"),
		},
		{
			name:           "ERROR: Empty payment ID in context",
			paymentID:      "",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "invalid token"),
		},
		{
			name:      "ERROR: GetPaymentDetailForPaymentUI fails",
			paymentID: "payment-123",
			setupMock: func() {
				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "ERROR: FindMerchantByID fails",
			paymentID: "payment-123",
			setupMock: func() {
				paymentDetail := &paymentModel.PaymentDetailForPaymentUIResponse{
					MerchantID: "merchant-123",
				}
				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Once().Return(paymentDetail, nil)

				merchantService.On(
					"FindMerchantByID", c.ValueCtxMockType(), "merchant-123",
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "ERROR: GetPaymentMethodByMerchant fails",
			paymentID: "payment-123",
			setupMock: func() {
				paymentDetail := &paymentModel.PaymentDetailForPaymentUIResponse{
					MerchantID: "merchant-123",
				}
				merchant := &merchantModel.Merchant{
					UUID:      "merchant-123",
					KYCStatus: sql.NullString{String: c.KYCStatusApproved, Valid: true},
				}

				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Once().Return(paymentDetail, nil)

				merchantService.On(
					"FindMerchantByID", c.ValueCtxMockType(), "merchant-123",
				).Once().Return(merchant, nil)

				paymentMethodService.On(
					"GetPaymentMethodByMerchant", c.ValueCtxMockType(), &paymentModel.GetPaymentMethodFilterRequest{
						MerchantID: "merchant-123",
						Category:   c.TypePayment,
						Payment:    paymentDetail,
					},
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS: Regular merchant (KYC Approved)",
			paymentID: "payment-123",
			setupMock: func() {
				paymentDetail := &paymentModel.PaymentDetailForPaymentUIResponse{
					MerchantID: "merchant-123",
				}
				merchant := &merchantModel.Merchant{
					UUID:      "merchant-123",
					KYCStatus: sql.NullString{String: c.KYCStatusApproved, Valid: true},
				}

				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Once().Return(paymentDetail, nil)

				merchantService.On(
					"FindMerchantByID", c.ValueCtxMockType(), "merchant-123",
				).Once().Return(merchant, nil)

				paymentMethodService.On(
					"GetPaymentMethodByMerchant", c.ValueCtxMockType(), &paymentModel.GetPaymentMethodFilterRequest{
						MerchantID: "merchant-123",
						Category:   c.TypePayment,
						Payment:    paymentDetail,
					},
				).Once().Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID: "method-1",
							Name: "Credit Card",
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"uuid":"method-1","type":"","category":"","name":"Credit Card","description":null,"logo":null,"acquirer":"","bankName":null,"instructions":null,"processor":"","activationMethod":"","countryOfOperation":"","supportedCurrency":"","requiredDocuments":null,"subtype":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","deletedAt":null,"merchantId":"","isActive":false,"activationStatus":"","channelType":"","merchantConfig":null}]}`,
		},
		{
			name:      "SUCCESS: Sub-merchant (KYC not required, uses parent ID)",
			paymentID: "payment-123",
			setupMock: func() {
				paymentDetail := &paymentModel.PaymentDetailForPaymentUIResponse{
					MerchantID: "merchant-123",
				}
				merchant := &merchantModel.Merchant{
					UUID:      "merchant-123",
					KYCStatus: sql.NullString{String: c.KYCStatusNotRequired, Valid: true},
					ParentID:  sql.NullString{String: "parent-merchant-456", Valid: true},
				}

				paymentService.On(
					"GetPaymentDetailForPaymentUI", c.ValueCtxMockType(), "payment-123",
				).Once().Return(paymentDetail, nil)

				merchantService.On(
					"FindMerchantByID", c.ValueCtxMockType(), "merchant-123",
				).Once().Return(merchant, nil)

				paymentMethodService.On(
					"GetPaymentMethodByMerchant", c.ValueCtxMockType(), &paymentModel.GetPaymentMethodFilterRequest{
						MerchantID: "parent-merchant-456",
						Category:   c.TypePayment,
						Payment:    paymentDetail,
					},
				).Once().Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID: "method-1",
							Name: "Virtual Account",
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"uuid":"method-1","type":"","category":"","name":"Virtual Account","description":null,"logo":null,"acquirer":"","bankName":null,"instructions":null,"processor":"","activationMethod":"","countryOfOperation":"","supportedCurrency":"","requiredDocuments":null,"subtype":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","deletedAt":null,"merchantId":"","isActive":false,"activationStatus":"","channelType":"","merchantConfig":null}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments/channels", nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			if test.paymentID != "" {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxPaymentID, test.paymentID))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
