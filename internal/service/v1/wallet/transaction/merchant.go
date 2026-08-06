package walletTransaction

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) GetMerchantTransactionHistoryList(ctx context.Context, req walletTransactionModel.MerchantTransactionHistoryListReq) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/wallet/transaction/GetMerchantTransactionHistoryList")
	defer segment.End()

	transactionList, totalRows, err := s.repo.GetMerchantTransactionHistoryList(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Failed while get merchant transaction history list", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	result := &commonModel.PaginationResponse{
		Data: transactionList,
		Meta: commonModel.Meta{
			Page:       req.Page,
			PerPage:    req.PerPage,
			TotalItems: totalRows,
			TotalPages: int64(math.Ceil(float64(totalRows) / float64(req.PerPage))),
		},
	}
	return result, nil
}

func (s *service) ExportMerchantTransactionHistoryList(ctx context.Context, request walletTransactionModel.MerchantTransactionHistoryListReq) (*commonModel.ExportResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/wallet/transaction/ExportMerchantTransactionHistoryList")
	defer segment.End()

	loc := util.GetTimeLocationFromContext(ctx)

	var (
		hashFilter = request.HashFilter(loc.String())
		result     = &commonModel.ExportResponse{}
		cacheKey   = fmt.Sprintf(constant.RedisKeyDownloadWalletMerchantTransactionHistoryFmt, hashFilter)
	)
	if err := s.cache.Get(ctx, cacheKey).Scan(result); err == nil {
		return result, nil

	} else if !errors.Is(err, redisExt.ErrNil) {
		s.log.Error(ctx, "Failed while retrieving download cache", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	transactionList, err := s.repo.GetMerchantTransactionHistoryListForExport(ctx, request)
	if err != nil {
		s.log.Error(ctx, "Failed while get merchant transaction history list", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	buf, err := s.internal.ExportExcelMerchantTransactionHistoryList(ctx, request, transactionList)
	if err != nil {
		s.log.Error(ctx, "Failed while export merchant transaction history list", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	objectName := "downloads/merchant-wallets/transaction-histories/" + hashFilter + ".xlsx"
	downloadFilename := "attachment; filename=merchant_wallet_transaction_histories.xlsx"

	if _, err := s.storage.UploadFile(ctx, objectName, buf, true, gcs.WriteContentDisposition(downloadFilename)); err != nil {
		s.log.Error(ctx, "Failed while upload file to gcs", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}
	result.ExpiresAt = time.Now().UTC().Add(expirationForTransactionHistory)
	result.DownloadURL, err = s.storage.CreateSignedURL(ctx, objectName, expirationForTransactionHistory)
	if err != nil {
		s.log.Error(ctx, "Failed while create signed URL", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	if err := s.cache.Set(ctx, cacheKey, result, expirationForTransactionHistory).Err(); err != nil {
		s.log.Error(ctx, "Failed while set signature URL", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}
	return result, nil
}

func (s *service) GetMerchantTransactionDetail(ctx context.Context, merchantId, id string) (*walletTransactionModel.MerchantTransactionDetailResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/wallet/transaction/GetMerchantTransactionDetail")
	defer segment.End()

	result, err := s.repo.GetMerchantTransactionDetail(ctx, merchantId, id)
	if err != nil {
		s.log.Error(ctx, "Failed while get merchant transaction detail", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)

	} else if result == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}
	return result, nil
}
