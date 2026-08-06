package internalXbController

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func TestSubmitRfiDetails(t *testing.T) {
	validPayoutID := "cb7cde08-afc0-49fc-bda1-b467d244c776"
	validMerchantID := "c95088e9-8fa8-4a94-9dc2-db87e2d84bc1"

	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
	logger := logger.NewSlogger(logger.Config{})

	controller := New(cfg, WithXbPayoutService(xbPayoutSvc), WithLogger(logger))

	validResponse := &xbModel.SubmitRfiDetailsResponse{
		Uuid:       validPayoutID,
		MerchantId: validMerchantID,
	}

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	tests := []struct {
		name           string
		payoutID       string
		modifierMock   func()
		reqSetting     func(r *http.Request)
		fileName       string
		fileContent    []byte
		formData       map[string]string
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:     "ERROR: Invalid PayoutID format",
			payoutID: "invalid",
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "documentId",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			reqSetting: func(r *http.Request) {
				validRequestID(r)
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_required", "error":{"details":[{"field":"id", "message":"Make sure id value is fulfilled"}], "traceId":"", "type":"API_ERROR"}, "message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: Missing required params",
			formData: map[string]string{
				"documentId":       "ab9fda37-6a6b-4f5e-9621-34050abb3ea4",
				"documentEntityId": "documentEntityId",
				"value":            "value",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				// empty modifier
			},
			reqSetting: func(r *http.Request) {
				validRequestID(r)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "ERROR: value and document cannot be exist at the same time",
			fileName:    "document",
			fileContent: []byte("document"),
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "ab9fda37-6a6b-4f5e-9621-34050abb3ea4",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				// empty modifier
			},
			reqSetting: func(r *http.Request) {
				validRequestID(r)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "ERROR: error when extracting file from form data",
			fileName:    "document",
			fileContent: nil,
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "ab9fda37-6a6b-4f5e-9621-34050abb3ea4",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
			},
			payoutID: validPayoutID,
			reqSetting: func(r *http.Request) {
				validRequestID(r)
				// Initialize a new multipart form to simulate the error case
				r.Form = map[string][]string{
					"merchantId":       {validMerchantID},
					"documentId":       {"ab9fda37-6a6b-4f5e-9621-34050abb3ea4"},
					"documentEntityId": {"documentEntityId"},
					"comment":          {"comment"},
				}
				r.MultipartForm = &multipart.Form{
					Value: map[string][]string{
						"merchantId":       {validMerchantID},
						"documentId":       {"ab9fda37-6a6b-4f5e-9621-34050abb3ea4"},
						"documentEntityId": {"documentEntityId"},
						"comment":          {"comment"},
					},
					File: map[string][]*multipart.FileHeader{
						"document": {
							{
								Filename: "document.pdf",
								Header:   map[string][]string{"Content-Disposition": {"form-data; name=\"document\"; filename=\"document.pdf\""}},
								Size:     123, // Size must be set, or you can adjust it accordingly
								// Set an empty io.Reader to simulate an error during file reading,
							},
						},
					},
				}
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "ERROR: either value or document must be exist",
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "ab9fda37-6a6b-4f5e-9621-34050abb3ea4",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				// empty modifier
			},
			reqSetting: func(r *http.Request) {
				validRequestID(r)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "ERROR: SubmitRfiDetails service error",
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "ab9fda37-6a6b-4f5e-9621-34050abb3ea4",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				xbPayoutSvc.On(
					"SubmitRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.SubmitRfiDetailsRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			reqSetting: func(r *http.Request) {
				validRequestID(r)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:     "SUCCESS",
			payoutID: validPayoutID,
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "ab9fda37-6a6b-4f5e-9621-34050abb3ea4",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			modifierMock: func() {
				xbPayoutSvc.On(
					"SubmitRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.SubmitRfiDetailsRequest"),
				).Return(validResponse, nil)
			},
			reqSetting: func(r *http.Request) {
				validRequestID(r)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"merchantId":"c95088e9-8fa8-4a94-9dc2-db87e2d84bc1", "referenceId":"", "uuid":"cb7cde08-afc0-49fc-bda1-b467d244c776"}, "message":"Success"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			// create multipart form
			body, contentType, _ := createMultipartForm(test.formData, test.fileName, test.fileContent)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/xb/payout/%s/submit-rfi", test.payoutID), body)
			if test.reqSetting != nil {
				test.reqSetting(req)
			}
			req.Header.Set("Content-Type", contentType)

			router := chi.NewRouter()
			router.Post("/xb/payout/{id}/submit-rfi", controller.SubmitRfiDetails)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func createMultipartForm(formData map[string]string, fileFieldName string, fileContent []byte) (*bytes.Buffer, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range formData {
		err := writer.WriteField(key, val)
		if err != nil {
			return nil, "", err
		}
	}

	if fileFieldName != "" && fileContent != nil {
		// Introduce an error during file creation to trigger the error handling in the controller
		part, err := writer.CreateFormFile(fileFieldName, "dummy.pdf")
		if err != nil {
			return nil, "", err
		}

		_, err = io.Copy(part, bytes.NewReader(fileContent))
		if err != nil {
			return nil, "", err
		}
	} else if fileFieldName != "" && fileContent == nil {
		// Simulate an error by skipping file content
		return body, "", fmt.Errorf("failed to extract file from form data")
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
