package walletTransaction_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/wallet/transaction"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantTransactionHistoryList(t *testing.T) {
	log := loggerMocks.NewILogger(t)
	repo := repoMocks.NewIWalletTransactionRepository(t)

	service := New(log, repo, nil, nil)

	transactionList := []walletTransactionModel.MerchantTransactionHistoryListResp{
		{Id: "0b7033a5-de62-4172-9137-bced3b09f702"},
	}

	now := time.Now()
	req := walletTransactionModel.MerchantTransactionHistoryListReq{
		Page:      1,
		PerPage:   5,
		StartDate: now.AddDate(0, -1, 0),
		EndDate:   now,
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *commonModel.PaginationResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				repo.On(
					"GetMerchantTransactionHistoryList", mock.Anything, req,
				).Once().Return(nil, int64(0), constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while get merchant transaction history list", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"GetMerchantTransactionHistoryList", mock.Anything, req,
				).Once().Return(transactionList, int64(1), nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: transactionList,
				Meta: commonModel.Meta{
					Page: 1, PerPage: 5, TotalItems: 1, TotalPages: 1,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetMerchantTransactionHistoryList(context.Background(), req)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetMerchantTransactionDetail(t *testing.T) {
	log := loggerMocks.NewILogger(t)
	repo := repoMocks.NewIWalletTransactionRepository(t)

	service := New(log, repo, nil, nil)

	merchantId := "abcfd277-2205-4467-b039-0164c35462e0"
	transactionId := "b037e3b1-549e-419a-b541-d128582e4ddb"
	result := &walletTransactionModel.MerchantTransactionDetailResp{
		Id: "b037e3b1-549e-419a-b541-d128582e4ddb",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *walletTransactionModel.MerchantTransactionDetailResp
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				repo.On(
					"GetMerchantTransactionDetail", mock.Anything, merchantId, transactionId,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while get merchant transaction detail", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				repo.On(
					"GetMerchantTransactionDetail", mock.Anything, merchantId, transactionId,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"GetMerchantTransactionDetail", mock.Anything, merchantId, transactionId,
				).Once().Return(result, nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetMerchantTransactionDetail(context.Background(), merchantId, transactionId)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestExportMerchantTransactionHistoryList(t *testing.T) {
	rdb, mocks := redismock.NewClientMock()

	log := loggerMocks.NewILogger(t)
	repo := repoMocks.NewIWalletTransactionRepository(t)
	internalSvc := serviceMocks.NewIWalletTransactionInternalService(t)
	storage := gcsMock.NewGCSService(t)

	service := New(log, repo, redisExt.WrapRedisClient(rdb, nil), storage, WithTestInternalService(internalSvc))

	request := walletTransactionModel.MerchantTransactionHistoryListReq{
		StartDate: time.Date(2025, 3, 1, 16, 59, 59, 0, time.UTC),
		EndDate:   time.Date(2025, 3, 10, 17, 00, 00, 0, time.UTC),
		Sort:      "date",
	}
	hashFilter := request.HashFilter(constant.TimeLoc)

	downloadCache := commonModel.ExportResponse{
		DownloadURL: "https://",
	}
	rawDownloadCache, _ := json.Marshal(downloadCache)

	cacheKey := fmt.Sprintf(constant.RedisKeyDownloadWalletMerchantTransactionHistoryFmt, hashFilter)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *commonModel.ExportResponse
	}{
		{
			name: "SUCCESS:Cache found",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).SetVal(string(rawDownloadCache))
			},
			wantResult: &downloadCache,
		},
		{
			name: "ERROR:Get download cache",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).SetErr(constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while retrieving download cache", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Get merchant transaction history",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).RedisNil()
				repo.On("GetMerchantTransactionHistoryListForExport", mock.Anything, request).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while get merchant transaction history list", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Export excel merchant transaction history",
			setupMock: func() {
				repo.On(
					"GetMerchantTransactionHistoryListForExport", mock.Anything, request,
				).Return([]walletTransactionModel.MerchantTransactionHistoryListResp{}, nil)

				mocks.ExpectGet(cacheKey).RedisNil()
				internalSvc.On(
					"ExportExcelMerchantTransactionHistoryList", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while export merchant transaction history list", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Upload file to gcs",
			setupMock: func() {
				internalSvc.On(
					"ExportExcelMerchantTransactionHistoryList", mock.Anything, mock.Anything, mock.Anything,
				).Return(new(bytes.Buffer), nil)

				mocks.ExpectGet(cacheKey).RedisNil()
				storage.On(
					"UploadFile", mock.Anything, mock.Anything, mock.Anything, true, mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while upload file to gcs", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Create signed URL",
			setupMock: func() {
				storage.On(
					"UploadFile", mock.Anything, mock.Anything, mock.Anything, true, mock.Anything,
				).Return(nil, nil)

				mocks.ExpectGet(cacheKey).RedisNil()
				storage.On(
					"CreateSignedURL", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return("", constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while create signed URL", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Set signed URL to cache",
			setupMock: func() {
				storage.On("CreateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://", nil)

				mocks.ExpectGet(cacheKey).RedisNil()
				mocks.Regexp().ExpectSet(cacheKey, "https://", 15*time.Minute).SetErr(constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while set signature URL", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).RedisNil()
				mocks.Regexp().ExpectSet(cacheKey, "https://", 15*time.Minute).SetVal("")
			},
			wantResult: &downloadCache,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.ExportMerchantTransactionHistoryList(context.Background(), request)
			assert.Equal(t, test.wantErr, err)
			if test.wantErr != nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, test.wantResult.DownloadURL, result.DownloadURL)
			}
		})
	}
}
