package disbursementService

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func TestGetReceiptByID(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}
	disbursementID := uuid.NewString()
	merchantID := uuid.NewString()
	merchant := &merchantModel.Merchant{UUID: merchantID, Name: "Merchant Name"}
	transactionStatusSuccess := constant.StatusSuccess
	defLockKey := fmt.Sprintf("backend-portal:disbursements:%s:receipt:lock", disbursementID)
	receiptKey := fmt.Sprintf(RedisTemplateDisbursementReceiptKey, disbursementID)

	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)

	gcsSvc := gcsMocks.NewGCSService(t)
	db, r := redismock.NewClientMock()

	tests := []struct {
		name         string
		request      *disbursementModel.GetDisbursementReceiptRequest
		modifierMock func()
		wantErr      string
	}{
		{
			name: "ERROR: Redis SetNX error",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetErr(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Got exclusive lock",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(false)
			},
			wantErr: "the same request is in progress",
		},
		{
			name: "SUCCESS: Got receipt from redis",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetVal("url-path")
			},
		},
		{
			name: "ERROR: FindByID service",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: FindByReferenceID",
			request: &disbursementModel.GetDisbursementReceiptRequest{
				ReferenceID: "reference-id",
				MerchantID:  merchantID,
			},
			modifierMock: func() {
				defLockKey := fmt.Sprintf("backend-portal:disbursements:%s:receipt:lock", "reference-id")
				receiptKey := fmt.Sprintf(RedisTemplateDisbursementReceiptKey, "reference-id")
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On(
					"FindByMerchantAndReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Disbursement not found",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
			wantErr: constant.ErrDisbursementNotFound.Error(),
		},
		{
			name: "ERROR: Transaction status not in success status",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{TransactionStatus: nil}, nil)
			},
			wantErr: constant.ErrDisbursementNotSuccessYet.Error(),
		},
		{
			name: "ERROR: Get Merchant ID",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					TransactionStatus: &transactionStatusSuccess,
					Disbursement: disbursementModel.Disbursement{
						MerchantID: uuid.NewString(),
					},
				}, nil)

				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(nil, errors.New(response.HttpErrInternal, constant.ErrFindMerchant)).Once()
			},
			wantErr: errors.New(response.HttpErrInternal, constant.ErrFindMerchant).Error(),
		},
		{
			name: "ERROR: Merchant ID is not valid",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					TransactionStatus: &transactionStatusSuccess,
					Disbursement: disbursementModel.Disbursement{
						MerchantID: uuid.NewString(),
					},
				}, nil)

				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{
					UUID: uuid.NewString(), // other merchant
				}, nil).Once()
			},
			wantErr: constant.ErrDisbursementNotFound.Error(),
		},
		{
			name: "ERROR: SetBucketWriter",
			modifierMock: func() {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				bulkID := uuid.NewString()
				beneficiaryBankName := "BRI"
				amount := decimal.NewFromInt(1000000)
				fee := decimal.NewFromFloat(constant.DefaultMerchantFee)
				totalAmount := amount.Add(fee)
				remark := "this is remark"
				bankReference := "bank-ref-01"
				createdBy := "John"
				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{
					TransactionStatus: &transactionStatusSuccess,
					Disbursement: disbursementModel.Disbursement{
						UUID:                   uuid.NewString(),
						MerchantID:             merchantID,
						ApprovedAt:             &util.TimeNow,
						BulkID:                 &bulkID,
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
				}, nil)

				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(merchant, nil).Once()

				gcsSvc.On(
					"SetBucketWriter",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrFailedToGenerateReceipt.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			tc.modifierMock()

			svc := New(
				&conf, pdkLoggerMock, nil, disbursementRepo, nil, nil,
				WithGoogleCloudStorage(gcsSvc), WithRedisClient(redisExt.WrapRedisClient(db, nil)),
				WithMerchantService(merchantSvc),
			)

			req := &disbursementModel.GetDisbursementReceiptRequest{
				DisbursementID: disbursementID,
				MerchantID:     merchantID,
			}
			if tc.request != nil {
				req = tc.request
			}
			_, err := svc.GetReceiptByID(ctx, req)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
