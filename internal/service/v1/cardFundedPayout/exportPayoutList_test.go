package cardFundedPayoutService

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/tealeg/xlsx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportPayoutList(t *testing.T) {
	cfg := &config.Config{}
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	gcsService := gcsMock.NewGCSService(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithDisbursementRepository(disbursementRepo),
		WithGCS(gcsService),
	).(*service)

	validFilter := &cardFundedPayoutModel.FilterGetPayoutList{
		MerchantID: "merchant-123",
	}

	testCases := []struct {
		name      string
		filter    *cardFundedPayoutModel.FilterGetPayoutList
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "ERROR: Get card funded payout list failed",
			filter: validFilter,
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutList", mock.Anything, mock.Anything).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
		},
		{
			name:   "SUCCESS: Export payout list with empty data",
			filter: validFilter,
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*cardFundedPayoutModel.GetPayoutListResponse{},
						Meta: commonModel.Meta{
							TotalItems: 0,
							TotalPages: 0,
							Page:       1,
						},
					}, nil).Once()
				gcsService.On("UploadFileToGCS", mock.Anything, mock.Anything, mock.Anything, true, (*time.Duration)(nil)).
					Return(&gcs.Response{SignedUrl: "https://signed-url"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "SUCCESS: Export payout list with data",
			filter: validFilter,
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*cardFundedPayoutModel.GetPayoutListResponse{
							{
								UUID:              "payout-123",
								ReferenceID:       "ref-123",
								Amount:            "10000",
								Fee:               "1000",
								TotalAmount:       "11000",
								TransactionStatus: "SUCCESS",
								ApprovalStatus:    "APPROVED",
								BankName:          "BCA",
								AccountNumber:     "1234567890",
								AccountName:       "Test Account",
								Remarks:           "Test remarks",
								VendorID:          "vendor-123",
								VendorName:        "Test Vendor",
								Card: cardFundedPayoutModel.CardInfo{
									LastFour: "1234",
									Brand:    "VISA",
									Channel:  "CREDIT",
									Issuer:   "BCA",
									Expiry:   "12/25",
								},
							},
						},
						Meta: commonModel.Meta{
							TotalItems: 1,
							TotalPages: 1,
							Page:       1,
						},
					}, nil).Once()
				gcsService.On("UploadFileToGCS", mock.Anything, mock.Anything, mock.Anything, true, (*time.Duration)(nil)).
					Return(&gcs.Response{SignedUrl: "https://signed-url"}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			ctx := context.WithValue(context.Background(), constant.CtxTimeZone, constant.TimeLoc)
			result, err := svc.ExportPayoutList(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Url)
			}

			disbursementRepo.AssertExpectations(t)
			gcsService.AssertExpectations(t)
		})
	}
}

func TestGenerateExcelFile(t *testing.T) {
	cfg := &config.Config{}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log).(*service)

	testCases := []struct {
		name           string
		payouts        []*cardFundedPayoutModel.GetPayoutListResponse
		wantErr        bool
		validateAmount bool
		expectedAmount string
	}{
		{
			name:    "SUCCESS: Generate Excel file with empty data",
			payouts: []*cardFundedPayoutModel.GetPayoutListResponse{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Generate Excel file with data",
			payouts: []*cardFundedPayoutModel.GetPayoutListResponse{
				{
					UUID:              "payout-123",
					ReferenceID:       "ref-123",
					Amount:            "10000",
					TransactionStatus: "SUCCESS",
					ApprovalStatus:    "APPROVED",
					VendorName:        "Test Vendor",
					CreatedAt:         time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC),
					Card: cardFundedPayoutModel.CardInfo{
						LastFour: "1234",
						Brand:    "VISA",
						Channel:  "CREDIT",
					},
				},
			},
			wantErr:        false,
			validateAmount: true,
			expectedAmount: util.ConvertFloatToCurrency(10000),
		},
		{
			name: "SUCCESS: Generate Excel file with large amount",
			payouts: []*cardFundedPayoutModel.GetPayoutListResponse{
				{
					UUID:              "payout-456",
					ReferenceID:       "ref-456",
					Amount:            "2500000",
					TransactionStatus: "FAILED",
					ApprovalStatus:    "WAITING",
					VendorName:        "Test Vendor 2",
					CreatedAt:         time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC),
					Card: cardFundedPayoutModel.CardInfo{
						LastFour: "5678",
						Brand:    "MASTERCARD",
						Channel:  "DEBIT",
					},
				},
			},
			wantErr:        false,
			validateAmount: true,
			expectedAmount: util.ConvertFloatToCurrency(2500000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constant.CtxTimeZone, constant.TimeLoc)
			result, err := svc.generateExcelFile(ctx, "test-file", tc.payouts)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)

				if tc.validateAmount {
					xlFile, err := xlsx.OpenFile(result)
					assert.NoError(t, err)

					sheet := xlFile.Sheets[0]
					dataRow := sheet.Rows[1]
					assert.Equal(t, tc.expectedAmount, dataRow.Cells[3].Value, "Amount format should match ConvertFloatToCurrency")

					os.Remove(result)
				}
			}
		})
	}
}

func TestUploadExcelFileToGCS(t *testing.T) {
	cfg := &config.Config{}
	gcsService := gcsMock.NewGCSService(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithGCS(gcsService),
	).(*service)

	testCases := []struct {
		name      string
		filename  string
		srcFile   string
		setupMock func()
		wantErr   bool
	}{
		{
			name:     "ERROR: Upload to GCS failed",
			filename: "test-file",
			srcFile:  "tmp/test-file.xlsx",
			setupMock: func() {
				gcsService.On("UploadFileToGCS", mock.Anything, mock.Anything, mock.Anything, true, (*time.Duration)(nil)).
					Return(nil, errors.New("upload failed")).Once()
			},
			wantErr: true,
		},
		{
			name:     "SUCCESS: Upload to GCS success",
			filename: "test-file",
			srcFile:  "tmp/test-file.xlsx",
			setupMock: func() {
				gcsService.On("UploadFileToGCS", mock.Anything, mock.Anything, mock.Anything, true, (*time.Duration)(nil)).
					Return(&gcs.Response{SignedUrl: "https://signed-url"}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := svc.uploadExcelFileToGCS(context.Background(), tc.filename, tc.srcFile)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}

			gcsService.AssertExpectations(t)
		})
	}
}
