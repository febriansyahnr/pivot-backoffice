package merchant

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	validMerchantId := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Second)

	testCases := []struct {
		name           string
		merchantId     string
		setupMock      func(svc *mockMerchant.IMerchantService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:       "SUCCESS",
			merchantId: validMerchantId,
			setupMock: func(svc *mockMerchant.IMerchantService) {
				svc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), validMerchantId).
					Return(&merchant.Merchant{
						UUID:              validMerchantId,
						Name:              "Test Merchant",
						Description:       "Test Description",
						Address:           "Test Address",
						DistrictId:        123,
						PostCode:          "12345",
						Logo:              "test-logo.png",
						MerchantEmail:     "test@merchant.com",
						MerchantPhone:     "1234567890",
						PICEmail:          "pic@merchant.com",
						PICPhone:          "0987654321",
						BusinessCountry:   sql.NullString{String: "IDN", Valid: true},
						BusinessStructure: sql.NullString{String: "PT", Valid: true},
						BusinessType:      sql.NullString{String: "type1", Valid: true},
						PICName:           sql.NullString{String: "Test PIC", Valid: true},
						PICJobTitle:       sql.NullString{String: "Test Job Title", Valid: true},
						KYCStatus:         sql.NullString{String: constant.KYCStatusApproved, Valid: true},
						Status:            constant.MerchantStatusActive,
						CreatedAt:         now,
						UpdatedAt:         now,
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uuid":"` + validMerchantId + `","externalId":"", "kycStatus":"APPROVED","name":"Test Merchant","shortName":"","description":"Test Description","website":"","address":"Test Address","districtId":123,"postcode":"12345","logo":"test-logo.png","merchantEmail":"test@merchant.com","merchantPhone":"1234567890","picEmail":"pic@merchant.com","picPhone":"0987654321","picName":"Test PIC","picJobTitle":"Test Job Title","mid":"","businessType":"type1","businessStructure":"PT","businessCountry":"IDN","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","status":"ACTIVE","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"` + now.Format(time.RFC3339) + `","updatedAt":"` + now.Format(time.RFC3339) + `"}}`,
		},
		{
			name:       "ERROR invalid merchant ID",
			merchantId: "invalid-uuid",
			setupMock: func(svc *mockMerchant.IMerchantService) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"merchant id is not valid"}`,
		},
		{
			name:       "ERROR merchant not found",
			merchantId: validMerchantId,
			setupMock: func(svc *mockMerchant.IMerchantService) {
				svc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), validMerchantId).
					Return(nil, nil)
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   `{"code":"44","errors":"merchant not found"}`,
		},
		{
			name:       "ERROR internal server error",
			merchantId: validMerchantId,
			setupMock: func(svc *mockMerchant.IMerchantService) {
				svc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), validMerchantId).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()

			tc.setupMock(mockMerchantSvc)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/merchants/"+tc.merchantId, nil)

			router := chi.NewRouter()
			router.Get("/merchants/{merchantId}", New(mockMerchantSvc, mockUserSvc, mockValidator, nil).Get)

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())
		})
	}
}
