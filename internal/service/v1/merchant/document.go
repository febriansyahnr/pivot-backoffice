package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) UploadDocument(ctx context.Context, document *merchantModel.UploadDocumentReq) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UploadDocument")
	defer segment.End()

	var (
		err error
	)

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	merchant, err := s.repo.FindMerchantByID(ctx, document.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "find merchant by id", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("FM: "+constant.InternalErrorFmt, traceId))

	} else if merchant == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIDNotValid)
	}

	var gcsUpload bool

	currentDoc, err := s.repo.FindDocumentByType(ctx, document.MerchantId, document.Type)
	if err != nil {
		s.logger.Error(ctx, "find document by document type", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("FD: "+constant.InternalErrorFmt, traceId))
	}

	if currentDoc == nil || currentDoc.Hash != document.Hash {
		gcsUpload = true
	}

	var docLocation *merchantModel.DocLocation

	if !gcsUpload {
		docLocation = &currentDoc.ObjLocation

	} else if document.File != nil {
		// Parent Folder Name + Merchant Id + Document Type + Unix Time
		objectName := fmt.Sprintf(
			"%s/%s/%s-%d%s",
			s.config.GCSConfig.MerchantDocumentFolderName, document.MerchantId, document.Type, time.Now().UTC().Unix(), filepath.Ext(document.File.Filename),
		)

		uploaded, err := s.gcs.UploadFileFromMultipart(ctx, objectName, document.File, true)
		if err != nil {
			s.logger.Error(ctx, "upload document to gcs", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UP: "+constant.InternalErrorFmt, traceId))
		}
		docLocation = &merchantModel.DocLocation{
			Bucket: uploaded.Bucket,
			Object: uploaded.ObjectName,
		}
	}

	if currentDoc == nil {
		data := document.ToInsertData(docLocation)
		err = s.repo.CreateDocument(ctx, data)
		if err != nil {
			s.logger.Error(ctx, "upsert data document", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UPS: "+constant.InternalErrorFmt, traceId))
		}
		return data.Id, nil
	}

	if document.File != nil {
		currentDoc.Hash = document.Hash
	}

	if document.Notes != "" {
		currentDoc.Notes = document.Notes
	}
	currentDoc.Identifier = document.Identifier
	currentDoc.UpdatedAt = time.Now().UTC()
	currentDoc.Location, _ = json.Marshal(docLocation)
	err = s.repo.UpdateDocument(ctx, currentDoc)
	if err != nil {
		s.logger.Error(ctx, "upsert data document", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	return currentDoc.Id, nil
}

// GetDocuments retrieves a paginated list of merchant documents based on the provided filter criteria.
func (s *MerchantService) GetDocuments(ctx context.Context, request *merchantModel.MerchantDocumentFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetDocuments")
	defer segment.End()

	result, err := s.repo.GetDocuments(ctx, request)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	documents, _ := result.Data.([]*merchantModel.DocumentFilterResponse)
	for _, doc := range documents {
		if doc.BucketName != nil && doc.URL != nil {
			url, err := s.gcs.CreateSignedURL(ctx, *doc.URL, merchantDocumentExp)
			if err != nil {
				s.logger.Error(ctx, "get signed url", logger.Error(err))
				doc.URL = nil
				continue
			}
			doc.URL = &url
		}
	}

	return result, nil
}

func (s *MerchantService) FindDocumentByType(ctx context.Context, merchantId, docType string) (doc *merchantModel.Document, err error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/merchant/FindDocumentByType")
	defer span.End()

	return s.repo.FindDocumentByType(ctx, merchantId, docType)
}
