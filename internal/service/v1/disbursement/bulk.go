package disbursementService

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
)

const (
	sheetNameToUpload  = "Template"
	maxRowDataToUpload = 1000

	columnReferenceID            = 0
	columnAmount                 = 1
	columnChannelCode            = 2
	columnBeneficiaryAccountNo   = 3
	columnBeneficiaryAccountName = 4
	columnRemark                 = 5
)

var (
	bulkUploadHeaders = map[int]string{
		columnReferenceID:            "Reference ID",
		columnAmount:                 "Amount",
		columnChannelCode:            "Channel Code",
		columnBeneficiaryAccountNo:   "Account Number",
		columnBeneficiaryAccountName: "Account Name",
		columnRemark:                 "Remarks",
	}

	bankDB = bankTransfer.NewBankDB()
)

func (*DisbursementService) getRowsAndValidateBulkUpload(f xlsx.Filer) ([][]string, error) {
	rows, err := f.GetRows(sheetNameToUpload, xlsx.Options{RawCellValue: true})
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("sheet to upload not found"))

	} else if len(rows) < 2 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty data to upload"))

	} else if len(rows) > maxRowDataToUpload+1 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("max row data is 1000"))
	}
	// Headers validation
	for idx, name := range rows[0] {
		if _, ok := bulkUploadHeaders[idx]; !ok {
			continue

		} else if bulkUploadHeaders[idx] != strings.TrimSpace(name) {
			return nil, constant.ErrHeaderColumnDoesNotMatchWithTemplate
		}
	}
	return rows, nil
}

func (s *DisbursementService) singleRowValidation(ctx context.Context, merchantID string, trxConfig *disbursementModel.TransactionConfig, referenceList map[string]bool, row []string) *disbursementModel.BulkPreviewResponse {
	var (
		result             = constant.BulkPreviewResultValid
		collectErrors      []string
		bankCode, bankName string
	)

	if len(row) == 0 {
		return nil

	} else if len(row) < len(bulkUploadHeaders) {
		row = append(row, make([]string, len(bulkUploadHeaders)-len(row))...)
	}

	referenceID := row[columnReferenceID]
	amountStr := row[columnAmount]
	amount, amountErr := strconv.ParseFloat(amountStr, 64)
	channelCode := row[columnChannelCode]
	beneficiaryAccountNo := row[columnBeneficiaryAccountNo]
	beneficiaryAccountName := row[columnBeneficiaryAccountName]
	remark := row[columnRemark]
	_, fraction := math.Modf(amount)

	if referenceID == "" && channelCode == "" && beneficiaryAccountNo == "" && beneficiaryAccountName == "" && amountStr == "" && remark == "" {
		return nil
	}

	if referenceID == "" {
		collectErrors = append(collectErrors, "Reference ID is required")
	}

	if channelCode == "" {
		collectErrors = append(collectErrors, "Channel code is required")

	} else {
		if bank := bankDB.FindByChannelCode(channelCode); bank != nil {
			bankCode, bankName = bank.Code, bank.Name
		} else {
			collectErrors = append(collectErrors, "Channel code not found")
		}
	}

	if beneficiaryAccountNo == "" {
		collectErrors = append(collectErrors, "Account number is required")

	} else if !util.IsNumericValue(beneficiaryAccountNo) {
		collectErrors = append(collectErrors, "Account number must be numeric value")
	}

	if beneficiaryAccountName == "" {
		collectErrors = append(collectErrors, "Account name is required")
	}

	if amountStr == "" {
		collectErrors = append(collectErrors, "Amount is required")

	} else if amountErr != nil || fraction != 0 {
		collectErrors = append(collectErrors, "Decimal amounts are not allowed")

	} else if amount < trxConfig.MinAmount {
		collectErrors = append(collectErrors, fmt.Sprintf("Min amount is Rp %s", util.ConvertFloatToCurrency(trxConfig.MinAmount)))
	}

	if len(beneficiaryAccountName) > constant.DisbursementMaxLengthBeneficiaryName {
		collectErrors = append(collectErrors, fmt.Sprintf("Max account name %d", constant.DisbursementMaxLengthBeneficiaryName))
	}

	if len(remark) > constant.DisbursementMaxLengthRemark {
		// if remark more than 40, then it will be truncated
		remark = remark[:constant.DisbursementMaxLengthRemark]
	}

	if s.isExistReferenceID(ctx, merchantID, referenceID, referenceList) {
		collectErrors = append(collectErrors, constant.ErrDisbursementReferenceIdAlreadyExist.Error())
	}

	if len(collectErrors) > 0 {
		result = constant.BulkPreviewResultInvalid
	}

	return &disbursementModel.BulkPreviewResponse{
		ReferenceID:            referenceID,
		BeneficiaryBankCode:    bankCode,
		BeneficiaryBankName:    bankName,
		BeneficiaryAccountNo:   beneficiaryAccountNo,
		BeneficiaryAccountName: beneficiaryAccountName,
		Amount:                 amountStr,
		Remark:                 remark,
		Result:                 result,
		Error:                  strings.Join(collectErrors, ", "),
		ChannelCode:            channelCode,
	}
}

