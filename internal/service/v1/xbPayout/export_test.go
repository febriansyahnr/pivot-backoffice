package xbPayoutService

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExportToExcel(t *testing.T) {
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	gcsSvc := gcsMocks.NewGCSService(t)

	startAt := time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC)
	endAt := time.Date(2025, 10, 9, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name         string
		request      *xbModel.ExportXbPayoutRequest
		modifierMock func()
		wantErr      string
	}{
		{
			name: "ERROR: GetList service",
			request: &xbModel.ExportXbPayoutRequest{
				MerchantID: "test-merchant",
				StartAt:    &startAt,
				EndAt:      &endAt,
				Status:     "SUCCESS",
			},
			modifierMock: func() {
				disbursementRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: UploadFile service",
			request: &xbModel.ExportXbPayoutRequest{
				MerchantID: "test-merchant",
				StartAt:    &startAt,
				EndAt:      &endAt,
				Status:     "SUCCESS",
			},
			modifierMock: func() {
				disbursementRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []*disbursementModel.DisbursementWithTransactionResponse{
						{
							DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
								Disbursement: disbursementModel.Disbursement{
									UUID:        "test-uuid",
									ReferenceID: "REF123",
									Currency:    "SAR",
									Amount:      decimal.NewFromInt(20),
									Fee:         func() *decimal.Decimal { d := decimal.NewFromInt(0); return &d }(),
									CreatedAt:   time.Now(),
									UpdatedAt:   time.Now(),
									ReasonType:  func() *string { s := constant.XbDisbursementReasonTypeSuccess; return &s }(),
									Remark:      func() *string { s := "Test remark"; return &s }(),
									MetadataObj: disbursementModel.Metadata{
										XbDetail: &xbModel.XbPayoutMetadata{
											SourceCurrency:    "IDR",
											SourceAmount:      decimal.NewFromInt(89122),
											TotalAmount:       decimal.NewFromInt(89122),
											FxRate:            decimal.NewFromFloat(0.00022441),
											DestinationFxRate: decimal.NewFromFloat(4456.1294),
											PurposeCode:       "IR020",
											SenderData: xbModel.SenderDataResponse{
												Name:                 "Test Sender",
												CountryCode:          "ID",
												CountryName:          "Indonesia",
												State:                "DKI Jakarta",
												City:                 "Jakarta",
												Address:              "Test Address",
												Postcode:             "12345",
												AccountType:          "Company",
												IdentificationType:   "Business Registration",
												IdentificationNumber: "123456",
												BankAccountNumber:    "9876543210",
												Dob:                  "1990-01-01",
												ContactCountryCode:   "+62",
												ContactNumber:        "812345678",
												SourceOfIncome:       "Salary",
											},
											BeneficiaryData: xbModel.BeneficiaryDataResponse{
												Name:               "Test Beneficiary",
												CountryCode:        "SA",
												CountryName:        "Saudi Arabia",
												State:              "Makkah",
												City:               "Husnah",
												Address:            "Test Address",
												Postcode:           "54321",
												AccountType:        "Individual",
												AccountNumber:      "SA12345",
												BankName:           "Test Bank",
												BankCode:           "12345",
												ContactCountryCode: "+81",
												ContactNumber:      "987654321",
												Email:              "test@example.com",
											},
										},
									},
								},
							},
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    1000,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)

				gcsSvc.On(
					"UploadFile",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrFailedToGenerateExcel.Error(),
		},
		{
			name: "SUCCESS: successfully export to excel",
			request: &xbModel.ExportXbPayoutRequest{
				MerchantID: "test-merchant",
				StartAt:    &startAt,
				EndAt:      &endAt,
				Status:     "SUCCESS",
			},
			modifierMock: func() {
				disbursementRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{
					Data: []*disbursementModel.DisbursementWithTransactionResponse{
						{
							DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
								Disbursement: disbursementModel.Disbursement{
									UUID:        "test-uuid",
									ReferenceID: "REF123",
									Currency:    "SAR",
									Amount:      decimal.NewFromInt(20),
									Fee:         func() *decimal.Decimal { d := decimal.NewFromInt(0); return &d }(),
									CreatedAt:   time.Now(),
									UpdatedAt:   time.Now(),
									ReasonType:  func() *string { s := constant.XbDisbursementReasonTypeSuccess; return &s }(),
									Remark:      func() *string { s := "Test remark"; return &s }(),
									MetadataObj: disbursementModel.Metadata{
										XbDetail: &xbModel.XbPayoutMetadata{
											SourceCurrency:    "IDR",
											SourceAmount:      decimal.NewFromInt(89122),
											TotalAmount:       decimal.NewFromInt(89122),
											FxRate:            decimal.NewFromFloat(0.00022441),
											DestinationFxRate: decimal.NewFromFloat(4456.1294),
											PurposeCode:       "IR020",
											SenderData: xbModel.SenderDataResponse{
												Name:                 "Test Sender",
												CountryCode:          "ID",
												CountryName:          "Indonesia",
												State:                "DKI Jakarta",
												City:                 "Jakarta",
												Address:              "Test Address",
												Postcode:             "12345",
												AccountType:          "Company",
												IdentificationType:   "Business Registration",
												IdentificationNumber: "123456",
												BankAccountNumber:    "9876543210",
												Dob:                  "1990-01-01",
												ContactCountryCode:   "+62",
												ContactNumber:        "812345678",
												SourceOfIncome:       "Salary",
											},
											BeneficiaryData: xbModel.BeneficiaryDataResponse{
												Name:               "Test Beneficiary",
												CountryCode:        "SA",
												CountryName:        "Saudi Arabia",
												State:              "Makkah",
												City:               "Husnah",
												Address:            "Test Address",
												Postcode:           "54321",
												AccountType:        "Individual",
												AccountNumber:      "SA12345",
												BankName:           "Test Bank",
												BankCode:           "12345",
												ContactCountryCode: "+81",
												ContactNumber:      "987654321",
												Email:              "test@example.com",
											},
										},
									},
								},
							},
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    1000,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)

				gcsSvc.On(
					"UploadFile",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&gcs.UploadMultipart{}, nil)

				gcsSvc.On(
					"CreateSignedURL",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
				).Return("https://signed-url.example.com", nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.modifierMock()

			ctx = context.WithValue(ctx, constant.CtxTimeZone, constant.TimeLoc)

			svc := New(pdkLoggerMock, disbursementRepo, nil, nil, WithGCS(gcsSvc))
			_, err := svc.ExportToExcel(ctx, tc.request)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			}

		})
	}
}
