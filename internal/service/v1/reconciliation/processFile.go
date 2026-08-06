package reconciliation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	reconModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

// ProcessFile implements service.IReconciliationService.
func (s *ReconciliationService) ProcessFile(ctx context.Context, id string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/ProcessFile")
	defer segment.End()

	reconDetail, err := s.reconRepo.GetByUUID(ctx, id)
	if err != nil {
		if err == constant.ErrDataNotFound {
			return nil
		}
		return err
	}
	// fetch file from gcs
	rawFile, err := s.gcs.ReadAll(ctx, s.config.GCSConfig.ServiceBucketName, reconDetail.FilePath)
	if err != nil {
		s.logger.Error(ctx, "Failed to get file from gcs", logger.Error(err))
		return err
	}

	f, err := s.excel.OpenReader(bytes.NewBuffer(rawFile))
	if err != nil {
		s.logger.Error(ctx, "Failed to open reconciliation file reader", logger.Error(err))
		return err
	}
	defer f.Close()

	// get rows data from excel
	rows, err := s.getRowsAndValidateBulkUpload(f)
	if err != nil {
		return err
	}

	// get date range from excel data for bulk update
	tempResult := &reconModel.ReconResult{Transactions: make([]*reconciliation.Transaction, 0)}
	for i, row := range rows[1:] {
		uploadedTransaction, err := convertRowsToUploadedTransaction(row, f, i)
		if err == nil {
			tempResult.Transactions = append(tempResult.Transactions, &uploadedTransaction)
		}
	}

	// Begin database transaction for consistent recon processing
	txCtx, err := s.accountTransactionRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to begin transaction for recon processing", logger.Error(err))
		return err
	}

	// Ensure transaction is handled properly
	defer func() {
		if r := recover(); r != nil {
			s.accountTransactionRepo.RollbackTransaction(txCtx)
			panic(r)
		}
	}()

	// update recon status to review for transaction not in recon file FIRST
	firstDate, lastDate := tempResult.GetFirstLatestDate()
	if firstDate != nil && lastDate != nil {
		scanningToleranceInDays := s.config.ReconConfig.ScanningToleranceInDays
		if scanningToleranceInDays == 0 {
			scanningToleranceInDays = 5
		}

		if err := s.accountTransactionRepo.UpdateBulkReconStatus(txCtx, &reconciliation.BulkUpatedStatus{
			StartTime:               *firstDate,
			EndTime:                 *lastDate,
			Status:                  constant.ReconStatusReview,
			TrxReference:            reconDetail.GetTransactionReferenceByTransactionType(),
			TrxType:                 reconDetail.TransactionType,
			ScanningToleranceInDays: scanningToleranceInDays,
		}); err != nil {
			s.logger.Error(txCtx, "Failed to bulk update recon status", logger.Error(err))
			s.accountTransactionRepo.RollbackTransaction(txCtx)
			return err
		}
	}

	// process row
	reconResult := s.processRows(txCtx, reconDetail.GetTransactionReferenceByTransactionType(), reconDetail.TransactionType, rows[1:], f)

	if reconResult.ShouldReconcileVAStatic() {
		// check total value of VAStatic
		totalByReference, err := s.accountTransactionRepo.GetTotalPaymentAmount(txCtx, &reconciliation.PaymentTotalAmountQuery{
			ReferenceIDs: reconResult.VAStatic.Keys(),
			StartTime:    *firstDate,
			EndTime:      *lastDate,
			Channel:      constant.ChannelVirtualAccount,
		})
		if err != nil {
			s.logger.Error(txCtx, "Failed to get total payment amount", logger.Error(err))
			s.accountTransactionRepo.RollbackTransaction(txCtx)
			return err
		}
		for key, vaStatic := range *reconResult.VAStatic {
			// check totalAmount equality
			if !totalByReference.GetTotalAmount(key).Equal(vaStatic.TotalAmount) {
				for _, index := range vaStatic.Indexes {
					reconResult.Transactions[index].Status = constant.ReconStatusReview
					reconResult.Transactions[index].Reason = "total amount not match"
				}
			} else {
				// update status in database to success
				for _, trxIds := range vaStatic.UUIDs {
					err := s.accountTransactionRepo.UpdateReconDetail(txCtx, trxIds, &reconciliation.ReconDetail{
						Status:   constant.ReconStatusSuccess,
						DateTime: util.SnapCompatible(time.Now()),
					})
					if err != nil {
						s.logger.Error(txCtx, "Failed to update recon detail", logger.Error(err))
						s.accountTransactionRepo.RollbackTransaction(txCtx)
						return err
					}
				}
			}
		}
	}

	// Commit transaction if everything succeeded
	if err := s.accountTransactionRepo.CommitTransaction(txCtx); err != nil {
		s.logger.Error(txCtx, "Failed to commit recon transaction", logger.Error(err))
		s.accountTransactionRepo.RollbackTransaction(txCtx)
		return err
	}
	if err := s.fillResultSheet(f, reconResult.Transactions); err != nil {
		return err
	}

	objectName := filepath.Join(
		constant.ReconciliationResultDir,
		fmt.Sprintf(
			constant.DefaultFilenameResultReconciliation,
			util.GetCurrentTimeWithMillisFormatted(),
		),
	) + constant.DefaultExtXlsx

	updatedFile, err := f.WriteToBuffer()
	if err != nil {
		s.logger.Error(ctx, "Failed to write file to buffer", logger.Error(err))
		return err
	}

	gcsFilePath, err := s.gcs.UploadFile(ctx, objectName, updatedFile, true)
	if err != nil {
		s.logger.Error(ctx, "Failed upload reconciliation file to GCS", logger.Error(err))
		return pkgErrs.New(response.HttpErrInternal, err)
	}

	reconDetail.ResultFilePath = gcsFilePath.ObjectName
	reconDetail.Status = reconResult.Status
	if reconResult.ErrorReason != "" {
		reconDetail.Reasons = sql.NullString{
			Valid:  true,
			String: reconResult.ErrorReason,
		}
	}
	if err := s.reconRepo.Update(ctx, reconDetail); err != nil {
		s.logger.Error(ctx, "Failed to update reconciliation detail", logger.Error(err))
		return err
	}

	return nil
}