func (s *DisbursementService) isExistReferenceID(ctx context.Context, merchantID, referenceID string, referenceList map[string]bool) bool {
	if _, ok := referenceList[strings.ToLower(referenceID)]; ok {
		return true
	}
	return s.IsExistReferenceID(ctx, merchantID, referenceID)
}

func (s *DisbursementService) beneficiaryAccountValidation(ctx context.Context, merchantID string, trxConfig *disbursementModel.TransactionConfig, previewResponse *disbursementModel.BulkPreviewResponse) *disbursementModel.BulkPreviewResponse {
	beneficiaryAccount, err := s.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(
		ctx,
		&beneficiaryAccountModel.CheckAccountRequest{
			BeneficiaryAccountNo: previewResponse.BeneficiaryAccountNo,
			BeneficiaryBankCode:  previewResponse.BeneficiaryBankCode,
			MerchantID:           merchantID,
			AdditionalInfo:       map[string]any{},
		},
	)
	if err != nil || beneficiaryAccount == nil {

		errorAddition := "Account number is invalid"
		if previewResponse.Error != "" {
			errorAddition = ", " + errorAddition
		}

		previewResponse.Result = constant.BulkPreviewResultInvalid
		previewResponse.Error = previewResponse.Error + errorAddition

		return previewResponse
	}

	if previewResponse.BeneficiaryAccountName == "" {
		previewResponse.Result = constant.BulkPreviewResultInvalid
		previewResponse.Error = "Account name is required"
		return previewResponse
	}

	// check name match
	if !strings.EqualFold(beneficiaryAccount.BeneficiaryAccountName, previewResponse.BeneficiaryAccountName) {
		previewResponse.Result = constant.BulkPreviewResultWarning

		errorAddition := fmt.Sprintf("Incorrect beneficiary name. Before : <b>%s</b>", previewResponse.BeneficiaryAccountName)
		if previewResponse.Error != "" {
			errorAddition = ", " + errorAddition
			previewResponse.Result = constant.BulkPreviewResultInvalid
		}
		previewResponse.Error = previewResponse.Error + errorAddition
		previewResponse.BeneficiaryAccountName = beneficiaryAccount.BeneficiaryAccountName
	}

	err = s.validateAmountAndLimits(ctx, merchantID, previewResponse, beneficiaryAccount, trxConfig)
	if err != nil {
		errorAddition := err.Error()
		if previewResponse.Error != "" {
			errorAddition = ", " + err.Error()
		}

		previewResponse.Result = constant.BulkPreviewResultInvalid
		previewResponse.Error += errorAddition
	}

	// Check if the payout destination is a Pivot internal VA. If it matches the internal VA pattern, the transaction is declined.
	if beneficiaryAccount.MetadataObj.IsVirtualAccount {
		if !constant.IsPayoutToVirtualAccountAllowed(beneficiaryAccount.BeneficiaryBankCode, beneficiaryAccount.BeneficiaryAccountNo) {
			errorAddition := constant.ErrDetailMsgPayoutDstNotEligible
			if previewResponse.Error != "" {
				errorAddition = ", " + errorAddition
			}
			previewResponse.Result = constant.BulkPreviewResultInvalid
			previewResponse.Error += errorAddition
		}
	}

	return previewResponse
}

func (s *DisbursementService) validateAmountAndLimits(ctx context.Context, merchantID string, previewResponse *disbursementModel.BulkPreviewResponse, beneficiaryAccount *beneficiaryAccountModel.Account, trxConfig *disbursementModel.TransactionConfig) error {
	var (
		isOverbooking bool
	)

	bank := bankDB.FindByChannelCode(previewResponse.BeneficiaryBankCode)
	if bank != nil {
		isOverbooking = s.IsBankcodeOverbookingChannelAllowed(ctx, bank.Code, merchantID)
	}

	amount, _ := strconv.ParseFloat(previewResponse.Amount, 32)
	maxAmount := util.ValueOfPtr(trxConfig).MaxAmount

	if isOverbooking {
		maxAmount = s.config.DisbursementConfig.OverbookingBankMaxAmount

		if max, isAllow := s.IsMerchantAllowedExcludeBeneficiaryRules(ctx, merchantID, amount); isAllow {
			maxAmount = max
		}

		if s.IsMerchantAllowedToUseBeneficiaryCustomRule(ctx, merchantID, beneficiaryAccount.MetadataObj.BeneficiaryPayoutLimitRule != nil) {
			maxAmount = s.config.DisbursementConfig.OverbookingBankMaxAmountForCustomRule
		}
	}

	// another validation already done in singleRowValidation
	// because max amount limit was coupled with beneficiary data and whitelist merchant
	if amount > maxAmount {
		return fmt.Errorf("Maximum amount is Rp %s", util.ConvertFloatToCurrency(maxAmount))
	}

	return nil
}
