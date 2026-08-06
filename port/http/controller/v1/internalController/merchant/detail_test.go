package internal_merchant

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestDetail(t *testing.T) {

	testCases := []struct {
		name             string
		setup            func(merchantSvc *mockSvc.IMerchantService)
		setupParam       func(chi *chi.Context)
		expectedCode     int
		expectedResponse string
	}{
		{
			name: "SUCCESS: get detail merchant",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					&merchant.Merchant{
						UUID:          uuid.Max.String(),
						ExternalId:    "external-id",
						Name:          "merchant-name",
						ShortName:     "merchant-short-name",
						Description:   "merchant-description",
						Address:       "merchant-address",
						DistrictId:    123,
						PostCode:      "123456",
						Logo:          "merchant-logo",
						MerchantEmail: "merchant@merchant.com",
						MerchantPhone: "08123456789",
						PICEmail:      "pic@merchant.com",
						PICPhone:      "08123456789",
						PICName: sql.NullString{
							String: "pic-name",
							Valid:  true,
						},
						PICJobTitle: sql.NullString{
							String: "pic-job-title",
							Valid:  true,
						},
						MID:    sql.NullString{String: "mid", Valid: true},
						Status: constant.MerchantStatusActive,
					},
					nil,
				)
			},
			setupParam: func(chi *chi.Context) {
				chi.URLParams.Add("id", uuid.NewString())
			},
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","externalId":"external-id", "kycStatus":"","name":"merchant-name","shortName":"merchant-short-name","description":"merchant-description","website":"","address":"merchant-address","districtId":123,"postcode":"123456","logo":"merchant-logo","merchantEmail":"merchant@merchant.com","merchantPhone":"08123456789","picEmail":"pic@merchant.com","picPhone":"08123456789","picName":"pic-name","picJobTitle":"pic-job-title","mid":"mid","businessType":"","businessStructure":"","businessCountry":"","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","status":"ACTIVE","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "ERROR: Invalid merchant id",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
			},
			setupParam: func(chi *chi.Context) {
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"merchant id is not valid","error":{"type":"API_ERROR","message":"merchant id is not valid","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Get detail merchant",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					nil,
					errors.New("error"),
				)
			},
			setupParam: func(chi *chi.Context) {
				chi.URLParams.Add("id", uuid.NewString())
			},
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Merchant not found",
			setup: func(merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					nil,
					nil,
				)
			},
			setupParam: func(chi *chi.Context) {
				chi.URLParams.Add("id", uuid.NewString())
			},
			expectedCode:     http.StatusNotFound,
			expectedResponse: `{"code":"44","message":"merchant not found","error":{"type":"API_ERROR","message":"merchant not found","recommendation":""},"data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantSvc := mockSvc.NewIMerchantService(t)
			tc.setup(merchantSvc)
			ctrl := New(nil, merchantSvc, nil)

			baseUrl := "/internal/v1/merchants/"
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			if tc.setupParam != nil {
				tc.setupParam(chiRouterCtx)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Detail)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedCode, httpRecorder.Code)
			assert.JSONEqf(t, tc.expectedResponse, httpRecorder.Body.String(), "response not match. Expect %v, got %v", tc.expectedResponse, httpRecorder.Body.String())
		})
	}
}