// processRows processes the rows of a reconciliation file, checking the transactions against the backend portal,
// updating the transaction status and metadata, and returning a slice of reconciliation transactions.
// The function takes a context and a slice of string slices representing the rows of the reconciliation file.
// It returns a slice of pointers to reconciliation.Transaction structs, representing the processed transactions.
func (s *ReconciliationService) processRows(ctx context.Context, transactionReference, transactionType string, rows [][]string, f xlsx.Filer) *reconModel.ReconResult {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/processRows")
	defer segment.End()

	execTime := util.SnapCompatible(time.Now())
	reconTransacations := make([]*reconciliation.Transaction, len(rows))
	reconStatus := constant.StatusSuccess
	errReason := ""
	errCounter := 0
	checkedRows := map[string]int{}
	checkedVAStaticRows := reconModel.ReconVAStatic{}

	for i, row := range rows {
		failedReason := ""
		status := constant.ReconStatusSuccess

		uploadedTransactions, err := convertRowsToUploadedTransaction(row, f, i)
		uploadedTransactions.ReconTime = execTime
		uploadedTransactions.Order = i
		uploadedTransactions.TransactionReference = transactionReference
		uploadedTransactions.TransactionType = transactionType

		// check duplicate row data
		rowCheck := strings.ToLower(strings.Join(row, ","))
		rowSha := sha256.Sum256([]byte(rowCheck))
		rowShaStr := fmt.Sprintf("%x", rowSha)
		if val, ok := checkedRows[rowShaStr]; ok {
			// check duplicate data entry
			uploadedTransactions.Status = constant.ReconStatusReview
			uploadedTransactions.Reason = "duplicate row with data at row " + fmt.Sprint(val+2) // data row in excel start at row 2 not 0
			reconTransacations[i] = &uploadedTransactions
			errCounter += 1
			continue
		}

		if err != nil {
			reconStatus = constant.StatusFailed
			uploadedTransactions.Status = constant.ReconStatusReview
			uploadedTransactions.Reason = err.Error()
			errReason = err.Error()
			reconTransacations[i] = &uploadedTransactions
			errCounter += 1

			s.logger.Warn(ctx, "Failed to convert row to uploaded transaction", logger.Error(err), logger.Any("row", row))
			continue
		}
		accountTrx, err := s.checkUploadedTransaction(ctx, &uploadedTransactions)
		if err != nil && accountTrx == nil {
			reconStatus = constant.StatusFailed
			actualErrReason := err.Error()
			if actualErrReason == "transaction not found" {
				actualErrReason = "transaction not found in backend portal"
			}
			uploadedTransactions.Status = constant.ReconStatusReview
			uploadedTransactions.Reason = actualErrReason
			reconTransacations[i] = &uploadedTransactions
			errCounter += 1
			continue
		}

		if accountTrx == nil && err == nil {
			uploadedTransactions.Status = constant.ReconStatusSuccess
			uploadedTransactions.Merchant = ""
			uploadedTransactions.ReconTime = execTime
			uploadedTransactions.Reason = ""
			uploadedTransactions.Order = i
			reconTransacations[i] = &uploadedTransactions
			continue
		}

		isStaticVA := accountTrx != nil &&
			uploadedTransactions.Channel == constant.ChannelVirtualAccount &&
			accountTrx.PaymentType == constant.UnifiedPaymentTypeMultiple

		if isStaticVA {
			amount := uploadedTransactions.Amount
			checkedVAStaticRows.Add(uploadedTransactions.Reference, i, amount, accountTrx)
		} else {
			// only add to checked duplication rows if payment type is not multiple
			checkedRows[rowShaStr] = i
		}

		if err != nil {
			reconStatus = constant.StatusFailed
			status = constant.ReconStatusReview
			failedReason = err.Error()
			errReason = err.Error()
			errCounter += 1
		}

		// update metadata and status (only status success in snap-core) UpdatePaymentTransactionStatusAndMetadataByID
		if !isStaticVA {
			if err := s.accountTransactionRepo.UpdateReconDetail(
				ctx,
				accountTrx.UUID,
				&reconciliation.ReconDetail{
					Status:   status,
					DateTime: execTime,
				},
			); err != nil {
				errReason = "error when update transaction"
				errCounter += 1
				s.logger.Error(ctx, "Failed to update transaction", logger.Error(err))
				continue
			}
		}

		// append transaction
		uploadedTransactions.Status = status
		uploadedTransactions.Merchant = accountTrx.MerchantName
		uploadedTransactions.ReconTime = execTime
		uploadedTransactions.Reason = failedReason
		uploadedTransactions.Order = i
		reconTransacations[i] = &uploadedTransactions
	}

	if errCounter > 1 {
		errReason = "there is multiple errors, check at result file for more details"
	}
	result := reconModel.ReconResult{
		Transactions: reconTransacations,
		Status:       reconStatus,
		ErrorReason:  errReason,
		VAStatic:     &checkedVAStaticRows,
	}

	return &result
}

