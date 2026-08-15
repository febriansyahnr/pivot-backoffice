package disbursementService_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	logMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkValidate(t *testing.T) {
	file := xlsxMock.NewFiler(t)
	file.On("Close").Return(nil)

	logger, _ := logMock.NewZapLogger(logMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	disbursementRepo.On(
		"CountByMerchantAndReference", mock.Anything, c.StringMockType(), c.StringMockType(),
	).Return(0, nil)
	feeSvc := serviceMocks.NewIFeeService(t)

	beneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)
	beneficiaryAccountSvc.On(
		"FindByBankCodeAndAccountNo", mock.Anything, mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
	).Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: "JOHN WICK"}, nil)

	statusHistoriesRepo := repoMocks.NewIStatusHistoriesRepository(t)
	statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

	excel := xlsxMock.NewExceler(t)
	rdb, clientMock := redismock.NewClientMock()

	service := New(
		&config.Config{}, logger, merchantRepo, disbursementRepo, nil, nil,
		WithExcelLibrary(excel),
		WithRedisClient(redisExt.WrapRedisClient(rdb, nil)),
		WithBeneficiaryAccService(beneficiaryAccountSvc), WithFeeService(feeSvc),
		WithStatusHistoriesRepository(statusHistoriesRepo),
	)

	traceId := uuid.NewString()
	merchantId := uuid.NewString()
	parentId := "afc60d8e-3cb6-444b-8594-265b44905ea0"
	trxConfigKey := fmt.Sprintf(c.DisbursementTransactionConfigFmt, parentId)
	emptyExcelData := [][]string{
		{"Reference ID", "Amount", "Channel Code", "Account Number", "Account Name", "Remarks"}, {},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []disbursementModel.BulkPreviewResponse
	}{
		{
			name: "ERROR:Open bulk disbursement file",
			setupMock: func() {
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
			name: "ERROR:Get merchant by id",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(emptyExcelData, nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(emptyExcelData, nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrRequest, c.ErrMerchantNotFound), // NOSONAR
		},
		{
			name: "ERROR:Get transaction config",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(emptyExcelData, nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID:      merchantId,
					Name:      "Test.Merchant",
					ParentID:  sql.NullString{String: parentId},
					KYCStatus: sql.NullString{String: c.KYCStatusNotRequired},
				}, nil)
				clientMock.ExpectGet(trxConfigKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)), // NOSONAR
		},
		{
			name: "ERROR:Empty data upload",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(emptyExcelData, nil)

				clientMock.ClearExpect()
				clientMock.ExpectGet(trxConfigKey).SetVal(`{"minAmount": 10000, "maxAmount": 1000000}`)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty data to upload")), // NOSONAR
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return([][]string{
					{"Reference ID", "Amount", "Channel Code", "Account Number", "Account Name", "Remarks"},
					{"REF001", "25000", "BRI", "000000002135", "JOHN WICK", "TEST"},
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID:      merchantId,
					Name:      "Test.Merchant",
					ParentID:  sql.NullString{String: parentId},
					KYCStatus: sql.NullString{String: c.KYCStatusNotRequired},
				}, nil)

				clientMock.ClearExpect()
				clientMock.ExpectGet(trxConfigKey).SetVal(`{"minAmount": 10000, "maxAmount": 1000000}`) // for single row validation
				clientMock.ExpectGet(trxConfigKey).SetVal(`{"minAmount": 10000, "maxAmount": 1000000}`)
				feeSvc.On(
					"GetPayoutTransactionFee", mock.Anything, mock.Anything,
				).Once().Return(&feeModel.FeeMetadataObject{}, nil)

			},
			wantResult: []disbursementModel.BulkPreviewResponse{
				{
					ReferenceID:            "REF001",
					ChannelCode:            "BRI",
					BeneficiaryBankCode:    "002",
					BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
					BeneficiaryAccountNo:   "000000002135",
					BeneficiaryAccountName: "JOHN WICK",
					Amount:                 "25000",
					Remark:                 "TEST",
					Result:                 "VALID",
					FeeAmount:              util.ValueToPtr(0.0),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			request := &disbursementModel.BulkPreviewRequest{
				MerchantId: merchantId, File: new(bytes.Buffer),
			}

			result, err := service.BulkValidate(
				context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId), request,
			)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
