package withdrawalCrmController_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInquiryTransaction(t *testing.T) {
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), service)

	router := chi.NewRouter()
	router.Post("/withdrawals/{id}/inquiry-transaction", handler.InquiryTransaction)

	merchantId := uuid.NewString()
	withdrawalId := uuid.NewString()
	requestBody := fmt.Sprintf(`{"merchantId": "%s"}`, merchantId)

	tests := []struct {
		name           string
		id             string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			id:             "A",
			requestBody:    "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:           "ERROR:Invalid data format",
			id:             "123",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Id":"Key: 'InquiryTransactionRequest.Id' Error:Field validation for 'Id' failed on the 'uuid' tag"}}`,
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				service.On(
					"InquiryTransaction", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.InquiryTransactionRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				service.On(
					"InquiryTransaction", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.InquiryTransactionRequest"),
				).Return(&withdrawal.InquiryTransactionResponse{
					WithdrawalDetailResponse: &withdrawal.WithdrawalDetailResponse{
						Id:                     "01920f9e-cbfc-71c8-a8c9-28b08549142c",
						CreatedAt:              time.Date(2024, 9, 25, 8, 30, 0, 0, time.UTC),
						UpdatedAt:              time.Date(2024, 9, 25, 8, 30, 1, 0, time.UTC),
						CreatedBy:              "John Wick",
						Type:                   "MANUAL",
						Amount:                 20_000,
						BankReferenceNo:        "1234",
						BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
						BeneficiaryAccountNo:   "0000001",
						BeneficiaryAccountName: "HENDRU",
					},
					UpdatedAt: time.Date(2024, 9, 25, 8, 30, 2, 0, time.UTC),
					Status:    "SUCCESS",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"01920f9e-cbfc-71c8-a8c9-28b08549142c","createdAt":"2024-09-25T08:30:00Z","createdBy":"John Wick","type":"MANUAL","amount":20000,"bankReferenceNo":"1234","beneficiaryBankName":"BANK RAKYAT INDONESIA","beneficiaryAccountNo":"0000001","beneficiaryAccountName":"HENDRU","updatedAt":"2024-09-25T08:30:02Z","status":"SUCCESS","reasonType":"","reasonDescription":""}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.id == "" {
				test.id = withdrawalId
			}
			if test.requestBody == "" {
				test.requestBody = requestBody
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPost, "/withdrawals/"+test.id+"/inquiry-transaction", strings.NewReader(test.requestBody),
			)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
