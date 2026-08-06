package qris

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validation"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/panjf2000/ants/v2"
)

const uploadWorker = 32

func (s *qrisService) RegistrationCallback(ctx context.Context, request *qris.RegistrationCallback) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/RegistrationCallback")
	defer segment.End()

	registration, err := s.repository.FindQrRegistrationForValidationById(ctx, request.Id)
	if err != nil {
		s.logger.Error(ctx, "Find qris registration for validation by id", logger.Error(err))
		return err

	} else if registration == nil {
		s.logger.Warn(ctx, "Qris registration not found", logger.String("id", request.Id), logger.Any("payload", request))
		return nil

	} else if registration.Status == constant.SuccessReg {
		s.logger.Warn(ctx, "Qris registration status is SUCCESS, update ignored", logger.String("id", request.Id), logger.Any("payload", request))
		return nil
	}
	return s.repository.UpdateCallbackQrRegistration(ctx, request.Id, request)
}

func (s *qrisService) Registration(ctx context.Context, request *qris.RegistrationReq) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/Registration")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	if request.Type == "" {
		request.Type = constant.EnterpriseType
	}

	merchantInfo, err := s.merchantValidation(ctx, request)
	if err != nil {
		return "", err
	}
	
	// Additional safety check for nilaway
	if merchantInfo == nil {
		return "", fmt.Errorf(constant.InternalErrorFmt, traceId)
	}

	if merchantInfo.RegStatus == "" {
		registration := &qris.Registration{
			Id:                       fmt.Sprintf("%s%d", merchantInfo.MID, time.Now().UTC().UnixNano()),
			ExternalId:               merchantInfo.ExternalId,
			Acquirer:                 request.Acquirer,
			MerchantType:             request.MerchantType,
			AcquirerParentMerchantId: merchantInfo.RegAcquirerMerchantId,
			MerchantName:             merchantInfo.Name,
			MerchantShortName:        merchantInfo.ShortName,
			AddressRaw:               merchantInfo.AddressRaw,
			BusinessInfoRaw:          []byte(`{}`),
			BusinessDocumentRaw:      []byte(`{}`),
			Status:                   constant.FillingFormReg,
			CreatedAt:                time.Now().UTC(),
			CreatedBy:                request.CreatedBy,
			UpdatedAt:                time.Now().UTC(),
		}
		if err = s.repository.InitRegistration(ctx, registration); err != nil {
			s.logger.Error(ctx, "init registration", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}

		merchantInfo.RegId = registration.Id
	}

	chanErr := make(chan error, 1)

	ctxwc, cancel := context.WithCancel(ctx)
	defer cancel()

	p, err := ants.NewPoolWithFunc(uploadWorker, func(data interface{}) {
		if ctxwc.Err() != nil {
			return
		}
		chanErr <- s.uploadDocument(
			ctx, merchantInfo.RegId, request.Acquirer, data.(merchant.QrisDocument),
		)
	})
	if err != nil {
		s.logger.Error(ctx, "failed to create worker pool", logger.Error(err))
		return "", fmt.Errorf(constant.InternalErrorFmt, traceId)
	}
	defer p.Release()

	mergeFiles := []pdf.MergeFile{}
	for _, doc := range merchantInfo.Documents {
		if doc.Type == certificateEstablishmentType {
			mergeFiles = append(mergeFiles, pdf.MergeFile{
				From:     pdf.GCSFile,
				Bucket:   doc.Location.Bucket,
				Location: doc.Location.Object,
			})
			continue
		}

		p.Invoke(doc)
	}

	for _, bod := range merchantInfo.BoardOfDirectors {
		mergeFiles = append(mergeFiles, pdf.MergeFile{
			From:     pdf.GCSFile,
			Bucket:   bod.IdentityFile.Bucket,
			Location: bod.IdentityFile.Object,
		})
	}

	go func() {
		mergedRaw, err := s.pdf.MergeFilesToPDF(ctx, mergeFiles)
		if err != nil {
			chanErr <- fmt.Errorf("merge file to pdf (mf): %w", err)
			return
		}

		objectName := fmt.Sprintf(
			"%s/%s/%s", s.config.GCSConfig.MerchantDocumentFolderName, merchantInfo.Id, fmt.Sprintf("%s-And-BOD.pdf", certificateEstablishmentType),
		)

		gcsResult, err := s.gcs.UploadFile(ctx, objectName, bytes.NewReader(mergedRaw), true)
		if err != nil {
			chanErr <- fmt.Errorf("upload file to gcs (mf): %w", err)
			return
		}

		p.Invoke(merchant.QrisDocument{
			Type:   certificateEstablishmentType,
			Number: "-",
			Location: merchant.DocLocation{
				Bucket: gcsResult.Bucket,
				Object: gcsResult.ObjectName,
			},
		})
	}()

	for i := 0; i < len(merchantInfo.Documents); i++ {
		if err := <-chanErr; err != nil {
			s.logger.Error(ctx, "upload merchant document", logger.Error(err))
			return "", err
		}
	}

	finalRegistration := &snapCoreModel.RegistrationReq{
		RegistrationId:   merchantInfo.RegId,
		Acquirer:         request.Acquirer,
		MerchantId:       merchantInfo.ExternalId,
		MerchantType:     request.MerchantType,
		ParentMerchantId: merchantInfo.RegAcquirerMerchantId,
		MerchantName:     merchantInfo.Name,
		Address: snapCoreModel.Address{
			ProvinceId: merchantInfo.Address.Province,
			CityId:     merchantInfo.Address.City,
			DistrictId: merchantInfo.Address.District,
			PostCode:   merchantInfo.Address.PostCode,
			Detail:     merchantInfo.Address.Detail,
		},
		BusinessShortname: merchantInfo.ShortName,
		MCC:               merchantInfo.MCC,
	}

	ctx = context.WithValue(
		ctx, constant.CtxClientReqKey, &outbound.Client{
			RequestId:   traceId,
			From:        "QRIS-Registration",
			OriginId:    merchantInfo.RegId,
			ReferenceId: request.MerchantId,
		})
	if err = s.snapRepo.QrFinalRegistration(ctx, finalRegistration); err != nil {
		s.logger.Error(ctx, "qris final registrations", logger.Error(err))
		return "", err
	}
	if err = s.repository.UpdateRegistrationStatus(ctx, merchantInfo.RegId, constant.SubmittedReg); err != nil {
		s.logger.Error(ctx, "update registration status", logger.Error(err))
		return "", fmt.Errorf(constant.InternalErrorFmt, traceId)
	}
	return merchantInfo.RegId, nil
}

