package merchant

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateKYC(t *testing.T) {
	merchantId := uuid.NewString()
	mockMerchantSvc := mocks.NewIMerchantService(t)
	mockValidator := validator.New()
	mockRmq := mockRabbitMq.NewRabbitMQExt(t)

	handler := New(mockMerchantSvc, nil, mockValidator, mockRmq)

	router := chi.NewRouter()
	router.Patch("/merchants/{id}/kyc", handler.UpdateKYC)

	testCases := []struct {
		name           string
		merchantID     string
		body           string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "when the merchant id is invalid, then should return 400",
			body:           `{"status": "APPROVED"}`,
			merchantID:     "-",
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid merchant id"}`,
		},
		{
			name:           "when the payload is broken, then should return 400",
			body:           `{"status": "APPROVED"`,
			merchantID:     constant.EmptyUUID,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"unexpected EOF"}`,
		},
		{
			name:           "when the payload request was anonymous, then should return 400",
			body:           `{"status": "DONE"}`,
			merchantID:     constant.EmptyUUID,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"KYCStatus":"Key: 'UpdateMerchantKYCRequest.KYCStatus' Error:Field validation for 'KYCStatus' failed on the 'oneof' tag"}}`,
		},
		{
			name:       "when failed to update merchant kyc, then should return 400",
			body:       `{"status": "APPROVED"}`,
			merchantID: constant.EmptyUUID,
			setupMock: func() {
				mockMerchantSvc.On("UpdateKYC", constant.ValueCtxMockType(), merchant.UpdateMerchantKYCRequest{
					MerchantID: constant.EmptyUUID,
					KYCStatus:  constant.KYCStatusApproved,
				}).Return(nil, pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest)).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "when merchant kyc updated, then should return 200",
			body:       `{"status": "APPROVED"}`,
			merchantID: constant.EmptyUUID,
			setupMock: func() {
				mockMerchantSvc.On("UpdateKYC", constant.ValueCtxMockType(), merchant.UpdateMerchantKYCRequest{
					MerchantID: constant.EmptyUUID,
					KYCStatus:  constant.KYCStatusApproved,
				}).Return(&merchant.UpdateMerchantKYCResponse{
					UUID:   constant.EmptyUUID,
					Status: constant.KYCStatusApproved,
				}, nil).Once()
				mockRmq.On("PublishActivity", constant.ValueCtxMockType(), util.ValueToPtr(constant.EmptyUUID), util.ValueToPtr(constant.UserSystemType), constant.TagMerchant, constant.ActivityUserChangeKYCInfo, mock.Anything).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"00000000-0000-0000-0000-000000000000", "status":"APPROVED"}}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			rec := httptest.NewRecorder()

			if tc.merchantID == "" {
				tc.merchantID = merchantId
			}
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/%s/kyc", tc.merchantID), strings.NewReader(tc.body))

			router.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())

		})
	}
}
