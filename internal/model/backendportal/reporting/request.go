package reportingModel

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cdc"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
)

const createdBySystem = "System"

type UpsertBalanceHistoryRequest struct {
	Event *cdcModel.Event[cdcModel.AccountTransaction]
}

func (r *UpsertBalanceHistoryRequest) ShouldExcludeEvent() bool {
	payload := r.Event.GetCurrent()
	reference := util.ValueOfPtr(payload.Reference)

	// amount validation
	// parent of split payment always has zero value in both debit and credit but have summary of the sub-payment transactions in metadata
	if (payload.Debit.Equal(decimal.Zero) && payload.Credit.Equal(decimal.Zero)) && payload.AdditionalInfo.SubPaymentSummary == nil {
		return true
	}

	return reference == constant.ReferencePaymentFundedPayout ||
		reference == constant.ReferenceSubPayment ||
		((r.Event.IsCreate() || r.Event.IsDelete()) && payload.Status != constant.StatusSuccess) ||
		(r.Event.IsUpdate() && r.Event.GetPrevious().Status != constant.StatusSuccess && payload.Status != constant.StatusSuccess) ||
		r.shouldExcludeFee(payload) ||
		payload.GetSettlementModel() == constant.SettlementModelDirect
}

func (r *UpsertBalanceHistoryRequest) shouldExcludeFee(payload *cdcModel.AccountTransaction) bool {
	if payload.Type != constant.TypeFee {
		return false
	}

	excluded := payload.Channel == constant.ChannelVirtualAccount &&
		slices.Contains([]string{constant.ReferenceDisbursement, constant.ReferenceWallet}, util.ValueOfPtr(payload.Reference))
	if excluded {
		return true
	}

	if payload.AdditionalInfo.Notes == "ON-BEHALF" {
		return false
	}

	included := payload.AdditionalInfo.Type != constant.ReferenceDisbursement &&
		payload.AdditionalInfo.Type != constant.ReferenceDisbursementVA &&
		payload.AdditionalInfo.Type != constant.ReferencePayment &&
		payload.AdditionalInfo.Type != constant.ReferenceXB &&
		payload.AdditionalInfo.Type != constant.ReferencePlatformTransaction &&
		payload.AdditionalInfo.Type != constant.ReferenceRefund &&
		!(payload.AdditionalInfo.Type == constant.TypeWallet && payload.AdditionalInfo.ReferenceType == constant.ReferenceTopUp)
	if included {
		return false
	}
	return true
}

func (u *UpsertBalanceHistoryRequest) ToCreateBalanceHistory() BalanceHistory {
	payload := u.Event.GetCurrent()

	result := BalanceHistory{
		TransactionID:     payload.UUID,
		MerchantID:        payload.MerchantID,
		BalanceType:       accountModel.GetAccountNameByUsecase(util.ValueOfPtr(payload.Reference)),
		Type:              payload.Type,
		TransactionType:   payload.Type,
		Channel:           payload.Channel,
		Currency:          payload.Currency,
		Remarks:           payload.Remarks,
		Status:            payload.Status,
		ReasonType:        payload.ReasonType,
		ReasonDescription: payload.ReasonDescription,
		SettlementModel:   payload.GetSettlementModel(),
		CreatedAt:         payload.CreatedAt,
		StatusUpdatedAt:   payload.UpdatedAt,
		SourceID:          payload.ReferenceID,
		SourceAccountID:   payload.AccountID,
		SourceCreatedAt:   &payload.CreatedAt,
		SourceCreatedBy:   createdBySystem,
		IngestedAt:        time.UnixMilli(u.Event.TsMs),
	}

	if payload.MerchantReferenceID != nil {
		result.ReferenceID = *payload.MerchantReferenceID
	}

	switch payload.Type {
	case constant.TypeWithdrawal:
		result.TransactionType = result.BalanceType + "_" + constant.TypeWithdrawal

	case constant.TypeTopUp, constant.TypeMerchantTopUp:
		result.Type = constant.TypeMerchantTopUp
	}

	if payload.Type == constant.TypeGeneralTopUp && payload.Channel == constant.ChannelBalanceTransfer {
		result.Type = constant.TypeWithdrawal
	}

	if payload.Type == constant.TypePayment && util.ValueOfPtr(payload.Reference) == constant.ReferenceVirtualTerminal {
		result.Type = constant.TypeVirtualTerminal
	}

	if payload.Type == constant.TypeDisbursement && util.ValueOfPtr(payload.ReasonType) == constant.ReasonTypeReversal {
		result.Type = constant.TypeReversal
	}

	if payload.Type == constant.TypeFee && util.ValueOfPtr(payload.Reference) != constant.ReferenceWallet && payload.AdditionalInfo.Type != "" {
		result.TransactionType = payload.AdditionalInfo.Type + "_FEE"
	}

	if result.Channel == "" && payload.AdditionalInfo.ReferenceType != "" {
		result.Channel = payload.AdditionalInfo.ReferenceType
	}

	if payload.Debit.GreaterThan(decimal.Zero) {
		result.Amount = payload.Debit.Neg()
	} else if payload.Credit.GreaterThan(decimal.Zero) {
		result.Amount = payload.Credit
	} else if payload.AdditionalInfo.SubPaymentSummary != nil {
		result.Amount = payload.AdditionalInfo.SubPaymentSummary.TotalCreditAmount
	}

	canFetchFeeAmount := payload.AdditionalInfo.FeeDetail != nil &&
		!slices.Contains([]string{constant.TypeFee, constant.TypeFeeRefund, constant.TypeFeeReversal}, payload.Type)
	if canFetchFeeAmount {
		result.Fee = decimal.NewFromFloat(payload.AdditionalInfo.FeeDetail.FinalAmount)
	}

	if payload.SettlementStatus == nil {
		result.SettlementStatus = constant.StatusSuccess
	} else {
		result.SettlementStatus = *payload.SettlementStatus
	}

	if payload.SettlementAt != nil && !payload.SettlementAt.IsZero() {
		result.SettlementAt = *payload.SettlementAt

	} else if payload.AdditionalInfo.SettlementDetail != nil && payload.AdditionalInfo.SettlementDetail.EstimateSettlementAt != nil {
		result.SettlementAt = *payload.AdditionalInfo.SettlementDetail.EstimateSettlementAt

	} else if payload.AdditionalInfo.SettlementDetail != nil && strings.HasPrefix(payload.AdditionalInfo.SettlementDetail.Type, "T+") {
		days, _ := strconv.Atoi(payload.AdditionalInfo.SettlementDetail.Type[2:])
		result.SettlementAt = payload.UpdatedAt.AddDate(0, 0, days)

	} else {
		result.SettlementAt = payload.UpdatedAt
	}

	return result
}
