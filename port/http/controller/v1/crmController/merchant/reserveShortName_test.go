package merchant_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/merchant"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadReservedShortName(t *testing.T) {
	merchantSvc := mocks.NewIMerchantService(t)

	router := chi.NewRouter()
	router.Post("/crm/v1/merchants/reserved-short-names/upload", New(merchantSvc, nil, nil, nil).UploadReservedShortName)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:Form parsing error",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/crm/v1/merchants/reserved-short-names/upload", nil)
				req.Header.Set("Content-Type", "multipart/form-data; boundary=invalid")
				return req
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "multipart: NextPart: EOF"),
		},
		{
			name: "SUCCESS:Missing file",
			setupRequest: func() *http.Request {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/crm/v1/merchants/reserved-short-names/upload", &buf)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			setupMock: func() {
				merchantSvc.On("UploadReservedShortName", c.ValueCtxMockType(), &merchant.ReservedMerchantShortNameRequest{
					File: (*multipart.FileHeader)(nil),
				}).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":true}`,
		},
		{
			name: "SUCCESS:With file",
			setupRequest: func() *http.Request {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)

				part, err := writer.CreateFormFile("file", "test.csv")
				if err != nil {
					t.Fatal(err)
				}
				io.WriteString(part, "short_name1\nshort_name2\n")
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/crm/v1/merchants/reserved-short-names/upload", &buf)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			setupMock: func() {
				merchantSvc.On("UploadReservedShortName", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.ReservedMerchantShortNameRequest")).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":true}`,
		},
		{
			name: "ERROR:Service error",
			setupRequest: func() *http.Request {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)

				part, err := writer.CreateFormFile("file", "test.csv")
				if err != nil {
					t.Fatal(err)
				}
				io.WriteString(part, "short_name1\nshort_name2\n")
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/crm/v1/merchants/reserved-short-names/upload", &buf)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			setupMock: func() {
				merchantSvc.On("UploadReservedShortName", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.ReservedMerchantShortNameRequest")).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := test.setupRequest()

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Expected:", test.wantRespBody)
				t.Log("Actual:", rec.Body.String())
			}
		})
	}
}
