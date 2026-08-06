package industry

import (
	"bytes"
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

func TestCreate(t *testing.T) {
	mockService := serviceMocks.NewIIndustryService(t)
	validate := validatorExt.New()

	router := chi.NewRouter()
	router.Post("/industry", NewController(mockService, validate).Create)

	now := time.Now().UTC()
	validIndustry := &industryModel.Industry{
		UUID:           "test-uuid-123",
		ParentIndustry: "Technology",
		ChildIndustry:  "Software",
		RiskLevel:      "Low",
		MCC:            "5734",
		CommonMCC:      "5734",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	validRequestBody := `{
		"parentIndustry": "Technology",
		"childIndustry": "Software",
		"riskLevel": "Low",
		"mcc": "5734",
		"commonMcc": "5734"
	}`

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "SUCCESS: Create industry with valid request",
			requestBody: validRequestBody,
			setupMock: func() {
				mockService.On(
					"CreateIndustry", mock.Anything, mock.Anything,
				).Once().Return(validIndustry, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":{"uuid":"test-uuid-123","parentIndustry":"Technology","childIndustry":"Software","riskLevel":"Low","mcc":"5734","commonMcc":"5734","createdAt":"%s","updatedAt":"%s"}}`, now.Format("2006-01-02T15:04:05.999999999Z"), now.Format("2006-01-02T15:04:05.999999999Z")),
		},
		{
			name:           "ERROR: Invalid JSON body",
			requestBody:    `{invalid json}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"error when decode body payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Missing parent industry",
			requestBody: `{
				"childIndustry": "Software",
				"riskLevel": "Low",
				"mcc": "5734",
				"commonMcc": "5734"
			}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"ParentIndustry","message":"Key: 'CreateIndustryRequest.ParentIndustry' Error:Field validation for 'ParentIndustry' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Missing child industry",
			requestBody: `{
				"parentIndustry": "Technology",
				"riskLevel": "Low",
				"mcc": "5734",
				"commonMcc": "5734"
			}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"ChildIndustry","message":"Key: 'CreateIndustryRequest.ChildIndustry' Error:Field validation for 'ChildIndustry' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Missing risk level",
			requestBody: `{
				"parentIndustry": "Technology",
				"childIndustry": "Software",
				"mcc": "5734",
				"commonMcc": "5734"
			}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"RiskLevel","message":"Key: 'CreateIndustryRequest.RiskLevel' Error:Field validation for 'RiskLevel' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid risk level",
			requestBody: `{
				"parentIndustry": "Technology",
				"childIndustry": "Software",
				"riskLevel": "Invalid",
				"mcc": "5734",
				"commonMcc": "5734"
			}`,
			setupMock: func() {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"RiskLevel","message":"Key: 'CreateIndustryRequest.RiskLevel' Error:Field validation for 'RiskLevel' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Duplicate industry",
			requestBody: validRequestBody,
			setupMock: func() {
				mockService.On(
					"CreateIndustry", mock.Anything, mock.Anything,
				).Return(nil, pkgErrs.New(response.HttpErrRequest, c.ErrDuplicateIndustry)).Once()
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"industry with this parent and child already exists","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/industry", bytes.NewBufferString(test.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
		})
	}
}
