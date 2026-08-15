package disbursementService

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) BulkValidate(ctx context.Context, request *disbursementModel.BulkPreviewRequest) (previewResult []disbursementModel.BulkPreviewResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/BulkValidate")
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

	trxConfig, ctx, err := s.GetMerchantTransactionConfig(ctx, request.MerchantId)
	if err != nil {
		return nil, err
	}

	var (
		wg sync.WaitGroup
		mx sync.Mutex

		referenceList = map[string]bool{}
		disbursements = map[string][]*disbursementModel.BulkPreviewResponse{}
	)

	// Uses a pool of workers to run beneficiaryAccountValidation process concurrently
	p, _ := ants.NewPoolWithFunc(s.getBulkValidateWorkers(), func(input interface{}) {
		defer wg.Done()

		accounts := input.([]*disbursementModel.BulkPreviewResponse)
		if len(accounts) == 0 {
			return
		}

		feeRequest := feeModel.GetPayoutTrxFeeRequest{
			MerchantId:   request.MerchantId,
			MerchantType: constant.MerchantTypeMerchant,
			ChannelCode:  accounts[0].ChannelCode,
		}
		if feeRequest.ParentMerchantId, _ = ctx.Value(constant.CtxParentMerchantId).(string); feeRequest.ParentMerchantId != "" {
			feeRequest.MerchantType = constant.MerchantTypeSubMerchant
		}

		for _, bulkRequest := range accounts {

			bulkRequest = s.beneficiaryAccountValidation(ctx, request.MerchantId, trxConfig, bulkRequest)

			feeRequest.TrxAmount, _ = strconv.ParseFloat(bulkRequest.Amount, 64)
			if fee, err := s.feeSvc.GetPayoutTransactionFee(ctx, feeRequest); err == nil {
				bulkRequest.FeeAmount = util.ValueToPtr(fee.ToFeeResponse().TotalAmount)
			}

			mx.Lock()
			previewResult = append(previewResult, *bulkRequest)
			mx.Unlock()
		}
	})
	defer p.Release()

	// Data grouping based on bank code and account number
	for _, row := range rows[1:] {
		bulkRequest := s.singleRowValidation(ctx, request.MerchantId, trxConfig, referenceList, row)
		if bulkRequest != nil {

			key := bulkRequest.BeneficiaryBankCode + bulkRequest.BeneficiaryAccountNo
			disbursements[key] = append(disbursements[key], bulkRequest)
		}

		if len(row) > columnReferenceID {
			referenceList[strings.ToLower(row[columnReferenceID])] = true
		}
	}

	// Distribution of task to workers
	for _, accounts := range disbursements {
		wg.Add(1)

		_ = p.Invoke(accounts)
	}

	wg.Wait()

	if len(previewResult) == 0 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty data to upload"))
	}
	return
}