// checkUploadedTransaction checks if the given transaction exists in the database and its status in SnapCore.
// It returns the corresponding ReconTransactionModel if the transaction is found, or an error if not.
// If the transaction is found but its status is not "success" in SnapCore, it will attempt to check the status
// in SnapCore and update the status accordingly.
func (s *ReconciliationService) checkUploadedTransaction(ctx context.Context, trx *reconciliation.Transaction) (*reconciliation.ReconTransactionModel, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/checkUploadedTransaction")
	defer segment.End()
	duration := time.Minute * 2
	withDuration := false

	if util.Contains([]string{constant.ChannelVirtualAccount, constant.ChannelQris, constant.ChannelCard, constant.ChannelBankTransfer}, trx.Channel) {
		withDuration = true
	}

	scanningToleranceInDays := s.config.ReconConfig.ScanningToleranceInDays
	if scanningToleranceInDays == 0 {
		scanningToleranceInDays = 5
	}

	t := trx.TransactionDate
	startUpdatedAt := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	endUpdatedAt := startUpdatedAt.Add(time.Duration(scanningToleranceInDays) * 24 * time.Hour)

	// check transaction in database
	accountTrx, err := s.accountTransactionRepo.GetTransactionForRecon(ctx, &reconciliation.ReconTransactionQuery{
		ReferenceID:      trx.Reference,
		Amount:           trx.Amount,
		TransactionDate:  trx.TransactionDate,
		Reference:        trx.TransactionReference,
		TransactionType:  trx.TransactionType,
		Channel:          strings.ToUpper(trx.Channel),
		WithTimeDuration: withDuration,
		Duration:         duration,
		SettlementModel:  constant.PaymentMethodChannelTypeAggregator,
		StartUpdatedAt:   startUpdatedAt,
		EndUpdatedAt:     endUpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	// check transaction in snapcore
	isSnapCoreChannel := util.Contains([]string{constant.ChannelVirtualAccount, constant.ChannelQris, constant.ChannelBankTransfer}, trx.Channel)
	if isSnapCoreChannel && (accountTrx == nil || accountTrx.Status == constant.StatusPending) {

		accountTrx, err = s.checkTransactionSnapCore(ctx, trx, trx.Amount)
		if err != nil {
			return accountTrx, err
		}
		if accountTrx == nil && err == nil {
			return nil, nil
		}
	}
	if trx.Channel == constant.ChannelCard &&
		(accountTrx == nil || accountTrx.Status == constant.StatusPending) {

		accountTrx, err = s.checkTransactionCCProcessor(ctx, trx, trx.Amount)
		if err != nil {
			return accountTrx, err
		}
		// If credit card processor validated successfully, trust it completely
		if accountTrx == nil && err == nil {
			return nil, nil
		}
	}

	if accountTrx != nil {
		isAmountValid := accountTrx.Amount.Equal(trx.Amount)
		if !isAmountValid {
			return accountTrx, fmt.Errorf("transaction amount mismatch")
		}

		if util.Contains([]string{constant.StatusPending, constant.StatusFailed}, accountTrx.Status) {
			if accountTrx.ReasonDesc.Valid && accountTrx.ReasonDesc.String != "" {
				return accountTrx, fmt.Errorf("%s", accountTrx.ReasonDesc.String)
			}
			return accountTrx, fmt.Errorf("transaction %s", strings.ToLower(accountTrx.Status))
		}

		return accountTrx, nil
	}

	return nil, fmt.Errorf("transaction not found")
}

// fillResultSheet writes the reconciliation transaction data to an XLSX file sheet.
// The sheet is named "Result" and contains the following columns:
// - Transaction date & time
// - Recon date & time
// - Merchant name
// - Transaction Reference
// - Transaction amount
// - Bank
// - Channel
// - Recon status
// - Recon reason
func (*ReconciliationService) fillResultSheet(f xlsx.Filer, transactions []*reconciliation.Transaction) error {
	sheetName := "Result"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	header := []string{
		"Transaction date & time", "Recon date & time", "Merchant name", "Transaction reference used for recon", "Transaction reference used for recon_2", "Transaction amount", "Bank", "Channel", "Recon status", "Recon reason",
	}
	if err := f.SetSheetRow(sheetName, "A1", &header); err != nil {
		return err
	}
	rowDataNumber := 2
	for _, trx := range transactions {
		if trx == nil {
			continue
		}
		rowData := []any{
			trx.TransactionDate.Format(constant.ReconTimeFormat),
			trx.ReconTime,
			trx.Merchant,
			trx.Reference,
			trx.Reference2,
			trx.Amount,
			trx.Bank,
			trx.Channel,
			trx.Status,
			trx.Reason,
		}
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowDataNumber), &rowData); err != nil {
			return err
		}
		rowDataNumber++
	}

	return nil
}

