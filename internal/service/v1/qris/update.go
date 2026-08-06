package qris

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *qrisService) ReuploadDocument(ctx context.Context, req *qris.ReuploadDocumentReq) (resp *qris.ReuploadDocumentResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/ReuploadDocument")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	registration, err := s.repository.FindRegistrationById(ctx, req.RegistrationId)
	if err != nil {
		s.logger.Error(ctx, "find registration by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("RM: "+constant.InternalErrorFmt, traceId))

	} else if registration == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("registration not found"))

	} else if registration.Status != constant.FailedReg {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("registration status must be failed to be able to reupload document"))
	}

	document, err := s.merchantRepo.FindDocumentByType(ctx, registration.MerchantId, req.DocumentType)
	if err != nil {
		s.logger.Error(ctx, "find document by type", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("FD: "+constant.InternalErrorFmt, traceId))
	}

	var qrisDocument merchant.QrisDocument

	if req.DocumentType == certificateEstablishmentType {
		mergeFiles := []pdf.MergeFile{
			{
				From:     pdf.GCSFile,
				Bucket:   document.ObjLocation.Bucket,
				Location: document.ObjLocation.Object,
			},
		}
		bods, err := s.merchantRepo.GetListMerchantBODs(ctx, registration.MerchantId)
		if err != nil {
			s.logger.Error(ctx, "get list merchant bod", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("MBOD: "+constant.InternalErrorFmt, traceId))
		}
		for _, bod := range bods {
			_ = json.Unmarshal(bod.File, &bod.ObjFile)

			mergeFiles = append(mergeFiles, pdf.MergeFile{
				From:     pdf.GCSFile,
				Bucket:   bod.ObjFile.Bucket,
				Location: bod.ObjFile.Object,
			})
		}

		mergedRaw, err := s.pdf.MergeFilesToPDF(ctx, mergeFiles)
		if err != nil {
			s.logger.Error(ctx, "merge document to single pdf file", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf("MF: "+constant.InternalErrorFmt, traceId))
		}

		objectName := fmt.Sprintf(
			"%s/%s/%s", s.config.GCSConfig.MerchantDocumentFolderName, registration.MerchantId, fmt.Sprintf("%s-And-BOD.pdf", certificateEstablishmentType),
		)

		gcsResult, err := s.gcs.UploadFile(ctx, objectName, bytes.NewReader(mergedRaw), true)
		if err != nil {
			s.logger.Error(ctx, "upload merging file to gcs", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf("UP: "+constant.InternalErrorFmt, traceId))
		}
		
		// Safety check for gcsResult
		if gcsResult == nil {
			s.logger.Error(ctx, "gcs upload returned nil result")
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf("UP: "+constant.InternalErrorFmt, traceId))
		}

		qrisDocument = merchant.QrisDocument{
			Type:   certificateEstablishmentType,
			Number: "-",
			Location: merchant.DocLocation{
				Bucket: gcsResult.Bucket,
				Object: gcsResult.ObjectName,
			},
		}

	} else {
		qrisDocument = merchant.QrisDocument{
			Type:        req.DocumentType,
			Number:      document.Identifier,
			LocationRaw: document.Location,
			Location:    document.ObjLocation,
		}
	}

	if err = s.uploadDocument(ctx, registration.Id, registration.Acquirer, qrisDocument); err != nil {
		return nil, err
	}
	return &qris.ReuploadDocumentResp{Uploaded: true}, nil
}
