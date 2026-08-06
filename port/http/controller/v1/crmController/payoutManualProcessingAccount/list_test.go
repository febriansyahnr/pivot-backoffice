package payoutManualProcessingAccount_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	payoutManualProcessingAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payoutManualProcessingAccount"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCRMPayoutManualProcessingAccountController_List(t *testing.T) {
	merchantID := uuid.New().String()

	tests := []struct {
		name           string
		query          string
		modifierMock   func(svc *serviceMocks.IPayoutManualProcessingAccountService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:  "ERROR: Bad Request - Invalid merchant id",
			query: "merchantId=invalid-uuid",
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid merchant id", "error":{"type":"API_ERROR","message":"invalid merchant id","recommendation":""}, "data":null}`,
		},
		{
			name:  "ERROR: Bad Request - Invalid status",
			query: "merchantId=" + merchantID + "&status=INVALID",
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid status, must be ACTIVE or INACTIVE", "error":{"type":"API_ERROR","message":"invalid status, must be ACTIVE or INACTIVE","recommendation":""}, "data":null}`,
		},
		{
			name:  "ERROR: Bad Request - Invalid page",
			query: "merchantId=" + merchantID + "&page=0",
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid page number", "error":{"type":"API_ERROR","message":"invalid page number","recommendation":""}, "data":null}`,
		},
		{
			name:  "ERROR: Bad Request - Invalid perPage",
			query: "merchantId=" + merchantID + "&perPage=0",
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid per page number", "error":{"type":"API_ERROR","message":"invalid per page number","recommendation":""}, "data":null}`,
		},
		{
			name:  "ERROR: Service error",
			query: "merchantId=" + merchantID + "&status=ACTIVE",
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				svc.On("List", constant.ValueCtxMockType(), mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "message":"some error", "error":{"type":"UNKNOWN","message":"some error","recommendation":""}, "data":null}`,
		},
		{
			name:  "SUCCESS - List payout manual processing accounts",
			query: "merchantId=" + merchantID + "&status=ACTIVE&bankCode=BCA&accountNumber=123456&sortBy=bankCode&sort=ASC&page=1&perPage=20",
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				svc.On("List", constant.ValueCtxMockType(), mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*payoutManualProcessingAccountModel.PayoutManualProcessingAccount{},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    20,
							TotalItems: 0,
							TotalPages: 0,
						},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "message":"Success", "data":{"data":[], "meta":{"page":1,"perPage":20,"totalItems":0,"totalPages":0}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := serviceMocks.NewIPayoutManualProcessingAccountService(t)
			test.modifierMock(svc)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/crm/v1/payout-manual-processing-accounts?"+test.query, nil)

			router := chi.NewRouter()
			router.Get("/crm/v1/payout-manual-processing-accounts", payoutManualProcessingAccountController.New(svc, nil).List)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
