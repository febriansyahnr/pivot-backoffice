package paymentService

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

var bufPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

const expires = 24 * time.Hour

func (s *PaymentService) Export(ctx context.Context, request *paymentModel.PaymentDownloadHistoryRequest) (*paymentModel.PaymentDownloadHistoryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/Export")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	internalErr := pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

	paymentReq := request.ToFilterPaymentHistoryOption()

	startDate := paymentReq.StartDate.In(time.UTC)
	endDateUsed := paymentReq.StartDate.In(time.UTC)
	if paymentReq.EndDate.In(time.UTC).Before(time.Now().UTC()) {

		endDateUsed = paymentReq.EndDate.In(time.UTC)
		if url, err := s.internal.GetCacheDownloadHistory(ctx, request.HashFilterKey(endDateUsed)); err != nil {
			s.logger.Error(ctx, "Failed when get cache download history previous day", logger.Error(err))
			return nil, internalErr

		} else if url != "" {
			return &paymentModel.PaymentDownloadHistoryResponse{URL: url}, nil
		}
	}

	// convert from +7 to UTC
	paymentReq.PaymentStartDate = paymentReq.PaymentStartDate.In(time.UTC)
	paymentReq.PaymentEndDate = paymentReq.PaymentEndDate.In(time.UTC)

	result, err := s.internal.FilterPaymentHistory(ctx, paymentReq)
	if err != nil {
		s.logger.Error(ctx, "Failed when getting payment history data", logger.Error(err))
		return nil, internalErr
	}
	transactions, _ := result.Data.([]paymentModel.PaymentHistoryItem)

	if endDateUsed.Equal(startDate) {

		for _, trx := range transactions {
			if trx.CreatedAt.After(endDateUsed) {
				endDateUsed = trx.CreatedAt
			}
		}
		if url, err := s.internal.GetCacheDownloadHistory(ctx, request.HashFilterKey(endDateUsed)); err != nil {
			s.logger.Error(ctx, "Failed when get cache download history", logger.Error(err))
			return nil, internalErr

		} else if url != "" {
			return &paymentModel.PaymentDownloadHistoryResponse{URL: url}, nil
		}
	}

	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	if err := s.internal.ExportToExcel(ctx, request, transactions, buf); err != nil {
		s.logger.Error(ctx, "Failed when generate file excel", logger.Error(err))
		return nil, internalErr
	}

	var (
		hashFilterKey    = request.HashFilterKey(endDateUsed)
		objectName       = "downloads/payment-histories/" + hashFilterKey + ".xlsx"
		downloadFilename = fmt.Sprintf("attachment; filename=payment_histories_%d.xlsx", time.Now().In(loc).Unix())
	)

	if _, err := s.gcs.UploadFile(ctx, objectName, buf, true, gcs.WriteContentDisposition(downloadFilename)); err != nil {
		s.logger.Error(ctx, "Failed when upload file to gcs", logger.Error(err))
		return nil, internalErr
	}

	signedURL, err := s.gcs.CreateSignedURL(ctx, objectName, expires)
	if err != nil {
		s.logger.Error(ctx, "Failed when create signed URL", logger.Error(err))
		return nil, internalErr
	}

	key := fmt.Sprintf(constant.RedisKeyDownloadPaymentHistoryFmt, hashFilterKey)

	_ = s.redis.Set(ctx, key, signedURL, expires).Err()
	return &paymentModel.PaymentDownloadHistoryResponse{URL: signedURL}, nil
}

func (s *PaymentService) GetCacheDownloadHistory(ctx context.Context, hashFilterKey string) (url string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetCacheDownloadHistory")
	defer segment.End()

	key := fmt.Sprintf(constant.RedisKeyDownloadPaymentHistoryFmt, hashFilterKey)
	if err = s.redis.Get(ctx, key).Scan(&url); err != nil && !errors.Is(err, redisExt.ErrNil) {
		return "", err
	}
	return url, nil
}
