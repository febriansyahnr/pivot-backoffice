package industry

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	mockService := serviceMocks.NewIIndustryService(t)
	validate := validatorExt.New()

	router := chi.NewRouter()
	router.Put("/industry/{id}", NewController(mockService, validate).Update)

	now := time.Now().UTC()
	updatedIndustry := &industryModel.Industry{
		UUID:           "test-uuid-123",
		ParentIndustry: "Technology",
		ChildIndustry:  "Software",
		RiskLevel:      "Medium",
		MCC:            "5734",
		CommonMCC:      "5734",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	validRequestBody := `{
		"riskLevel": "Medium"
	}`

	tests := []struct {
		name           string
		industryID     string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
		isJSON         bool
	}{
		{
			name:        "SUCCESS: Update industry risk level",
			industryID:  "test-uuid-123",
			requestBody: validRequestBody,
			setupMock: func() {
				mockService.On(
					"UpdateIndustry", mock.Anything, mock.Anything,
				).Once().Return(updatedIndustry, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":{"uuid":"test-uuid-123","parentIndustry":"Technology","childIndustry":"Software","riskLevel":"Medium","mcc":"5734","commonMcc":"5734","createdAt":"%s","updatedAt":"%s"}}`, now.Format("2006-01-02T15:04:05.999999999Z"), now.Format("2006-01-02T15:04:05.999999999Z")),
			isJSON:         true,
		},
		{
			name:           "ERROR: Missing industry ID",
			industryID:     "",
			requestBody:    validRequestBody,
			setupMock:      func() {},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   "404 page not found\n",
			isJSON:         false,
		},
		{
			name:           "ERROR: Invalid JSON body",
			industryID:     "test-uuid-123",
			requestBody:    `{invalid json}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"error when decode body payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
		{
			name:       "ERROR: Invalid risk level",
			industryID: "test-uuid-123",
			requestBody: `{
				"riskLevel": "Invalid"
			}`,
			setupMock: func() {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"RiskLevel","message":"Key: 'UpdateIndustryRequest.RiskLevel' Error:Field validation for 'RiskLevel' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
			isJSON:         true,
		},
		{
			name:        "ERROR: Industry not found",
			industryID:  "non-existent-uuid",
			requestBody: validRequestBody,
			setupMock: func() {
				mockService.On(
					"UpdateIndustry", mock.Anything, mock.Anything,
				).Once().Return(nil, pkgErrs.New(response.HttpErrRequest, c.ErrIndustryNotFound))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"industry not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
		{
			name:        "ERROR: Duplicate industry",
			industryID:  "test-uuid-123",
			requestBody: `{"parentIndustry": "Finance", "childIndustry": "Banking"}`,
			setupMock: func() {
				mockService.On(
					"UpdateIndustry", mock.Anything, mock.Anything,
				).Once().Return(nil, pkgErrs.New(response.HttpErrRequest, c.ErrDuplicateIndustry))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"industry with this parent and child already exists","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
		{
			name:        "ERROR: Service error",
			industryID:  "test-uuid-123",
			requestBody: validRequestBody,
			setupMock: func() {
				mockService.On(
					"UpdateIndustry", mock.Anything, mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			url := "/industry/" + test.industryID
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(test.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.industryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if test.isJSON {
				assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
			} else {
				assert.Equal(t, test.wantRespBody, rec.Body.String())
			}
		})
	}
}
