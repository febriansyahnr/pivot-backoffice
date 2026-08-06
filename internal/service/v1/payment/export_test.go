package paymentService_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExport(t *testing.T) {
	gcs := gcsMock.NewGCSService(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	rdb := redisMock.NewIRedisExt(t)
	internalFunc := serviceMocks.NewIPaymentInternalDirectFunc(t)

	service := New(
		nil, logger, nil, nil, nil, nil, nil,
		WithGCSService(gcs),
		WithRedisClient(rdb),
		WithInternalDirectFunc(internalFunc),
	)

	now := time.Now().UTC()
	setupDay := now.Day() - 1
	if now.Hour() >= 17 {
		setupDay = now.Day()
	}
	earlyDays := time.Date(now.Year(), now.Month(), setupDay, 17, 0, 0, 0, time.UTC)

	request1 := &paymentModel.PaymentDownloadHistoryRequest{
		StartDate: earlyDays.Add(-24 * time.Hour).Format(time.DateOnly),
		EndDate:   earlyDays.Add(-time.Second).Format(time.DateOnly),
	}
	request2 := &paymentModel.PaymentDownloadHistoryRequest{
		StartDate: earlyDays.Format(time.DateOnly),
		EndDate:   earlyDays.Add(24 * time.Hour).Format(time.DateOnly),
	}

	traceId := "e9be40d9-0dc7-421f-af07-67dc7ab93798"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)
	internalErr := pkgErrs.New(response.HttpErrInternal, fmt.Errorf(c.InternalErrorFmt, traceId))
	paymentHistories := []paymentModel.PaymentHistoryItem{
		{
			UUID: "9bf1ca55-831a-4c95-b612-20023168cc11", CreatedAt: earlyDays.Add(6 * time.Hour),
		},
	}
	tests := []struct {
		name       string
		request    *paymentModel.PaymentDownloadHistoryRequest
		setupMock  func()
		wantErr    error
		wantResult *paymentModel.PaymentDownloadHistoryResponse
	}{
		{
			name:    "ERROR:Get download cache #01",
			request: request1,
			setupMock: func() {
				internalFunc.On(
					"GetCacheDownloadHistory", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantErr: internalErr,
		},
		{
			name:    "SUCCESS:Cache found #1",
			request: request1,
			setupMock: func() {
				internalFunc.On(
					"GetCacheDownloadHistory", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return("https://cache1", nil)
			},
			wantResult: &paymentModel.PaymentDownloadHistoryResponse{URL: "https://cache1"},
		},
		{
			name:    "ERROR:Get list of withdrawal",
			request: request2,
			setupMock: func() {
				internalFunc.On(
					"FilterPaymentHistory", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: internalErr,
		},
		{
			name:    "ERROR:Get download cache #02",
			request: request2,
			setupMock: func() {
				internalFunc.On(
					"FilterPaymentHistory", c.ValueCtxMockType(), mock.Anything,
				).Return(&commonModel.PaginationResponse{Data: paymentHistories}, nil)
				internalFunc.On(
					"GetCacheDownloadHistory", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)

			},
			wantErr: internalErr,
		},
		{
			name:    "SUCCESS:Cache found #2",
			request: request1,
			setupMock: func() {
				internalFunc.On(
					"GetCacheDownloadHistory", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return("https://cache2", nil)
			},
			wantResult: &paymentModel.PaymentDownloadHistoryResponse{URL: "https://cache2"},
		},
		{
			name:    "ERROR:Generate file excel",
			request: request1,
			setupMock: func() {
				internalFunc.On(
					"GetCacheDownloadHistory", c.ValueCtxMockType(), c.StringMockType(),
				).Return("", nil)
				internalFunc.On(
					"ExportToExcel", c.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: internalErr,
		},
		{
			name:    "ERROR:Upload file excel",
			request: request1,
			setupMock: func() {
				internalFunc.On(
					"ExportToExcel", c.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)
				gcs.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)

			},
			wantErr: internalErr,
		},
		{
			name:    "ERROR:Create signed URL",
			request: request1,
			setupMock: func() {
				gcs.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything,
				).Return(nil, nil)
				gcs.On(
					"CreateSignedURL", c.ValueCtxMockType(), c.StringMockType(), c.DurationMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantErr: internalErr,
		},
		{
			name:    "SUCCESS",
			request: request1,
			setupMock: func() {
				gcs.On(
					"CreateSignedURL", c.ValueCtxMockType(), c.StringMockType(), c.DurationMockType(),
				).Return("https://no-cache", nil)
				rdb.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.DurationMockType(),
				).Return(&redis.StatusCmd{})
			},
			wantResult: &paymentModel.PaymentDownloadHistoryResponse{URL: "https://no-cache"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.Export(ctx, test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
