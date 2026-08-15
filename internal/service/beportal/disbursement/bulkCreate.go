package disbursementService

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

func (s *DisbursementService) BulkCreate(ctx context.Context, request *disbursementModel.BulkCreateRequest) (*disbursementModel.BulkCreateResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/BulkCreate")
	defer segment.End()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "Failed find merchant by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	merchantConfigID := request.MerchantId
	if merchant.ParentID.String != "" { // Sub-Merchant
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)

		if merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			merchantConfigID = merchant.ParentID.String
		}
	}

	rawFile, _ := io.ReadAll(request.File)
	defer func() { rawFile = nil }()

	f, err := s.excel.OpenReader(bytes.NewBuffer(rawFile))
	if err != nil {
		s.logger.Error(ctx, "Failed to open file reader", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrOpenFileReader)
	}
	defer f.Close()

	rows, err := s.getRowsAndValidateBulkUpload(f)
	if err != nil {
		return nil, err
	}

	queueKeys, isCompleted := []string{}, false
	defer func() {
		if !isCompleted && len(queueKeys) > 0 {
			if e := s.redisExt.Del(ctx, queueKeys...).Err(); e != nil {
				s.logger.Error(ctx, "clears the process queue lock list", logger.Error(e))
			}
		}
	}()

	trxConfig, err := s.GetTransactionConfig(ctx, merchantConfigID)
	if err != nil {
		return nil, err
	}

	var (
		referenceList                = map[string]bool{}
		invalidRow                   = []*disbursementModel.BulkPreviewResponse{}
		validDataCreateDisbursements = []disbursementModel.CreateSingleRequest{}

		totalData, totalAmount = 0, 0.0
		traceId, _             = ctx.Value(pdkConst.CtxTraceIdKey).(string)
		queueTTLLock           = time.Duration(s.config.AppConfig.BulkDisbursementExpireLockMinute) * time.Minute
	)

	for _, row := range rows[1:] {
		bulkRequest := s.singleRowValidation(ctx, request.MerchantId, trxConfig, referenceList, row)

		if bulkRequest == nil {
			continue
		}

		if len(row) > columnReferenceID {
			referenceList[strings.ToLower(row[columnReferenceID])] = true
		}

		bulkRequest = s.beneficiaryAccountValidation(ctx, request.MerchantId, trxConfig, bulkRequest)

		if bulkRequest.Result == constant.BulkPreviewResultInvalid {
			invalidRow = append(invalidRow, bulkRequest)
			continue
		}

		queueKey := fmt.Sprintf(constant.BulkDisbursementQueueLockFmt, merchant.UUID, bulkRequest.ReferenceID)
		if ok, err := s.redisExt.SetNX(ctx, queueKey, true, queueTTLLock).Result(); err != nil {
			s.logger.Error(ctx, "set exclusive queue with key "+queueKey, logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

		} else if !ok {
			return nil, pkgErrs.New(response.HttpErrDupCheck, constant.ErrDisbursementReferenceIdAlreadyExist)
		}

		queueKeys = append(queueKeys, queueKey)

		amount, err := decimal.NewFromString(bulkRequest.Amount)
		if err != nil {
			s.logger.Error(ctx, "failed to convert amount to decimal", logger.Error(err))
			invalidRow = append(invalidRow, &disbursementModel.BulkPreviewResponse{
				Result: constant.BulkPreviewResultInvalid,
				Error:  "Invalid amount type / format",
			})
			continue
		}

		validDataCreateDisbursements = append(validDataCreateDisbursements, disbursementModel.CreateSingleRequest{
			ReferenceID:            bulkRequest.ReferenceID,
			BeneficiaryBankCode:    bulkRequest.BeneficiaryBankCode,
			BeneficiaryBankName:    bulkRequest.BeneficiaryBankName,
			BeneficiaryAccountNo:   bulkRequest.BeneficiaryAccountNo,
			BeneficiaryAccountName: bulkRequest.BeneficiaryAccountName,
			Amount:                 amount,
			Remark:                 bulkRequest.Remark,
			MerchantID:             merchant.UUID,
			MerchantName:           merchant.Name,
			CreatedBy:              &request.CreatedBy,
		})

		amountFloat, _ := strconv.ParseFloat(bulkRequest.Amount, 64)

		totalData += 1
		totalAmount += amountFloat
	}

	if len(validDataCreateDisbursements) == 0 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("there is no valid data"))
	}

	objectName := filepath.Join(
		"disbursements/bulk-transactions", constant.ExportBulkDisbursementSuccessDir,
		fmt.Sprintf(
			constant.DefaultFilenameUploadBulkDisbursement, util.GetCurrentTimeWithMillisFormatted(),
		),
	) + constant.DefaultExtXlsx

	gcsFilePath, err := s.gcs.UploadFile(ctx, objectName, bytes.NewBuffer(rawFile), false)
	if err != nil {
		s.logger.Error(ctx, "Failed upload file to GCS", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	bulkDisbursement := &disbursementModel.BulkDisbursement{
		UUID:       uuid.NewString(),
		MerchantID: request.MerchantId,
		File:       gcsFilePath.PublicURL,
		Status:     constant.BulkDisbursementStatusUploading,
		CreatedBy:  &request.CreatedBy,
		CreatedAt:  time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err = s.disbursementRepo.InsertBulkDisbursement(ctx, bulkDisbursement); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if len(invalidRow) > 0 {

		fileFailedPath, err := s.GenerateExcelAndUpdateInvalidBulkDisbursement(ctx, bulkDisbursement.UUID, invalidRow)
		if err != nil {
			return nil, err
		}
		bulkDisbursement.FileFailed = &fileFailedPath
	}

	wg := sync.WaitGroup{}
	chunkSize := constant.BulkDisbursementMaxDataRequestPerBatch

	// Looping batch, then send to rmq
	for i := 0; i < len(validDataCreateDisbursements); i += chunkSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()

			end := start + chunkSize
			if end > len(validDataCreateDisbursements) {
				end = len(validDataCreateDisbursements)
			}
			batchRequest := &pb.BatchCreateDisbursementRequest{
				BulkId:       bulkDisbursement.UUID,
				MerchantId:   merchant.UUID,
				MerchantName: merchant.Name,
				CreatedBy:    request.CreatedBy,
				CreatedFrom:  constant.DisbursementCreatedFromMerchantPortal,
				TotalTrx:     int64(len(validDataCreateDisbursements)),
				AutoApprove:  false,
				Data:         disbursementModel.TransformArrayCreateSingleRequestToProtobufType(validDataCreateDisbursements[start:end]),
			}
			payload, _ := proto.Marshal(batchRequest)

			_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.BulkDisbursementBatchCreateRoutingKey, nil, payload)
		}(i)
	}

	wg.Wait()
	isCompleted = true

	return &disbursementModel.BulkCreateResponse{
		UUID:        bulkDisbursement.UUID,
		MerchantID:  bulkDisbursement.MerchantID,
		File:        bulkDisbursement.File,
		FileFailed:  bulkDisbursement.FileFailed,
		Status:      bulkDisbursement.Status,
		CreatedBy:   bulkDisbursement.CreatedBy,
		CreatedAt:   bulkDisbursement.CreatedAt,
		UpdatedAt:   bulkDisbursement.UpdatedAt,
		TotalData:   totalData,
		TotalAmount: totalAmount,
	}, nil
}
