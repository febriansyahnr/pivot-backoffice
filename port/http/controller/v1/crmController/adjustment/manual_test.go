package adjustment_test

import (
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	funcMock "github.com/paper-indonesia/pivot-backoffice/mocks/func"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/adjustment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateManualTopup(t *testing.T) {

	adjustSvcMock := serviceMocks.NewIAdjustmentService(t)

	router := chi.NewRouter()
	router.Post("/balances/topup/manual", New(adjustSvcMock).CreateManualTopup)

	uniqueID := "66eb9f9e-2b7b-4267-9666-589e051cfa08"
	postForm := func(r *http.Request) {
		r.PostForm = url.Values{
			"merchant_id":       {uniqueID},
			"bank_reference_id": {"REF-001"},
			"bank_name":         {"BCA"},
			"bank_account":      {"123456"},
			"currency":          {"IDR"},
			"created_by":        {"JOHN WICK"},
			"notes":             {"P.R"},
			"amount":            {"10000"},
		}
	}

	tests := []struct {
		name           string
		fileName       string
		fileContent    string
		reqSetup       func(r *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:Invalid request body",
			reqSetup: func(r *http.Request) {
				postForm(r)
				r.PostForm.Del("merchant_id")
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantID":"Key: 'ManualTopupRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR:Invalid request amount",
			reqSetup: func(r *http.Request) {
				postForm(r)
				r.PostForm.Set("amount", "invalid")
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"Make sure amount request format is correct"}`,
		},
		{
			name: "ERROR:Invalid request send_callback",
			reqSetup: func(r *http.Request) {
				postForm(r)
				r.PostForm.Set("send_callback", "invalid")
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"Make sure send_callback request format is correct"}`,
		},
		{
			name: "ERROR:No file sent",
			reqSetup: func(r *http.Request) {
				postForm(r)
				r.MultipartForm = &multipart.Form{}
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"http: no such file"}`,
		},
		{
			name:     "ERROR:Transfer proof format is not supported",
			fileName: "bukti_transfer.pdf",
			reqSetup: func(r *http.Request) {
				postForm(r)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"transfer proof format is not supported"}`,
		},
		{
			name:     "ERROR:Create manual topup",
			fileName: "bukti_transfer.jpg",
			reqSetup: func(r *http.Request) {
				postForm(r)
				adjustSvcMock.On(
					"CreateManualTopup", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return("", errors.New("Invalid db session"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"Invalid db session"}`,
		},
		{
			name:     "SUCCESS: Send callback is true",
			fileName: "bukti_transfer.jpg",
			reqSetup: func(r *http.Request) {
				postForm(r)
				r.PostForm.Set("send_callback", "true")
				adjustSvcMock.On(
					"CreateManualTopup", constant.ValueCtxMockType(), mock.MatchedBy(func(req *model.ManualTopupRequest) bool {
						return req.SendCallback
					}),
				).Once().Return(uniqueID, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"` + uniqueID + `"}}`,
		},
		{
			name:     "SUCCESS: Send callback is false",
			fileName: "bukti_transfer.jpg",
			reqSetup: func(r *http.Request) {
				postForm(r)
				adjustSvcMock.On(
					"CreateManualTopup", constant.ValueCtxMockType(), mock.MatchedBy(func(req *model.ManualTopupRequest) bool {
						return !req.SendCallback
					}),
				).Once().Return(uniqueID, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"` + uniqueID + `"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.fileName == "" {
				test.fileName = "testfile.txt"
			}
			body, contentType, err := funcMock.CreateMultipartFormFile("proof_of_transfer", test.fileName, test.fileContent)
			require.NoError(t, err)
			require.NotEqual(t, "", contentType)
			defer body.Reset()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/balances/topup/manual", body)
			req.Header.Set("Content-Type", contentType)

			test.reqSetup(req)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