// checkTransactionSnapCore checks the status of a reconciliation transaction in the Snap Core service.
// It takes a context and a reconciliation transaction model as input, and returns a TransactionCheckResponse.
// The function first determines the transaction type based on the channel, then calls the Snap Core repository to check the transaction status.
// If the transaction is found in Snap Core with a successful status, the function returns a successful TransactionCheckResponse.
// If the transaction is found in Snap Core with a failed status, the function returns a failed TransactionCheckResponse.
// If there is an error checking the transaction in Snap Core, the function returns a failed TransactionCheckResponse with the error reason.
func (s *ReconciliationService) checkTransactionSnapCore(ctx context.Context, req *reconciliation.Transaction, amount decimal.Decimal) (*reconciliation.ReconTransactionModel, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/checkTransactionSnapCore")
	defer segment.End()
	var trxType snapCoreModel.TransactionType
	var referenceType = constant.TypePayment
	response := reconciliation.TransactionCheckResponse{}

	switch req.Channel {
	case constant.ChannelVirtualAccount:
		trxType = snapCoreModel.TypeVA
	case constant.ChannelQris:
		trxType = snapCoreModel.TypeQRIS
	case constant.ChannelBankTransfer:
		referenceType = constant.TypeDisbursement
		trxType = snapCoreModel.TypeBankTransfer
	}

	resp, err := s.snapCoreRepo.CheckReconTransaction(ctx, &snapCoreModel.AutoReconTrxRequest{
		Type:         trxType,
		ReferenceNo:  req.Reference,
		ReferenceNo2: req.Reference2,
		TrxTimestamp: &req.TransactionDate,
		Amount:       amount,
		Bank:         req.Bank,
	})

	if err != nil {
		response.Status = constant.StatusFailed
		response.Reason = "error when check transaction to snap core"
		s.logger.Warn(ctx, "Failed to check transaction to snap core", logger.Error(err))
		return nil, errors.New(response.Reason)
	}

	if resp.Status != constant.ReconSnapStatusValid {
		response.Status = constant.StatusFailed
		response.Reason = resp.Message
		if response.Reason == "" {
			response.Reason = resp.Code.Message()
		}
		return nil, errors.New(response.Reason)
	}

	if resp.Code == constant.ReconCodeOk {
		accountTrx, err := s.accountTransactionRepo.GetTransactionByProcessorID(ctx, referenceType, constant.SnapCoreProcessor, resp.ProcessorReferenceID)
		if err == nil && accountTrx != nil {
			s.logger.Info(ctx, "Snap-core validation successful, transaction found in backend portal, start updating reconciliation info", logger.String("processor_reference_id", resp.ProcessorReferenceID))
			if updateErr := s.accountTransactionRepo.SetAdditionalInfoReconciliation(
				ctx,
				accountTrx.UUID,
				&reconciliation.ReconDetail{
					Status:   constant.ReconStatusSuccess,
					DateTime: time.Now().UTC().Format(time.RFC3339),
				},
			); updateErr != nil {
				s.logger.Error(ctx, "Failed to update recon detail for successful snap-core validation", logger.Error(updateErr))
			}
		}

		return nil, nil
	}

	response.Status = constant.StatusFailed
	response.Reason = fmt.Sprintf("snap-core returned VALID status but unexpected code: %s", resp.Code)
	return nil, errors.New(response.Reason)
}

