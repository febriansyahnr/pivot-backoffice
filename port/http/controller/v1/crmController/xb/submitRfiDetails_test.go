package crmXbController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestSubmitRfiDetails(t *testing.T) {
	svc := serviceMocks.NewIXbPayoutService(t)
	validPayoutID := uuid.NewString()
	validMerchantID := uuid.NewString()

	validResponse := &xbModel.SubmitRfiDetailsResponse{
		Uuid:       validPayoutID,
		MerchantId: validMerchantID,
	}
	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
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
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:        "ERROR: value and document cannot be exist at the same time",
			fileName:    "document",
			fileContent: []byte("document"),
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "documentId",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"value and document cannot be exist at the same time"}`,
		},
		{
			name:        "ERROR: error when extracting file from form data",
			fileName:    "document",
			fileContent: nil,
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "documentId",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
			},
			payoutID: validPayoutID,
			reqSetting: func(r *http.Request) {
				// Initialize a new multipart form to simulate the error case
				r.Form = map[string][]string{
					"merchantId":       {validMerchantID},
					"documentId":       {"documentId"},
					"documentEntityId": {"documentEntityId"},
					"comment":          {"comment"},
				}
				r.MultipartForm = &multipart.Form{
					Value: map[string][]string{
						"merchantId":       {validMerchantID},
						"documentId":       {"documentId"},
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
			wantRespBody:   `{"code":"99", "errors":"open : no such file or directory"}`,
		},
		{
			name: "ERROR: either value or document must be exist",
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "documentId",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"either value or document must be exist"}`,
		},
		{
			name: "ERROR: SubmitRfiDetails service error",
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "documentId",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			payoutID: validPayoutID,
			modifierMock: func() {
				svc.On(
					"SubmitRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.SubmitRfiDetailsRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:     "SUCCESS",
			payoutID: validPayoutID,
			formData: map[string]string{
				"merchantId":       validMerchantID,
				"documentId":       "documentId",
				"documentEntityId": "documentEntityId",
				"comment":          "comment",
				"value":            "value",
			},
			modifierMock: func() {
				svc.On(
					"SubmitRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.SubmitRfiDetailsRequest"),
				).Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
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
			router.Post("/xb/payout/{id}/submit-rfi", New(svc).SubmitRfiDetails)

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
