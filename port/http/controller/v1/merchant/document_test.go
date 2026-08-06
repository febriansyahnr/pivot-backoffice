package merchant_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	funcMock "github.com/paper-indonesia/pivot-backoffice/mocks/func"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/merchant"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadDocument(t *testing.T) {

	service := serviceMocks.NewIMerchantService(t)

	handler := New(service, validator.New(), nil)

	router := chi.NewRouter()
	router.Post("/{id}/upload", handler.UploadDocument)

	requestFormData := func(r *http.Request) {
		r.PostForm = url.Values{
			"type":      {"NationalIdentityCard"},
			"number":    {"1234567"},
			"createdBy": {"John Wick"},
		}
	}
	fileName := "ktp.jpt"
	merchantId := uuid.NewString()
	documentId := uuid.NewString()

	tests := []struct {
		name           string
		merchantId     string
		fileName       string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "1",
			fileName:       fileName,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'UploadDocumentReq.MerchantId' Error:Field validation for 'MerchantId' failed on the 'uuid' tag"}}`,
		},
		{
			name:       "ERROR:Upload document",
			merchantId: merchantId,
			fileName:   fileName,
			setupMock: func() {
				service.On(
					"UploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.UploadDocumentReq"),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:       "SUCCESS",
			merchantId: merchantId,
			fileName:   fileName,
			setupMock: func() {
				service.On(
					"UploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.UploadDocumentReq"),
				).Return(documentId, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s"}}`, documentId),
		},
		{
			name:       "when no file is provided, then should not return error",
			merchantId: merchantId,
			setupMock: func() {
				service.On(
					"UploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.UploadDocumentReq"),
				).Return(documentId, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s"}}`, documentId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			body, contentType, err := funcMock.CreateMultipartFormFile("file", test.fileName, "")
			require.NoError(t, err)
			require.NotEmpty(t, contentType)
			defer body.Reset()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s/upload", test.merchantId), body)
			req.Header.Add(c.HeaderContentType, contentType)

			requestFormData(req)

			if test.setupMock != nil {
				test.setupMock()
			}
			router.ServeHTTP(rec, req)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
		})
	}
}
func TestParseMerchantDocumentFilterParam(t *testing.T) {
	handler := New(nil, validator.New(), nil)

	tests := []struct {
		name           string
		queryParams    string
		wantResult     merchant.MerchantDocumentFilterRequest
		wantErrMessage string
	}{
		{
			name:        "SUCCESS: Default values",
			queryParams: "",
			wantResult: merchant.MerchantDocumentFilterRequest{
				Page:         1,
				PerPage:      10,
				Sort:         "ASC",
				SortBy:       "createdAt",
				DocumentType: "",
				Identifier:   "",
				DocumentID:   "",
			},
		},
		{
			name:        "SUCCESS: Valid query params",
			queryParams: "?page=2&perPage=5&sort=DESC&sortBy=identifier&documentType=passport&keyword=123&documentID=abc",
			wantResult: merchant.MerchantDocumentFilterRequest{
				Page:         2,
				PerPage:      5,
				Sort:         "DESC",
				SortBy:       "identifier",
				DocumentType: "passport",
				Identifier:   "123",
				DocumentID:   "abc",
			},
		},
		{
			name:           "ERROR: Invalid page format",
			queryParams:    "?page=invalid",
			wantErrMessage: "invalid page format. Use number format instead",
		},
		{
			name:           "ERROR: Invalid perPage format",
			queryParams:    "?perPage=invalid",
			wantErrMessage: "invalid perPage format. Use number format instead",
		},
		{
			name:           "ERROR: Invalid startCreatedAt format",
			queryParams:    "?startCreatedAt=invalid",
			wantErrMessage: "invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format",
		},
		{
			name:           "ERROR: Invalid endCreatedAt format",
			queryParams:    "?endCreatedAt=invalid",
			wantErrMessage: "invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/documents"+test.queryParams, nil)

			result, err := handler.ParseMerchantDocumentFilterParam(req)

			if test.wantErrMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)
			}
		})
	}
}
func TestGetDocuments(t *testing.T) {
	service := serviceMocks.NewIMerchantService(t)
	handler := New(service, validator.New(), nil)

	router := chi.NewRouter()
	router.Get("/{id}/documents", handler.GetDocuments)

	merchantId := c.EmptyUUID

	tests := []struct {
		name           string
		merchantId     string
		queryParams    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid merchant id",
			merchantId:     "invalid-id",
			queryParams:    "",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"":"Key: '' Error:Field validation for '' failed on the 'uuid' tag"}}`,
		},
		{
			name:        "ERROR: Service error",
			merchantId:  merchantId,
			queryParams: "?page=1&perPage=10",
			setupMock: func() {
				service.On(
					"GetDocuments", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantDocumentFilterRequest"),
				).Return(nil, c.ErrSomeErrorForUnitTest).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS: Get documents",
			merchantId:  merchantId,
			queryParams: "?page=1&perPage=10",
			setupMock: func() {
				service.On(
					"GetDocuments", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantDocumentFilterRequest"),
				).Return(&commonModel.PaginationResponse{
					Data: []merchant.DocumentFilterResponse{
						{
							DocumentID: "doc1",
							MerchantId: merchantId,
							Type:       "passport",
							Identifier: "123456",
							BucketName: nil,
							URL:        nil,
							Status:     c.StatusApproved,
							CreatedBy:  "John Doe",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalPages: 1,
						TotalItems: 1,
					},
				}, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"id":"doc1","merchantID":"00000000-0000-0000-0000-000000000000","type":"passport","identifier":"123456","url":null,"status":"APPROVED","notes":"","createdBy":"John Doe","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}],"pagination":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/documents%s", test.merchantId, test.queryParams), nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)

			if test.name == "SUCCESS: Get documents" {
				// Replace <time> placeholder with actual time value in the response
				test.wantRespBody = strings.Replace(test.wantRespBody, "<time>", rec.Body.String()[strings.Index(rec.Body.String(), `"createdAt":"`)+13:strings.Index(rec.Body.String(), `"createdAt":"`)+32], 1)
			}

			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
		})
	}
}
