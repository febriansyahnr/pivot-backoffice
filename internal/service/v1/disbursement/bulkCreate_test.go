package disbursementService_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	gcsModel "github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	logMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkCreate(t *testing.T) {
	file := xlsxMock.NewFiler(t)

	logger, _ := logMock.NewZapLogger(logMock.Config{})
	gcs := gcsMock.NewGCSService(t)
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	beneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)

	file.On("Close").Return(nil)
	disbursementRepo.On(
		"CountByMerchantAndReference", mock.Anything, c.StringMockType(), c.StringMockType(),
	).Return(0, nil)
	rmq.On(
		"Publish", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
	).Return(nil)

	excel := xlsxMock.NewExceler(t)
	rdb, clientMock := redismock.NewClientMock()

	service := New(
		&config.Config{}, logger, merchantRepo, disbursementRepo, nil, nil,
		WithExcelLibrary(excel),
		WithRabbitMQClient(rmq),
		WithGoogleCloudStorage(gcs),
		WithRedisClient(redisExt.WrapRedisClient(rdb, nil)),
		WithBeneficiaryAccService(beneficiaryAccountSvc),
	)

	traceId := uuid.NewString()
	merchantId := uuid.NewString()
	createBy := uuid.NewString()
	parentId := "98d41141-2dcd-4cf8-8cb7-72be22495a52"

	trxConfigKey := fmt.Sprintf(c.DisbursementTransactionConfigFmt, parentId)
	queueKey := fmt.Sprintf(c.BulkDisbursementQueueLockFmt, merchantId, "REF001")

	merchantData := &merchant.Merchant{
		UUID:      merchantId,
		Name:      "Test.Merchant",
		ParentID:  sql.NullString{String: parentId},
		KYCStatus: sql.NullString{String: c.KYCStatusNotRequired},
	}
	headers := []string{"Reference ID", "Amount", "Channel Code", "Account Number", "Account Name", "Remarks"}

	emptyDisbursementData := [][]string{headers, {}}
	disbursementData := [][]string{headers, {"REF001", "25000", "BRI", "00012", "JOHN"}}

	rcases := []func(){}
	runRedisFinalCase := func() {
		clientMock.ClearExpect()

		for _, f := range rcases {
			f()
		}
	}
	addRedisFinalCases := func(f func()) {
		rcases = append(rcases, f)
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *disbursementModel.BulkCreateResponse
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrMerchantNotFound), // NOSONAR
		},
		{
			name: "ERROR:Open bulk disbursement file",
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(merchantData, nil)

				excel.On("OpenReader", c.PtrBytesBufferMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, c.ErrOpenFileReader), // NOSONAR
		},
		{
			name: "ERROR:Get rows and validate bulk upload",
			setupMock: func() {
				excel.On("OpenReader", c.PtrBytesBufferMockType()).Return(file, nil)
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("sheet to upload not found")), // NOSONAR
		},
		{
			name: "ERROR:Get transaction config",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(emptyDisbursementData, nil)
				clientMock.ExpectGet(trxConfigKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)), // NOSONAR
		},
		{
			name: "ERROR:Empty data upload",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(emptyDisbursementData, nil)
				addRedisFinalCases(func() { clientMock.ExpectGet(trxConfigKey).SetVal(`{"minAmount": 10000, "maxAmount": 1000000}`) })

				runRedisFinalCase()
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("there is no valid data")), // NOSONAR
		},
		{
			name: "ERROR:Invalid data and beneficiary account name",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return([][]string{
					headers, {"REF001", "9000", "BRI", "012", "KKK", "TEST"},
				}, nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.ValueCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)

				runRedisFinalCase()
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("there is no valid data")), // NOSONAR
		},
		{
			name: "ERROR:Invalid amount data type",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return([][]string{
					headers, {"REF001", "invalid", "BRI", "012", "KKK", "TEST"},
				}, nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.ValueCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: "JOHN"}, nil).Once()

				runRedisFinalCase()
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("there is no valid data")), // NOSONAR
		},
		{
			name: "ERROR:Set exclusive lock",
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.ValueCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: "JOHN"}, nil)
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Return(disbursementData, nil)

				runRedisFinalCase()
				clientMock.ExpectSetNX(queueKey, true, 0).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)), // NOSONAR
		},
		{
			name: "ERROR:Transaction is in queue",
			setupMock: func() {
				runRedisFinalCase()
				clientMock.ExpectSetNX(queueKey, true, 0).SetVal(false)
			},
			wantErr: pkgErrs.New(response.HttpErrDupCheck, c.ErrDisbursementReferenceIdAlreadyExist), // NOSONAR
		},
		{
			name: "ERROR:Upload disbursement file to GCS",
			setupMock: func() {
				addRedisFinalCases(func() { clientMock.ExpectSetNX(queueKey, true, 0).SetVal(true) })

				gcs.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)

				runRedisFinalCase()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Insert bulk disbursement",
			setupMock: func() {
				gcs.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(),
				).Return(&gcsModel.UploadMultipart{PublicURL: "publicURL"}, nil)

				disbursementRepo.On(
					"InsertBulkDisbursement", c.ValueCtxMockType(), mock.AnythingOfType("*disbursementModel.BulkDisbursement"),
				).Once().Return(c.ErrSomeErrorForUnitTest)

				runRedisFinalCase()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "SUCCCESS",
			setupMock: func() {
				disbursementRepo.On(
					"InsertBulkDisbursement", c.ValueCtxMockType(), mock.AnythingOfType("*disbursementModel.BulkDisbursement"),
				).Return(nil)

				runRedisFinalCase()
			},
			wantResult: &disbursementModel.BulkCreateResponse{
				MerchantID: merchantId,
				File:       "publicURL",
				Status:     "UPLOADING",
				CreatedBy:  &createBy,
				TotalData:  1, TotalAmount: 25_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			request := &disbursementModel.BulkCreateRequest{
				MerchantId: merchantId,
				CreatedBy:  createBy,
				File:       new(bytes.Buffer),
			}

			result, err := service.BulkCreate(
				context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId), request,
			)
			if test.wantResult != nil && result != nil {
				test.wantResult.UUID = result.UUID
				test.wantResult.CreatedAt = result.CreatedAt
				test.wantResult.UpdatedAt = result.UpdatedAt
			}
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
