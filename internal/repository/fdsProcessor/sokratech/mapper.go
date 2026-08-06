package sokratech

import (
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/sokratech"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func toPayoutWorkflowRequest(data fdscommon.AssessPayoutTransactionRequest) model.PayoutWorkflowRequest {
	result := model.PayoutWorkflowRequest{
		Merchant: model.Merchant{
			ID:        data.Merchant.ID,
			Name:      data.Merchant.Name,
			RiskLevel: data.Merchant.RiskLevel,
		},
		Transaction: model.Transaction{
			ID:                data.Transaction.ID,
			ClientReferenceID: data.Transaction.ClientReferenceID,
			Amount: common.Amount2{
				Value:    data.Transaction.Amount.Value,
				Currency: data.Transaction.Amount.Currency,
			},
			CreatedAt:   data.Transaction.CreatedAt,
			UpdatedAt:   data.Transaction.UpdatedAt,
			CreatedFrom: data.Transaction.CreatedFrom,
		},
		Destination: model.PayoutDestination{
			BankCode:      data.Destination.BankCode,
			AccountNumber: data.Destination.AccountNumber,
			AccountName:   data.Destination.AccountName,
		},
		Metadata: data.Metadata,
	}
	result.Destination.AccountNumberTypeNumber, _ = strconv.ParseInt(result.Destination.AccountNumber, 10, 64)

	return result
}

func toPaymentWorkflowRequest(data *fdscommon.CheckRequest) model.PaymentWorkflowRequest {
	result := model.PaymentWorkflowRequest{
		Merchant: model.Merchant{
			ID:        data.Partner.ID,
			Name:      util.ValueOfPtr(data.Partner.Company),
			RiskLevel: data.Partner.RiskLevel,
		},
		Customer: model.Customer{
			ID:          data.Customer.ID,
			Name:        util.ValueOfPtr(data.Customer.FirstName),
			Email:       util.ValueOfPtr(data.Customer.Email),
			PhoneNumber: util.ValueOfPtr(data.Customer.Phone),
			Address:     util.ValueOfPtr(data.Customer.Address1),
		},
		Transaction: model.Transaction{
			ID:                data.Transaction.ID,
			ClientReferenceID: data.Transaction.ClientReferenceID,
			Type:              util.ValueOfPtr(data.Payment.Type),
			Amount: common.Amount2{
				Value:    data.Transaction.OrderTotal.InexactFloat64(),
				Currency: *data.Transaction.OrderCurrency,
			},
			CreatedAt: data.Transaction.CreatedAt,
			UpdatedAt: data.Transaction.UpdatedAt,
		},
		PaymentMethod: model.PaymentMethod{
			Type: constant.MapToUnifiedPaymentMethod(data.Payment.MethodType),
		},
		Device: model.Device{
			IPType:      data.Device.IPType,
			IPAddress:   util.ValueOfPtr(data.Device.IPAddress),
			Fingerprint: util.ValueOfPtr(data.Device.FingerprintID),
			UserAgent:   util.ValueOfPtr(data.Device.UserAgent),
		},
		Metadata: map[string]any{},
	}
	if result.PaymentMethod.Type == constant.UnifiedPaymentMethodCard {
		result.PaymentMethod.Card = toPaymentMethodTypeCard(data)
	}
	return result
}

func toPaymentMethodTypeCard(data *fdscommon.CheckRequest) *model.PaymentMethodTypeCard {
	result := &model.PaymentMethodTypeCard{
		ThreeDsMethod:   data.Payment.ThreeDsMethod,
		CardFingerprint: data.Payment.Fingerprint,
		CardNumber:      data.Payment.MaskedCardNumber,
		CardBrand:       data.Payment.CardBrand,
		CardCountryCode: data.Payment.CardCountryCode,
		CardType:        data.Payment.CardType,
		IssuerName:      data.Payment.CardIssuing,
		ECICode:         util.ValueOfPtr(data.Payment.ThreeDsEci),
		ApprovalCode:    util.ValueOfPtr(data.Payment.AuthCode),
		CvvCode:         util.ValueOfPtr(data.Payment.CvvResultCode),
	}
	result.Last4, _ = strconv.Atoi(data.Payment.Last4)
	result.First6, _ = strconv.Atoi(util.TrimLengthRight(data.Payment.First8, 6))
	if data.Custom != nil {
		result.BankMerchantID = util.ValueOfPtr(data.Custom.Number)
		result.AcquirerName = util.ValueOfPtr(data.Custom.AcquiringName)
	}
	return result
}
