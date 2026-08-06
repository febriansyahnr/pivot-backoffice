package disbursementService

import (
	"context"
	"errors"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) BulkPreview(ctx context.Context, request *disbursementModel.BulkPreviewRequest) (previewResult []disbursementModel.BulkPreviewResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/BulkPreview")
	defer segment.End()

	f, err := s.excel.OpenReader(request.File)
	if err != nil {
		s.logger.Error(ctx, "Failed to open reader file", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrOpenFileReader)
	}
	defer f.Close()

	rows, err := s.getRowsAndValidateBulkUpload(f)
	if err != nil {
		return nil, err
	}

	trxConfig, _, err := s.GetMerchantTransactionConfig(ctx, request.MerchantId)
	if err != nil {
		return nil, err
	}

	referenceList := map[string]bool{}

	for _, row := range rows[1:] {
		bulkRequest := s.singleRowValidation(ctx, request.MerchantId, trxConfig, referenceList, row)
		if bulkRequest != nil {
			previewResult = append(previewResult, *bulkRequest)
		}
		if len(row) > columnReferenceID {
			referenceList[strings.ToLower(row[columnReferenceID])] = true
		}
	}

	if len(previewResult) == 0 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty data to upload"))
	}
	return
}
