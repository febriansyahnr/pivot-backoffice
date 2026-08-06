package withdrawalService

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
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

func (s *withdrawalService) Export(ctx context.Context, request *withdrawal.WithdrawalListRequest) (*withdrawal.WithdrawalDownloadResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/Export")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	internalErr := pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

	endDateUsed := request.StartDate
	if request.EndDate.Before(time.Now().UTC()) {

		endDateUsed = request.EndDate
		if url, err := s.internal.GetCacheDownloadHistory(ctx, request.HashFilterKey(endDateUsed)); err != nil {
			s.logger.Error(ctx, "Failed when get cache download history previous day", logger.Error(err))
			return nil, internalErr

		} else if url != "" {
			return &withdrawal.WithdrawalDownloadResponse{URL: url}, nil
		}
	}

	listRequest := &withdrawal.WithdrawalHistoryRequest{
		WithdrawalListRequest: request, Page: 1, PerPage: 1_048_576,
	}
	result, err := s.internal.GetList(ctx, listRequest)
	if err != nil {
		s.logger.Error(ctx, "Failed when get list of withdrawal histories", logger.Error(err))
		return nil, internalErr
	}
	transactions, _ := result.Data.([]withdrawal.WithdrawalHistoryResponse)

	if endDateUsed.Equal(request.StartDate) {
		for _, trx := range transactions {
			if trx.Date.After(endDateUsed) {
				endDateUsed = trx.Date
			}
		}
		if url, err := s.internal.GetCacheDownloadHistory(ctx, request.HashFilterKey(endDateUsed)); err != nil {
			s.logger.Error(ctx, "Failed when get cache download history", logger.Error(err))
			return nil, internalErr

		} else if url != "" {
			return &withdrawal.WithdrawalDownloadResponse{URL: url}, nil
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
		objectName       = "downloads/withdrawal-history/" + hashFilterKey + ".xlsx"
		downloadFilename = fmt.Sprintf("attachment; filename=withdrawal_histories_%d.xlsx", time.Now().In(loc).Unix())
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

	key := fmt.Sprintf(constant.RedisKeyDownloadWithdrawalHistoryFmt, hashFilterKey)

	_ = s.redis.Set(ctx, key, signedURL, expires).Err()
	return &withdrawal.WithdrawalDownloadResponse{URL: signedURL}, nil
}

func (s *withdrawalService) GetCacheDownloadHistory(ctx context.Context, hashFilterKey string) (url string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/GetCacheDownloadHistory")
	defer segment.End()

	key := fmt.Sprintf(constant.RedisKeyDownloadWithdrawalHistoryFmt, hashFilterKey)
	if err = s.redis.Get(ctx, key).Scan(&url); err != nil && !errors.Is(err, redisExt.ErrNil) {
		return "", err
	}
	return url, nil
}