func (s *ReconciliationService) checkTransactionCCProcessor(ctx context.Context, req *reconciliation.Transaction, amount decimal.Decimal) (*reconciliation.ReconTransactionModel, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/checkTransactionCCProcessor")
	defer segment.End()
	response := reconciliation.TransactionCheckResponse{}

	resp, err := s.creditCardCoreProcessorRepo.CheckReconTransaction(ctx, &creditcardCoreProcessorModel.AutoReconTrxRequest{
		Type:         req.Channel,
		ReferenceNo:  req.Reference,
		ReferenceNo2: req.Reference2,
		TrxTimestamp: &req.TransactionDate,
		Amount:       amount,
		Bank:         req.Bank,
	})

	if err != nil {
		response.Status = constant.StatusFailed
		response.Reason = "error when check transaction to credit card core processor"
		s.logger.Warn(ctx, "Failed to check transaction to credit card core processor", logger.Error(err))
		return nil, errors.New(response.Reason)
	}

	if resp.Status != constant.ReconCCStatusValid {
		response.Status = constant.StatusFailed
		response.Reason = "transaction found in credit card core processor with status failed"
		return nil, errors.New(response.Reason)
	}
	accountTrx, err := s.accountTransactionRepo.GetTransactionByProcessorID(ctx, constant.TypePayment, constant.CreditCardCoreProcessor, resp.ProcessorReferenceID)
	if err != nil || accountTrx == nil {
		response.Status = constant.StatusFailed
		response.Reason = "error when get transaction from account transaction"
		s.logger.Warn(ctx, "Failed to get transaction from account transaction", logger.Error(err))
		return nil, errors.New(response.Reason)
	}
	if resp.Code == constant.ReconCCCodeOk {
		accountTrx.Status = constant.StatusSuccess
	}
	return accountTrx, nil
}
