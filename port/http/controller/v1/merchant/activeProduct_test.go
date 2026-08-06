package merchant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantActiveProducts(t *testing.T) {
	expectations := []*product.MerchantWithProductName{
		{
			ProductID:   "valid-uuid",
			ProductName: "Product Name",
			Active:      true,
		},
	}

	mockProductSvc := mockMerchant.NewIProductService(t)
	controller := New(nil, nil, nil, WithProductService(mockProductSvc))

	router := chi.NewRouter()
	router.Get("/merchants/actived-products", controller.GetActiveProducts)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	testCase := []struct {
		name             string
		userClaim        *user.UserTokenClaims
		mockSetup        func()
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name:      "SUCCESS",
			userClaim: userTokenClaims,
			mockSetup: func() {
				mockProductSvc.On("GetMerchantActiveProducts",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectations, nil).Once()
			},
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00","message":"OK","data":[{"productId":"valid-uuid","productName":"Product Name","active":true,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:      "ERROR: Service Error",
			userClaim: userTokenClaims,
			mockSetup: func() {
				mockProductSvc.On("GetMerchantActiveProducts",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: When the user claim is nil",
			mockSetup: func() {
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/merchants/actived-products", nil)
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userClaim))
			}
			rctx := chi.NewRouteContext()

			rr := httptest.NewRecorder()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			router.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tt.expectedRespBody, rr.Body.String()) {
				t.Log("Result:", rr.Body.String())
			}
		})
	}
}
