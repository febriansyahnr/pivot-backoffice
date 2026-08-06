package disbursementService

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExportToExcel(t *testing.T) {
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	gcsSvc := gcsMocks.NewGCSService(t)
	// log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	tests := []struct {
		name         string
		filter       *disbursementModel.GetDisbursementFilterRequest
		modifierMock func()
		wantErr      string
	}{
		{
			name:   "ERROR: GetList service",
			filter: nil,
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
			name: "ERROR: UploadFileToGCS service",
			filter: &disbursementModel.GetDisbursementFilterRequest{
				BulkID: "",
			},
			modifierMock: func() {
				disbursementRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{Data: []*disbursementModel.DisbursementWithTransactionResponse{}}, nil)

				gcsSvc.On(
					"UploadFileToGCS",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)

				// log.On("Error", mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantErr: constant.ErrFailedToGenerateExcel.Error(),
		},
		{
			name: "SUCCESS: For default template",
			filter: &disbursementModel.GetDisbursementFilterRequest{
				BulkID: "",
			},
			modifierMock: func() {
				bulkID := uuid.NewString()
				beneficiaryBankName := "BRI"
				amount := decimal.NewFromInt(1000000)
				fee := decimal.NewFromFloat(constant.DefaultMerchantFee)
				totalAmount := amount.Add(fee)
				remark := "this is remark"
				transactionStatus := constant.StatusSuccess
				bankReference := "bank-ref-01"
				createdBy := "John"
				disbursementRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{Data: []*disbursementModel.DisbursementWithTransactionResponse{
					{
						DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
							Disbursement: disbursementModel.Disbursement{
								UUID:                   uuid.NewString(),
								ApprovedAt:             &util.TimeNow,
								BulkID:                 &bulkID,
								BeneficiaryBankCode:    "002",
								BeneficiaryBankName:    &beneficiaryBankName,
								Amount:                 amount,
								Fee:                    &fee,
								TotalAmount:            totalAmount,
								Remark:                 &remark,
								BankReferenceNo:        &bankReference,
								CreatedBy:              &createdBy,
								ApprovedBy:             &createdBy,
								CreatedAt:              util.TimeNow,
								UpdatedAt:              util.TimeNow,
								BeneficiaryAccountNo:   "1234",
								BeneficiaryAccountName: "Doe",
								ReferenceID:            "ref-001",
								Status:                 constant.DisbursementStatusApproved,
							},
							TransactionStatus: &transactionStatus,
						},
					},
				}}, nil)

				gcsSvc.On(
					"UploadFileToGCS",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&gcs.Response{}, nil)
			},
		},
		{
			name: "SUCCESS: For upload template",
			filter: &disbursementModel.GetDisbursementFilterRequest{
				BulkID: uuid.NewString(),
			},
			modifierMock: func() {
				bulkID := uuid.NewString()
				beneficiaryBankName := "BRI"
				amount := decimal.NewFromInt(1000000)
				fee := decimal.NewFromFloat(constant.DefaultMerchantFee)
				totalAmount := amount.Add(fee)
				remark := "this is remark"
				transactionStatus := constant.StatusSuccess
				bankReference := "bank-ref-01"
				createdBy := "John"
				disbursementRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{Data: []*disbursementModel.DisbursementWithTransactionResponse{
					{
						DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
							Disbursement: disbursementModel.Disbursement{
								UUID:                   uuid.NewString(),
								ApprovedAt:             &util.TimeNow,
								BulkID:                 &bulkID,
								BeneficiaryBankCode:    "002",
								BeneficiaryBankName:    &beneficiaryBankName,
								Amount:                 amount,
								Fee:                    &fee,
								TotalAmount:            totalAmount,
								Remark:                 &remark,
								BankReferenceNo:        &bankReference,
								CreatedBy:              &createdBy,
								ApprovedBy:             &createdBy,
								CreatedAt:              util.TimeNow,
								UpdatedAt:              util.TimeNow,
								BeneficiaryAccountNo:   "1234",
								BeneficiaryAccountName: "Doe",
								ReferenceID:            "ref-001",
								Status:                 constant.DisbursementStatusApproved,
							},
							TransactionStatus: &transactionStatus,
						},
					},
				}}, nil)

				gcsSvc.On(
					"UploadFileToGCS",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&gcs.Response{}, nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.modifierMock()

			ctx = context.WithValue(ctx, constant.CtxTimeZone, constant.TimeLoc)

			svc := New(&conf, pdkLoggerMock, nil, disbursementRepo, nil, nil, WithGoogleCloudStorage(gcsSvc))
			_, err := svc.ExportToExcel(ctx, tc.filter)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			}

			os.RemoveAll(constant.ExportTempDir)
		})
	}
}
