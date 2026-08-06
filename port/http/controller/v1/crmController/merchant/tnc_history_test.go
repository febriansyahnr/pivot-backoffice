//nolint:testpackage
package merchant

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantTNCHistory(t *testing.T) {
	validMerchantId := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Second)

	histories := []*tncModel.MerchantTNCSigningHistoryResponse{
		{
			ID:            uuid.NewString(),
			MerchantID:    validMerchantId,
			TNCVersionID:  uuid.NewString(),
			Version:       "1.2.0",
			SignedBy:      "user-1",
			SignedByEmail: "user@merchant.com",
			SignedAt:      now,
			DocumentURL:   "https://example.com/doc.pdf",
			CreatedAt:     now,
		},
		{
			ID:            uuid.NewString(),
			MerchantID:    validMerchantId,
			TNCVersionID:  uuid.NewString(),
			Version:       "1.1.0",
			SignedBy:      "user-2",
			SignedByEmail: "user2@merchant.com",
			SignedAt:      now.Add(-24 * time.Hour),
			DocumentURL:   "https://example.com/doc2.pdf",
			CreatedAt:     now.Add(-24 * time.Hour),
		},
	}

	testCases := []struct {
		name           string
		merchantId     string
		queryParams    url.Values
		setupMock      func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:       "SUCCESS: returns paginated TNC history",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"page":    []string{"1"},
				"perPage": []string{"10"},
			},
			setupMock: func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {
				paginatedResp := &commonModel.PaginationResponse{
					Data: histories,
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 2,
						TotalPages: 1,
					},
				}
				tncSvc.On("GetSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).
					Return(paginatedResp, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[{"uuid":"`,
		},
		{
			name:       "SUCCESS: filters by version",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"page":    []string{"1"},
				"perPage": []string{"10"},
				"version": []string{"1.2.0"},
			},
			setupMock: func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {
				paginatedResp := &commonModel.PaginationResponse{
					Data: histories[:1],
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}
				tncSvc.On("GetSigningHistory", mock.Anything, mock.MatchedBy(func(q *tncModel.SigningHistoryQuery) bool {
					return q.MerchantID == validMerchantId && q.TNCVersion == "1.2.0" && q.Page == 1 && q.PageSize == 10
				})).
					Return(paginatedResp, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[{"uuid":"`,
		},
		{
			name:       "SUCCESS: empty result",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"page":    []string{"1"},
				"perPage": []string{"10"},
			},
			setupMock: func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {
				paginatedResp := &commonModel.PaginationResponse{
					Data: []*tncModel.MerchantTNCSigningHistoryResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 0,
						TotalPages: 0,
					},
				}
				tncSvc.On("GetSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).
					Return(paginatedResp, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[],"meta":{"page":1,"perPage":10,"totalItems":0,"totalPages":0}}}`,
		},
		{
			name:        "SUCCESS: uses default pagination when no params provided",
			merchantId:  validMerchantId,
			queryParams: url.Values{},
			setupMock: func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {
				paginatedResp := &commonModel.PaginationResponse{
					Data: histories,
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 2,
						TotalPages: 1,
					},
				}
				tncSvc.On("GetSigningHistory", mock.Anything, mock.MatchedBy(func(q *tncModel.SigningHistoryQuery) bool {
					return q.MerchantID == validMerchantId && q.Page == constant.DefaultPage && q.PageSize == constant.DefaultPaginationPageSize
				})).
					Return(paginatedResp, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[{"uuid":"`,
		},
		{
			name:           "ERROR: empty merchant ID",
			merchantId:     "",
			queryParams:    url.Values{},
			setupMock:      func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid merchant id","data":null,"error":{"type":"API_ERROR","details":[],"traceId":""}}`,
		},
		{
			name:       "ERROR: invalid page number",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"page": []string{"invalid"},
			},
			setupMock:      func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid page number","data":null,"error":{"type":"API_ERROR","details":[],"traceId":""}}`,
		},
		{
			name:       "ERROR: page number less than 1",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"page": []string{"0"},
			},
			setupMock:      func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid page number","data":null,"error":{"type":"API_ERROR","details":[],"traceId":""}}`,
		},
		{
			name:       "ERROR: invalid perPage value",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"perPage": []string{"invalid"},
			},
			setupMock:      func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid per page number","data":null,"error":{"type":"API_ERROR","details":[],"traceId":""}}`,
		},
		{
			name:       "ERROR: perPage less than 1",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"perPage": []string{"0"},
			},
			setupMock:      func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid per page number","data":null,"error":{"type":"API_ERROR","details":[],"traceId":""}}`,
		},
		{
			name:       "ERROR: service returns error",
			merchantId: validMerchantId,
			queryParams: url.Values{
				"page":    []string{"1"},
				"perPage": []string{"10"},
			},
			setupMock: func(merchantSvc *mockService.IMerchantService, tncSvc *mockService.ITNCService) {
				tncSvc.On("GetSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).
					Return(nil, constant.ErrGetTNCSigningHistory)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"error when getting tnc signing history","data":null,"error":{"type":"UNKNOWN","details":[],"traceId":""}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMerchantSvc := mockService.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockTncSvc := mockService.NewITNCService(t)
			validate := validator.New()

			tc.setupMock(mockMerchantSvc, mockTncSvc)

			controller := New(mockMerchantSvc, mockUserSvc, validate, nil, WithTNCService(mockTncSvc))

			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/crm/v1/merchants/"+tc.merchantId+"/tncs?"+tc.queryParams.Encode(), nil)

			// Set up chi router
			router := chi.NewRouter()
			router.Get("/crm/v1/merchants/{id}/tncs", controller.GetMerchantTNCHistory)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request through the router
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tc.wantStatusCode, w.Code, "status code should match")

			// Assert response body if provided
			if tc.wantRespBody != "" {
				if tc.wantStatusCode == http.StatusOK {
					// For success responses, just check that the response contains expected fields
					assert.Contains(t, w.Body.String(), `"code":"00"`)
					assert.Contains(t, w.Body.String(), `"message":"OK"`)
					assert.Contains(t, w.Body.String(), `"data":{"data":[`)
				} else {
					// For error responses, use exact match
					assert.JSONEq(t, tc.wantRespBody, w.Body.String())
				}
			}

			// Verify mock expectations for service calls only
			if tc.wantStatusCode == http.StatusOK || tc.wantStatusCode == http.StatusInternalServerError {
				mockTncSvc.AssertExpectations(t)
			}
		})
	}
}