func (s *qrisService) uploadDocument(ctx context.Context, id, acquirer string, doc merchant.QrisDocument) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/uploadDocument")
	defer segment.End()

	req := &snapCoreModel.UploadDocumentReq{
		RegistrationId: id,
		Acquirer:       acquirer,
		DocumentType:   doc.Type,
		DocumentNumber: doc.Number,
		ObjectName:     filepath.Base(doc.Location.Object),
	}
	req.RawFile, err = s.gcs.ReadAll(ctx, doc.Location.Bucket, doc.Location.Object)
	if err != nil {
		return err
	}

	resp, err := s.snapRepo.QrUploadDocument(ctx, req)
	if err != nil {
		return err
	}

	uploaded := &qris.UpdateDocument{
		Type:   doc.Type,
		Number: doc.Number,
		Media: qris.Media{
			Internal: qris.Bucket{
				Bucket: doc.Location.Bucket,
				Object: doc.Location.Object,
			},
			External: resp.MediaId,
		},
	}
	return s.repository.UpdateUploadedDocument(ctx, id, uploaded)
}

func (s *qrisService) CreateManualRegistration(ctx context.Context, request *qris.RegistrationReq) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/createManualRegistration")
	defer segment.End()

	merchantInfo, err := s.merchantValidation(ctx, request)
	if err != nil {
		return "", err
	}

	registration := &qris.Registration{
		Id:                       fmt.Sprintf("%s%d", merchantInfo.MID, time.Now().UTC().UnixNano()),
		ExternalId:               merchantInfo.ExternalId,
		Acquirer:                 request.Acquirer,
		MerchantType:             request.MerchantType,
		AcquirerParentMerchantId: merchantInfo.RegAcquirerMerchantId,
		MerchantName:             merchantInfo.Name,
		MerchantShortName:        merchantInfo.ShortName,
		AddressRaw:               merchantInfo.AddressRaw,
		BusinessInfoRaw:          []byte(`{}`),
		BusinessDocumentRaw:      []byte(`{}`),
		Status:                   constant.FillingFormReg,
		CreatedAt:                time.Now().UTC(),
		CreatedBy:                request.CreatedBy,
		UpdatedAt:                time.Now().UTC(),
	}

	if err = s.repository.InitRegistration(ctx, registration); err != nil {
		s.logger.Error(ctx, "init registration", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, ctx.Value(pdkConst.CtxTraceIdKey).(string)))
	}

	return registration.Id, nil
}

func (s *qrisService) UpdateQrRegistration(ctx context.Context, id string, acquirerMerchantId string, acquirerTerminalId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/UpdateQrRegistration")
	defer segment.End()

	return s.repository.UpdateQrRegistration(ctx, id, acquirerMerchantId, acquirerTerminalId)
}

func (s *qrisService) merchantValidation(ctx context.Context, request *qris.RegistrationReq) (*merchant.QrisMerchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/merchantValidation")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	data, err := s.merchantRepo.FindMerchantForQrRegistration(ctx, request.MerchantId, request.Acquirer)
	if err != nil {
		s.logger.Error(ctx, "find merchant for qr registration", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if data == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIDNotValid)

	} else if data.RegStatus == constant.SubmittedReg {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("registration has been submitted"))

	} else if data.RegStatus == constant.SuccessReg {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("merchant is already registered"))
	}

	data.Type = request.MerchantType

	if err = s.validator.ScanStruct(data); err != nil {
		return nil, err
	}

	docs := map[string]bool{}
	maps.Copy(docs, documents[request.Type])
	for _, doc := range data.Documents {
		delete(docs, doc.Type)
	}
	if len(docs) > 0 {
		docRequired := validation.Fields{
			"documents": map[string]string{},
		}
		for name := range docs {
			docRequired["documents"].(map[string]string)[name] = "document required"
		}
		return nil, &docRequired
	}
	if request.Type == constant.EnterpriseType {
		bodErrs := validation.Fields{}
		if data.BODCount == 0 {
			bodErrs["bod"] = "board of directors is required"
		}
		if data.BOCCount == 0 {
			bodErrs["boc"] = "board of commissioner is required"
		}
		if len(bodErrs) > 0 {
			return nil, &bodErrs
		}
	}

	// Validate MCC is required for QRIS registration
	if data.MCC == "" && request.Acquirer == constant.BANK_ACQUIRER_BNC {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("mcc is required"))
	}

	return data, nil
}
