package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const expiredForSignedURLs = 30 * time.Minute

func (s *MerchantService) UpsertMerchantBOD(ctx context.Context, request *merchantModel.UpsertBoardOfDirectorReq) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpsertMerchantBOD")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	if merchant, err := s.repo.FindMerchantByID(ctx, request.MerchantId); err != nil {
		s.logger.Error(ctx, "find merchant by id", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("FM: "+constant.InternalErrorFmt, traceId))

	} else if merchant == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIDNotValid)
	}

	merchantData, err := s.repo.ValidateMerchantBODData(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "validate merchant BOD data", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("VLD: "+constant.InternalErrorFmt, traceId))

	} else if !merchantData.Valid {
		if request.Method == constant.ActionPost {
			return "", pkgErrs.New(response.HttpErrDupCheck, constant.ErrDuplicateData)
		}
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	var docLocation *merchantModel.DocLocation

	gcsUpload := merchantData.IsCreate && request.IdentityFile != nil
	if !merchantData.IsCreate && request.IdentityFile != nil {
		docLocation = &merchantData.ObjIdentityFile
		gcsUpload = merchantData.Hash != request.Hash
	}

	if gcsUpload {
		// {Parent Folter Name}/{Merchant Id}/board-of-directors/{Name}-{Identity Number}-{Unix Time}.{ext}
		objectName := fmt.Sprintf(
			"%s/%s/board-of-directors/%s%s",
			s.config.GCSConfig.MerchantDocumentFolderName, request.MerchantId, util.GenerateULID(), filepath.Ext(request.IdentityFile.Filename),
		)
		uploaded, err := s.gcs.UploadFileFromMultipart(ctx, objectName, request.IdentityFile, true)
		if err != nil {
			s.logger.Error(ctx, "upload document to gcs", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UP: "+constant.InternalErrorFmt, traceId))
		}

		docLocation = &merchantModel.DocLocation{
			Bucket: uploaded.Bucket,
			Object: uploaded.ObjectName,
		}
	}

	data := request.ToUpsertData(docLocation)
	if err := s.repo.UpsertMerchantBOD(ctx, request.Method, data); err != nil {
		s.logger.Error(ctx, "upsert merchant board of director", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("BOD: "+constant.InternalErrorFmt, traceId))
	}
	if request.Method == constant.ActionPut {
		key := fmt.Sprintf(
			constant.MerchantBODFileSignedURLFmt, request.MerchantId, request.Id,
		)
		_ = s.redis.Del(ctx, key)
	}
	return data.Id, nil
}

func (s *MerchantService) GetListMerchantBODs(ctx context.Context, merchantId string) (resp []merchantModel.BoardOfDirectorResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetListMerchantBODs")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	data, err := s.repo.GetListMerchantBODs(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "get list merchant bods", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("GLS: "+constant.InternalErrorFmt, traceId))

	} else if data == nil {
		return data, nil
	}

	for i, bod := range data {
		if err = s.merchantBODSetSignedURL(ctx, merchantId, &bod); err != nil {
			return nil, err
		}
		data[i] = bod
	}
	return data, nil
}

func (s *MerchantService) merchantBODSetSignedURL(ctx context.Context, merchantId string, dst *merchantModel.BoardOfDirectorResp) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/merchantBODSetSignedURL")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	key := fmt.Sprintf(
		constant.MerchantBODFileSignedURLFmt, merchantId, dst.Id,
	)

	_ = s.redis.Get(ctx, key).Scan(&dst.IdentityFile)

	if dst.IdentityFile == "" {

		loc := &merchantModel.DocLocation{}
		if err = json.Unmarshal(dst.File, loc); err != nil {
			s.logger.Error(ctx, "unmarshal document location", logger.Error(err))
			return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("UNM: "+constant.InternalErrorFmt, traceId))
		}
		if loc.Object == "" {
			return
		}

		if dst.IdentityFile, err = s.gcs.CreateSignedURL(ctx, loc.Object, expiredForSignedURLs); err != nil {
			s.logger.Error(ctx, "generate signed url", logger.Error(err))
			return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("GEN: "+constant.InternalErrorFmt, traceId))
		}

		_ = s.redis.Set(ctx, key, dst.IdentityFile, expiredForSignedURLs)
	}
	return
}
