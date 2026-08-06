package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	fdsCommonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	qrisSnapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) FindPaymentById(ctx context.Context, id, merchantID string) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/FindPaymentById")
	defer segment.End()

	var (
		paymentResponse        *paymentModel.PaymentResponse
		paymentItemsResponse   []paymentModel.PaymentResponseItem
		paymentVAResponse      *paymentModel.PaymentVirtualAccountResponse
		snapCoreResp           snapCoreModel.CreateVirtualAccountResponseData
		isSnap                 bool
		customer               *customerModel.Customer
		paymentRequestCustomer paymentModel.PaymentRequestCustomer
		paymentReferenceId     = ""
		minAmount              *paymentModel.Amount
		maxAmount              *paymentModel.Amount
		isUnifiedPayment       bool
		fdsRiskAssessment      *fdsCommonModel.FdsRiskAssessment
	)

	// Get Payment Method by category
	payment, err := s.paymentRepo.GetPaymentById(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by id", logger.Error(err))
		return nil, err
	}

	if payment == nil {
		err := pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found"))
		s.logger.Error(ctx, "payment not found", logger.String("id", id))
		return nil, err
	}

	// Get payment_method by id
	paymentMethod, err := s.paymentMethodRepo.GetPaymentMethodById(ctx, payment.PaymentMethodID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method data by id", logger.Error(err))
		return nil, err
	}

	// Get customer by id
	if payment.CustomerID != "" {
		customer, err = s.customerRepo.FindCustomerById(ctx, payment.CustomerID)
		if err != nil {
			s.logger.Error(ctx, "error when get customer data by id", logger.Error(err))
			return nil, err
		}

		// convert paymentRequestCustomer
		paymentRequestCustomer = paymentModel.PaymentRequestCustomer{
			CustomerID: customer.UUID,
			Name:       customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName),
			Email:      customer.Email,
			Phone:      customer.PhoneNumber,
			Metadata:   nil,
		}
	}

	// check if customer merchant id is same with request merchant id
	if payment.MerchantID != merchantID {
		s.logger.Error(ctx, "merchant id not match", logger.Error(fmt.Errorf("payment not found, merchant id not match")))
		return nil, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("payment not found"))
	}

	// Get account transaction to extract FDS risk assessment
	// This is optional data, so we don't fail the request if it's not available
	if s.accountTransactionRepo != nil {
		accountTransaction, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
		if err != nil {
			s.logger.Error(ctx, "error when get account transaction by reference", logger.Error(err))
			// Don't return error here, just log it and continue without FDS data
		} else if accountTransaction != nil && accountTransaction.AdditionalInfo.Valid {
			// Extract FDS risk assessment from additional_info
			var additionalInfo map[string]interface{}
			if err := json.Unmarshal(accountTransaction.AdditionalInfo.JSONText, &additionalInfo); err == nil {
				if fdsData, exists := additionalInfo["fdsRiskAssessment"]; exists {
					// Convert the fdsData to JSON and then unmarshal to FdsRiskAssessment
					if fdsBytes, err := json.Marshal(fdsData); err == nil {
						var fdsAssessment fdsCommonModel.FdsRiskAssessment
						if err := json.Unmarshal(fdsBytes, &fdsAssessment); err == nil {
							fdsRiskAssessment = &fdsAssessment
						}
					}
				}
			}
		}
	}

	// Get payment items
	paymentItems, err := s.paymentRepo.GetPaymentItemsByPaymentId(ctx, payment.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment items data by payment id", logger.Error(err))
		return nil, err
	}

	for _, item := range paymentItems {
		paymentItemsResponse = append(paymentItemsResponse, *item.ToPaymentResponseItem())
	}

	// create amount
	amount := paymentModel.Amount{
		Value:    payment.Amount,
		Currency: payment.Currency,
	}

	// unmarshal payment.Metadata map[string]any to createVirtualAccountResponseData
	if payment.Metadata != nil {
		jsonData, errMarshal := json.Marshal(payment.Metadata)
		if errMarshal != nil {
			s.logger.Error(ctx, "error when marshal payment metadata", logger.Error(errMarshal))
			return nil, errMarshal
		}

		json.Unmarshal(jsonData, &struct {
			SnapCore         interface{} `json:"snapCore"`
			IsSnap           *bool       `json:"isSnap"`
			IsUnifiedPayment *bool       `json:"isUnifiedPayment"`
		}{
			SnapCore:         &snapCoreResp,
			IsSnap:           &isSnap,
			IsUnifiedPayment: &isUnifiedPayment,
		})

		if snapCoreResp.MinAmount.Value != "" {
			minAmount = &paymentModel.Amount{
				Value:    decimal.RequireFromString(snapCoreResp.MinAmount.Value),
				Currency: snapCoreResp.MinAmount.Currency,
			}
		}

		if snapCoreResp.MaxAmount.Value != "" {
			maxAmount = &paymentModel.Amount{
				Value:    decimal.RequireFromString(snapCoreResp.MaxAmount.Value),
				Currency: snapCoreResp.MaxAmount.Currency,
			}
		}

		paymentVAResponse = &paymentModel.PaymentVirtualAccountResponse{
			Issuer:                snapCoreResp.Acquirer,
			VirtualAccountTrxType: snapCoreModel.FindVaTrxTypeByCriteria(snapCoreResp.IsClosedAmount, snapCoreResp.IsSingleUse),
			VirtualAccountNumber:  snapCoreResp.VirtualAccountNo,
			VirtualAccountName:    snapCoreResp.AccountName,
			MinAmount:             minAmount,
			MaxAmount:             maxAmount,
			ExpiredDate:           &snapCoreResp.ExpiredAt,
			IsSnap:                isSnap,
			BillDetails:           snapCoreResp.BillDetails,
		}
	}

	if payment.ReferenceID != nil {
		paymentReferenceId = *payment.ReferenceID
	}

	paymentResponse = &paymentModel.PaymentResponse{
		UUID:              payment.UUID,
		MerchantID:        payment.MerchantID,
		ReferenceID:       paymentReferenceId,
		Customer:          &paymentRequestCustomer,
		Status:            payment.Status,
		TotalAmount:       &amount,
		PaymentMethod:     paymentMethod.Type,
		VirtualAccount:    paymentVAResponse,
		PaymentItems:      &paymentItemsResponse,
		LastUpdateDate:    &payment.UpdatedAt,
		CreatedAt:         payment.CreatedAt,
		ExpiredAt:         payment.ExpiredAt,
		PaymentURL:        payment.PaymentURL,
		IsUnifiedPayment:  isUnifiedPayment,
		FdsRiskAssessment: fdsRiskAssessment,
	}

	return paymentResponse, nil
}

// GetTotalPaymentBalance return common model of amount that contain merchant payment balance
// it will calculate using orchestrator service to recalculate the balance
func (s *PaymentService) GetTotalPaymentBalance(ctx context.Context, merchantID uuid.UUID) (*commonModel.Amount, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetTotalPaymentBalance")
	defer segment.End()

	account, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantID, constant.TypePayment)
	if err != nil {
		return &commonModel.Amount{}, err
	}

	if account == nil {
		return &commonModel.Amount{
			Value: strconv.FormatFloat(0, 'f', 2, 64),
		}, nil
	}

	balance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, merchantID.String(), account.Name)
	if err != nil {
		return &commonModel.Amount{}, err
	}

	return &commonModel.Amount{
		Value:    strconv.FormatFloat(balance, 'f', 2, 64),
		Currency: account.Currency,
	}, nil
}

// GetPaymentHistoryDetail return detail information for dashboard history
func (s *PaymentService) GetPaymentHistoryDetail(ctx context.Context, opt paymentModel.PaymentHistoryDetailOption) (*paymentModel.PaymentHistoryDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentHistoryDetail")
	defer segment.End()

	var (
		paymentDetail       paymentModel.PaymentHistoryDetailResponse
		processorRefNumber  string
		referenceNumber     string
		paymentChanel       string
		creditAmount        float64
		creditCurrency      = constant.CurrencyIDR
		splitRoutingConfigs []paymentModel.SplitRoutingConfiguration
		settlementModel     = constant.PaymentMethodChannelTypeAggregator
		clientMetadata      *map[string]interface{}
	)

	payment, err := s.paymentRepo.GetPaymentByIdAndMerchantId(ctx, opt.PaymentID, opt.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by id", logger.Error(err))
		return &paymentDetail, err
	}

	if payment == nil {
		return &paymentDetail, pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found"))
	}

	charges, err := s.paymentRepo.GetChargeList(ctx, &unifiedPaymentModel.FilterChargeRequest{
		MerchantID:       opt.MerchantID,
		PaymentSessionID: payment.UUID,
		Page:             1,
		PerPage:          1000,
	})
	if err != nil {
		return &paymentDetail, err
	}

	unifiedRefund := payment.ToUnifiedPaymentResponse()
	payment.AutoSplitPayment = unifiedRefund.AutoSplitPayment

	if listChargeData, ok := charges.Data.([]*unifiedPaymentModel.ChargeResponse); ok {
		unifiedRefund.ChargeDetails = listChargeData
	}

	ledgerMetadata := []byte{}
	var fdsRiskAssessment *fdsCommonModel.FdsRiskAssessment

	if len(unifiedRefund.ChargeDetails) > 0 {
		for _, charge := range unifiedRefund.ChargeDetails {
			if charge.Status == constant.StatusSuccess {
				creditAmount += charge.Amount.Value
				creditCurrency = charge.Amount.Currency
			}
		}
	}

	accountTransaction, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err == nil && accountTransaction != nil {
		if accountTransaction.AdditionalInfo.Valid {
			ledgerMetadata = accountTransaction.AdditionalInfo.JSONText

			// Extract FDS risk assessment from additional_info
			var additionalInfo map[string]interface{}
			if err := json.Unmarshal(ledgerMetadata, &additionalInfo); err == nil {
				if fdsData, exists := additionalInfo["fdsRiskAssessment"]; exists {
					// Convert the fdsData to JSON and then unmarshal to FdsRiskAssessment
					if fdsBytes, err := json.Marshal(fdsData); err == nil {
						var fdsAssessment fdsCommonModel.FdsRiskAssessment
						if err := json.Unmarshal(fdsBytes, &fdsAssessment); err == nil {
							fdsRiskAssessment = &fdsAssessment
						}
					}
				}
			}
		}
	}

	if accountTransaction != nil && accountTransaction.SettlementModel.Valid &&
		constant.IsDirectPSP(accountTransaction.SettlementModel.String) {
		settlementModel = accountTransaction.SettlementModel.String
	}

	customer, err := s.customerRepo.GetCustomerById(ctx, payment.CustomerID, payment.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when get customer data by id", logger.Error(err))
		return nil, err
	}

	paymentTypeDetail := s.GetPaymentTypeDetail(ctx, payment.PaymentMethod.Type, payment.Metadata, ledgerMetadata)

	paymentAmount, _ := payment.Amount.Float64()

	if payment.Metadata != nil {
		var totalSplitAmount float64
		var fee float64

		// Get fee from feeDetail
		if feeDetail, ok := (*payment.Metadata)["feeDetail"].(map[string]interface{}); ok {
			if finalAmount, ok := feeDetail["finalAmount"].(float64); ok {
				fee = finalAmount
			}
		}

		clientMetadata, _ = util.ConvertToStruct[*map[string]interface{}]((*payment.Metadata)["clientMetadata"])

		if configs, ok := (*payment.Metadata)[constant.SplitRoutingPaymentConfigKey].([]interface{}); ok {
			configsB, err := json.Marshal(configs)
			if err != nil {
				s.logger.Error(ctx, "failed to marshal split routing configuration",
					logger.String("paymentId", payment.UUID),
					logger.Error(err))
				return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed)
			}

			if err := json.Unmarshal(configsB, &splitRoutingConfigs); err != nil {
				s.logger.Error(ctx, "failed to unmarshal split routing configuration",
					logger.String("paymentId", payment.UUID),
					logger.Error(err))
				return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed)
			}

			for i, config := range splitRoutingConfigs {
				if config.TransferID == "" {
					continue
				}
				// Calculate split amount based on type
				finalAmount := 0.0
				if config.Type == "FIXED" {
					finalAmount = config.FixedAmount
					totalSplitAmount += finalAmount
				} else if config.Type == "PERCENTAGE" {
					finalAmount = (payment.Amount.InexactFloat64() * config.PercentageAmount) / 100
					totalSplitAmount += finalAmount
				}

				transfer, err := s.transferSvc.GetById(ctx, config.TransferID, payment.MerchantID)
				if err != nil {
					return nil, pkgErrors.New(response.HttpErrDatabase, err)
				} else if transfer == nil {
					s.logger.Info(ctx, "transfer to destination is not found", logger.String("paymentId", payment.UUID), logger.String("transferId", config.TransferID))
					return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed)
				}
				config.Status = transfer.Status
				config.Beneficiary = transfer.Beneficiary
				config.FinalAmount = finalAmount
				splitRoutingConfigs[i] = config
			}
		}

		paymentDetail = paymentModel.PaymentHistoryDetailResponse{
			UUID:                  payment.UUID,
			MerchantID:            payment.MerchantID,
			CustomerID:            payment.CustomerID,
			ReferenceID:           "",
			RecurringID:           util.ValueOfPtr(payment.RecurringContractID),
			PaymentMethod:         payment.PaymentMethod.Type,
			PaymentMethodCategory: s.GetPaymentSubType(ctx, payment.PaymentMethod.Type, payment.Metadata),
			TypeDetail:            paymentTypeDetail,
			Amount: commonModel.Amount{
				Currency: payment.Currency,
				Value:    strconv.FormatFloat(paymentAmount, 'f', 2, 64),
			},
			AmountPaid: commonModel.Amount{
				Currency: creditCurrency,
				Value:    strconv.FormatFloat(creditAmount, 'f', 2, 64),
			},
			BankReferenceID:        "-",
			Channel:                "",
			ProcessorRefNumber:     "",
			Status:                 payment.Status,
			CreatedAt:              payment.CreatedAt,
			UpdatedAt:              payment.UpdatedAt,
			ExpiredAt:              payment.ExpiredAt,
			CancelledAt:            unifiedRefund.CancelledAt,
			InvestigationStartedAt: payment.InvestigationStartedAt,
			CancellationReason:     unifiedRefund.CancellationReason,
			SplitRoutingConfigs:    splitRoutingConfigs,
			FdsRiskAssessment:      fdsRiskAssessment,
			Charges:                unifiedRefund.ChargeDetails,
			Metadata:               clientMetadata,
		}

		if payment.CreatedFrom == constant.DisbursementCreatedFromMerchantPortal {
			shortPaymentUrl, exist := (*payment.Metadata)["shortPaymentUrl"]
			if !exist {
				s.logger.Warn(ctx, "shortPaymentUrl not found in metadata", logger.String("paymentId", payment.UUID))
				paymentDetail.PaymentLink = payment.PaymentURL
			} else {
				paymentDetail.PaymentLink = shortPaymentUrl.(string)
			}
		}

		// Set the amount fields after initializing the main struct
		paymentDetail.TotalSplitAmount = commonModel.Amount{
			Currency: payment.Currency,
			Value:    decimal.NewFromFloat(totalSplitAmount).StringFixed(2),
		}
		paymentDetail.Fee = commonModel.Amount{
			Currency: payment.Currency,
			Value:    decimal.NewFromFloat(fee).StringFixed(2),
		}
		paymentDetail.SettledAmount = commonModel.Amount{
			Currency: payment.Currency,
			Value:    decimal.NewFromFloat(paymentAmount - totalSplitAmount).StringFixed(2),
		}

		paymentDetail.LoadPaymentV2CustomerOrderInformation(payment, customer)

		refundList, err := s.refundSvc.GetExistingRefundList(ctx, refundModel.GetExistingRefundListRequest{
			PaymentID: payment.UUID,
		})
		if err != nil {
			s.logger.Warn(ctx, "failed to get refund list", logger.String("paymentId", payment.UUID), logger.Error(err))
			return nil, err
		}
		if len(refundList) > 0 {
			paymentDetail.RefundDetails = refundList
		}
	}

	if payment.ProcessorReferenceNumber != nil {
		processorRefNumber = *payment.ProcessorReferenceNumber
	}

	if payment.ReferenceID != nil {
		referenceNumber = *payment.ReferenceID
	}

	if payment.PaymentMethod.BankName != nil {
		paymentChanel = *payment.PaymentMethod.BankName
	}

	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		paymentChanel = paymentTypeDetail.CardBrand
	}

	if payment.IsAutoSplitPaymentAuth() && payment.GetAutoSplitTotalSuccessAmount() != nil {
		paymentDetail.AmountPaid = *payment.GetAutoSplitTotalSuccessAmount()
	}

	paymentDetail.Channel = paymentChanel
	paymentDetail.ProcessorRefNumber = processorRefNumber
	paymentDetail.ReferenceID = referenceNumber
	paymentDetail.SettlementModel = settlementModel
	paymentDetail.PaymentType = payment.Type

	// Fetch and populate status history
	statusHistoryList, err := s.statusHistoriesRepo.GetByReference(ctx, constant.TypePayment, payment.UUID)
	if err != nil {
		s.logger.Warn(ctx, "failed to get status history", logger.String("paymentId", payment.UUID), logger.Error(err))
	} else if len(statusHistoryList) > 0 {
		paymentDetail.StatusHistory = make([]unifiedPaymentModel.PaymentStatusHistoryResponse, 0)
		seenLabel := make(map[string]bool)
		for _, statusHistory := range statusHistoryList {
			label := s.getPaymentStatusLabel(statusHistory.Status)
			description := s.getPaymentStatusDescription(statusHistory.Status)
			recommendation := s.getPaymentStatusRecommendation(statusHistory.Status)
			if _, ok := seenLabel[label]; ok {
				continue
			}
			paymentDetail.StatusHistory = append(paymentDetail.StatusHistory, unifiedPaymentModel.PaymentStatusHistoryResponse{
				Status:         statusHistory.Status,
				Label:          label,
				Description:    description,
				Recommendation: recommendation,
				Timestamp:      &statusHistory.CreatedAt,
			})
			seenLabel[label] = true
		}
	}

	return &paymentDetail, nil
}

// GetPaymentMethodSubType return sub type of the payment method
// it will check the payment method, then read the metadata to decide the category of the payment method.
func (s *PaymentService) GetPaymentSubType(ctx context.Context, method string, metadata *map[string]any) string {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentMethodSubType")
	defer segment.End()

	if metadata == nil {
		return ""
	}

	if method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
		vaMetadata, err := buildVAMetadataSimulation(metadata)
		if err != nil {
			s.logger.Error(ctx, "invalid va metadata", logger.String("method", method), logger.Any("metadata", metadata))
			return ""
		}

		return snapCoreModel.FindVaTrxTypeByCriteria(vaMetadata.IsClosedAmount, vaMetadata.IsSingleUse)
	}

	if method == paymentConstant.PAYMENT_METHOD_QRIS {
		// error should not ocurred after preventing pass the nil metadata
		qrisMetadata, _ := buildQrisMetadata(metadata)
		return qrisMetadata.QrType
	}

	if method == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		var (
			ccMetadata creditcardModel.CreditcardMetadata
		)

		// the metadata should always valid due internal service who insert it
		bytes, _ := json.Marshal(metadata)
		_ = json.Unmarshal(bytes, &ccMetadata)

		if ccMetadata.CardData == nil {
			s.logger.Error(ctx, "missing card data", logger.Any("metadata", ccMetadata))
			return ""
		}

		return ccMetadata.CardData.CardType
	}

	return ""
}

// GetPaymentTypeDetail return Detail of the payment method, which target that doing the payment
// each payment method have different detail
// check this url https://www.figma.com/design/E3lHK0VCVyaqXdc66bBO5I/PG---Payment-Dashboard?node-id=0-1&node-type=canvas&t=uilXyBvJ8jft7Y6N-0
func (s *PaymentService) GetPaymentTypeDetail(ctx context.Context, method string, metadata *map[string]any, ledgerMetadata []byte) paymentModel.PaymentTypeDetail {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentTypeDetail")
	defer segment.End()

	var (
		result       paymentModel.PaymentTypeDetail
		defaultValue string
	)

	if metadata == nil {
		return result
	}

	paymentDataB, _ := json.Marshal(metadata)
	unifiedPaymentMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: false,
	}
	_ = json.Unmarshal(paymentDataB, &unifiedPaymentMetadata)

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	if unifiedPaymentMetadata.IsUnifiedPaymentV2 && len(ledgerMetadata) > 0 {
		err := json.Unmarshal(ledgerMetadata, &struct {
			MethodDetail interface{} `json:"methodDetail"`
		}{
			MethodDetail: chargeMethodDetails,
		})
		if err != nil {
			s.logger.Warn(ctx, "failed to unmarshal ledger metadata",
				logger.String("method", method),
				logger.Error(err),
				logger.String("ledgerMetadata", string(ledgerMetadata)))
		}
	}

	if method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
		if unifiedPaymentMetadata.IsUnifiedPaymentV2 {
			vaName := defaultValue
			vaNumber := defaultValue

			if chargeMethodDetails.VirtualAccount != nil {
				vaName = chargeMethodDetails.VirtualAccount.VirtualAccountName
				vaNumber = chargeMethodDetails.VirtualAccount.VirtualAccountNumber
			}

			return paymentModel.PaymentTypeDetail{
				VirtualAccountName:   &vaName,
				VirtualAccountNumber: &vaNumber,
			}
		}

		vaMetadata, err := buildVAMetadataSimulation(metadata)
		if err != nil {
			s.logger.Error(ctx, "invalid va metadata", logger.String("method", method), logger.Any("metadata", metadata))
			return paymentModel.PaymentTypeDetail{
				VirtualAccountName:   &defaultValue,
				VirtualAccountNumber: &defaultValue,
			}
		}

		return paymentModel.PaymentTypeDetail{
			VirtualAccountName:   &vaMetadata.AccountName,
			VirtualAccountNumber: &vaMetadata.Number,
		}
	}

	if method == paymentConstant.PAYMENT_METHOD_QRIS {
		if unifiedPaymentMetadata.IsUnifiedPaymentV2 {
			qrContent := defaultValue
			merchantName := defaultValue
			qrUrl := defaultValue

			if chargeMethodDetails.Qr != nil {
				if chargeMethodDetails.Qr.QrContent != "" {
					qrContent = chargeMethodDetails.Qr.QrContent
				}
				if chargeMethodDetails.Qr.MerchantName != "" {
					merchantName = chargeMethodDetails.Qr.MerchantName
				}
				if chargeMethodDetails.Qr.QrUrl != "" {
					qrUrl = chargeMethodDetails.Qr.QrUrl
				}
			}

			return paymentModel.PaymentTypeDetail{
				QrContent:        &qrContent,
				QRISMerchantName: &merchantName,
				QRISURL:          &qrUrl,
			}
		}

		var qrisMetadata paymentModel.SnapQrAdditionalInfo
		snapCore, ok := (*metadata)["snapCore"]
		if !ok {
			s.logger.Error(ctx, "QRIS snapCore data not found", logger.String("method", method), logger.Any("metadata", metadata))
			return paymentModel.PaymentTypeDetail{
				QrContent:        &defaultValue,
				QRISMerchantName: &defaultValue,
				QRISURL:          &defaultValue,
			}
		}

		// the metadata should always valid due internal service who insert it
		bytes, _ := json.Marshal(snapCore)
		_ = json.Unmarshal(bytes, &qrisMetadata)

		return paymentModel.PaymentTypeDetail{
			QrContent:        &qrisMetadata.QrContent,
			QRISMerchantName: &qrisMetadata.MerchantName,
			QRISURL:          &qrisMetadata.QrUrl,
		}
	}

	if method == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		if unifiedPaymentMetadata.IsUnifiedPaymentV2 {
			if chargeMethodDetails.Card == nil {
				return paymentModel.PaymentTypeDetail{
					CardIssuer: &defaultValue,
					CardNumber: &defaultValue,
					CardBrand:  defaultValue,
				}
			}

			paymentTypeDetail := paymentModel.PaymentTypeDetail{}
			if chargeMethodDetails.Card.BinInformations.IssuingBank != "" {
				paymentTypeDetail.CardIssuer = &chargeMethodDetails.Card.BinInformations.IssuingBank
			}
			if chargeMethodDetails.Card.Last4 != "" {
				paymentTypeDetail.CardNumber = &chargeMethodDetails.Card.Last4
			}
			if chargeMethodDetails.Card.BinInformations.Brand != "" {
				paymentTypeDetail.CardBrand = chargeMethodDetails.Card.BinInformations.Brand
			}
			if chargeMethodDetails.Card.ExpYear != "" {
				paymentTypeDetail.CardExpiryYear = chargeMethodDetails.Card.ExpYear.String()
			}
			if chargeMethodDetails.Card.ExpMonth != "" {
				paymentTypeDetail.CardExpiryMonth = chargeMethodDetails.Card.ExpMonth.String()
			}
			if chargeMethodDetails.Card.CardName != "" {
				paymentTypeDetail.CardName = chargeMethodDetails.Card.CardName
			}
			return paymentTypeDetail
		}

		var (
			ccMetadata creditcardModel.CreditcardMetadata
		)
		// the metadata should always valid due internal service who insert it
		bytes, _ := json.Marshal(metadata)
		_ = json.Unmarshal(bytes, &ccMetadata)

		if ccMetadata.CardData == nil {
			s.logger.Error(ctx, "missing card data", logger.Any("metadata", ccMetadata))
			return paymentModel.PaymentTypeDetail{
				CardIssuer: &defaultValue,
				CardNumber: &defaultValue,
				CardBrand:  defaultValue,
			}
		}

		cardBrand := defaultValue
		if ccMetadata.CardData.CardBrand != "" {
			cardBrand = ccMetadata.CardData.CardBrand
		}

		return paymentModel.PaymentTypeDetail{
			CardIssuer:      &ccMetadata.CardData.CardIssuing,
			CardNumber:      &ccMetadata.CardData.Last4Digit,
			CardBrand:       cardBrand,
			CardName:        ccMetadata.CardData.CardName,
			CardExpiryMonth: ccMetadata.CardData.ExpiryMonth,
			CardExpiryYear:  ccMetadata.CardData.ExpiryYear,
		}
	}

	if method == paymentConstant.PAYMENT_METHOD_EWALLET {
		if unifiedPaymentMetadata.IsUnifiedPaymentV2 {
			if chargeMethodDetails.Ewallet == nil {
				s.logger.Warn(ctx, "ewallet charge details is nil for unified payment v2")
			} else {
				return paymentModel.PaymentTypeDetail{
					EWalletAppRedirectURL: chargeMethodDetails.Ewallet.AppRedirectURL,
					EWalletWebRedirectURL: chargeMethodDetails.Ewallet.WebRedirectURL,
					EWalletChannel:        chargeMethodDetails.Ewallet.Channel,
				}
			}
		}
	}
	return result
}

// GetPaymentInsight return total payment and the amount based on the payment status
// it will calculate the data only for today
// when it found error, then it will return the error with nil payment item
func (s *PaymentService) GetTodayPaymentInsight(ctx context.Context, opt paymentModel.PaymentInsightOption) (*paymentModel.PaymentInsightItem, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentInsight")
	defer segment.End()

	var (
		eg             = new(errgroup.Group)
		defaultInsight = paymentModel.PaymentInsightItem{
			Total: 0,
			TotalAmount: commonModel.Amount{
				Value: strconv.FormatFloat(0, 'f', 2, 64),
			},
		}
		account *accountModel.Account
		insight *paymentModel.PaymentInsightItem
	)

	merchantID, err := uuid.Parse(opt.MerchantID)
	if err != nil {
		return nil, err
	}

	eg.Go(func() error {
		var err error
		// currently we are merging the account, so we should use disbursement account type
		account, err = s.accountRepo.FindMerchantAccountByName(ctx, merchantID, constant.TypePayment)
		if err != nil {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		var err error
		insight, err = s.paymentRepo.GetTodayPaymentStatusInsight(ctx, opt)
		if err != nil {
			return err
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	if account == nil || insight == nil {
		return &defaultInsight, nil
	}

	insight.TotalAmount.Currency = account.Currency
	return insight, nil
}

func (s *PaymentService) GetPaymentDetailForPaymentUI(ctx context.Context, paymentID string) (*paymentModel.PaymentDetailForPaymentUIResponse, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentDetailForPaymentUI")
	defer segment.End()

	var (
		processorRefNumber string
		referenceNumber    string
		paymentChanel      string
		paidAmount         = 0.0
		paidCurrency       = constant.CurrencyIDR
		transactionID      string
		paidAt             *time.Time // Using transaction createdAt for temp
		chargeMethodDetail = &unifiedPaymentModel.ChargePaymentMethodDetails{}
		chargeStatus       string
		failureCode        string
	)

	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Info(ctx, constant.ErrPaymentNotFound.Error(), logger.String("paymentId", paymentID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	defer func() {
		if payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_EWALLET {
			return
		}

		_, err = s.unifiedPaymentSvc.UpdateEWalletPaymentSession(ctx, payment.UUID)
		if err != nil {
			s.logger.Warn(ctx, "error when update ewallet payment sessions", logger.String("paymentId", paymentID), logger.Error(err))
		}
	}()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, payment.MerchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		s.logger.Info(ctx, constant.ErrMerchantNotFound.Error(), logger.String("paymentId", paymentID), logger.String("merchantId", payment.MerchantID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	accountTransaction, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	ledgerMetadata := []byte{}
	var fdsRiskAssessment *fdsCommonModel.FdsRiskAssessment
	if accountTransaction != nil {
		_ = json.Unmarshal(accountTransaction.AdditionalInfo.JSONText, &struct {
			ChargeStatus *string     `json:"chargeStatus"`
			FailureCode  *string     `json:"failureCode"`
			MethodDetail interface{} `json:"methodDetail"`
		}{
			ChargeStatus: &chargeStatus,
			MethodDetail: chargeMethodDetail,
			FailureCode:  &failureCode,
		})

		if accountTransaction.Status == constant.StatusSuccess || chargeStatus == constant.ChargeStatusHistoryWaitingForCapture {
			paidAmount = accountTransaction.Credit
			paidCurrency = accountTransaction.Currency
			transactionID = accountTransaction.UUID.String()
			paidAt = &accountTransaction.TransactionTimestamp
		}

		if accountTransaction.AdditionalInfo.Valid {
			ledgerMetadata = accountTransaction.AdditionalInfo.JSONText

			// Extract FDS risk assessment from additional_info
			var additionalInfo map[string]interface{}
			if err := json.Unmarshal(ledgerMetadata, &additionalInfo); err == nil {
				if fdsData, exists := additionalInfo["fdsRiskAssessment"]; exists {
					// Convert the fdsData to JSON and then unmarshal to FdsRiskAssessment
					if fdsBytes, err := json.Marshal(fdsData); err == nil {
						var fdsAssessment fdsCommonModel.FdsRiskAssessment
						if err := json.Unmarshal(fdsBytes, &fdsAssessment); err == nil {
							fdsRiskAssessment = &fdsAssessment
						}
					}
				}
			}
		}
	}

	paymentTypeDetail := s.GetPaymentTypeDetail(ctx, payment.PaymentMethod.Type, payment.Metadata, ledgerMetadata)
	paymentAmount, _ := payment.Amount.Float64()

	if payment.ProcessorReferenceNumber != nil {
		processorRefNumber = *payment.ProcessorReferenceNumber
	}

	if payment.ReferenceID != nil {
		referenceNumber = *payment.ReferenceID
	}

	if payment.PaymentMethod.BankName != nil {
		paymentChanel = *payment.PaymentMethod.BankName
	}

	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		paymentChanel = paymentTypeDetail.CardBrand
	}

	paymentMetadata := payment.ToUnifiedPaymentMetadata()
	payment.AutoSplitPayment = paymentMetadata.AutoSplitPayment

	responseData := &paymentModel.PaymentDetailForPaymentUIResponse{
		UUID:        payment.UUID,
		MerchantID:  payment.MerchantID,
		CustomerID:  payment.CustomerID,
		ReferenceID: referenceNumber,
		Merchant: paymentModel.MerchantDetail{
			Name: merchant.ShortName,
			Logo: merchant.Logo,
		},
		PaymentMethod: paymentModel.PaymentMethodDetail{
			Name:     payment.PaymentMethod.Name,
			Method:   payment.PaymentMethod.Type,
			Logo:     payment.PaymentMethod.Logo,
			Category: s.GetPaymentSubType(ctx, payment.PaymentMethod.Type, payment.Metadata),
		},
		TypeDetail: paymentTypeDetail,
		Amount: commonModel.Amount{
			Currency: payment.Currency,
			Value:    strconv.FormatFloat(paymentAmount, 'f', 2, 64),
		},
		AmountPaid: commonModel.Amount{
			Currency: paidCurrency,
			Value:    strconv.FormatFloat(paidAmount, 'f', 2, 64),
		},
		InfoBanner: &paymentModel.InfoBanner{
			Message: chargeMethodDetail.GetNaturalPaymentFailureMessage(payment.PaymentMethod.Type, failureCode),
		},
		ChargeStatus:       chargeStatus,
		BankReferenceID:    "-",
		Channel:            paymentChanel,
		ProcessorRefNumber: processorRefNumber,
		Status:             payment.Status,
		TransactionID:      transactionID,
		CreatedAt:          payment.CreatedAt,
		UpdatedAt:          payment.UpdatedAt,
		ExpiredAt:          payment.ExpiredAt,
		ExpirationMode:     paymentMetadata.ExpirationMode,
		PaidAt:             paidAt,
		FdsRiskAssessment:  fdsRiskAssessment,
		RedirectUrl:        s.GetRedirectUrlFromMetadata(payment.Metadata),
		CreatedFrom:        payment.CreatedFrom,
		Mode:               paymentMetadata.Mode,
		BypassStatusPage:   paymentMetadata.GetBypassStatusPageState(),
		PaymentType:        payment.Type,
	}

	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_EWALLET && payment.Status == constant.UnifiedPaymentSessionStatusProcessing {
		inquiryPayment, err := s.unifiedPaymentSvc.InquiryEWalletPayment(ctx, payment)
		if err != nil {
			s.logger.Warn(ctx, "error when inquiry ewallet payment", logger.String("paymentId", paymentID), logger.Error(err))
		}

		if inquiryPayment != nil && inquiryPayment.InquiryDetail != nil {
			responseData.InquiryDetail = &paymentModel.InquiryDetailResponse{
				HasFinalStatus: inquiryPayment.InquiryDetail.HasFinalStatus(),
			}
		}
	}

	if s.shouldShowAuthCaptureBanner(ctx, payment, accountTransaction) {
		responseData.InfoBanner = s.getAuthCaptureBannerFromConsul(ctx)
	}

	if accountTransaction != nil && accountTransaction.Status == constant.StatusSuccess &&
		payment.IsAutoSplitPaymentAuth() &&
		payment.GetAutoSplitTotalSuccessAmount() != nil {
		responseData.AmountPaid = *payment.GetAutoSplitTotalSuccessAmount()
	}

	// set derived merchant id when its not approved and merchant has parent merchant
	if merchant.ParentID.String != "" && merchant.KYCStatus.String != constant.KYCStatusApproved {
		responseData.DerivedMerchantID = merchant.ParentID.String
	}

	switch payment.Type {
	case constant.UnifiedPaymentOneDollarAuthorization:
		if paymentMetadata != nil {
			responseData.Metadata = &paymentModel.PaymentDetailForPaymentUIMetadata{
				UseCase: paymentMetadata.OneDollarAuthorization.UseCase,
			}
		}
	case constant.PaymentTypeCardFundedPayout:
		if cardFundPayout := s.getCardFundedPayoutMetadata(ctx, payment); cardFundPayout != nil {
			responseData.Metadata = &paymentModel.PaymentDetailForPaymentUIMetadata{
				CardFundedPayout: cardFundPayout,
			}
		}
	}

	return responseData, nil
}

// GetRedirectUrlFromMetadata extracts and returns the client redirect URLs from the payment metadata.
// It checks if the metadata contains the "isUnifiedPaymentV2" flag to determine which format the redirect URLs are in.
// If it's using unified payment v2, it extracts the redirect URLs from the v2 format.
// Otherwise, it extracts the redirect URLs from the standard format.
// If metadata is nil, it returns an empty UnifiedPaymentRedirectUrl structure.
func (s *PaymentService) GetRedirectUrlFromMetadata(metadata *map[string]any) (clientRedirectUrl paymentModel.UnifiedPaymentRedirectUrl) {
	if metadata == nil {
		return
	}

	if _, ok := (*metadata)["isUnifiedPaymentV2"]; ok {
		unifiedV2Redirection, _ := util.ConvertToStruct[unifiedPaymentModel.RedirectUrl]((*metadata)["clientRedirectUrl"])
		clientRedirectUrl.SuccessUrl = unifiedV2Redirection.SuccessReturnUrl
		clientRedirectUrl.ExpiredUrl = unifiedV2Redirection.ExpirationReturnUrl
		clientRedirectUrl.FailedUrl = unifiedV2Redirection.FailureReturnUrl
		return
	}

	clientRedirectUrl, _ = util.ConvertToStruct[paymentModel.UnifiedPaymentRedirectUrl]((*metadata)["clientRedirectUrl"])
	return clientRedirectUrl
}

func (s *PaymentService) buildPaymentForQris(ctx context.Context, payment *paymentModel.Payment) *paymentModel.PaymentResponse {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/findPaymentForQris")
	defer segment.End()

	// Parse metadata
	metadata := map[string]any{}
	if payment.Metadata != nil {
		metadata = *payment.Metadata
	}

	// Parse snap core response
	var snapCoreResp qrisSnapCoreModel.GenerateQrMpmResponseData
	snapCoreB, _ := json.Marshal(metadata["snapCore"])
	json.Unmarshal(snapCoreB, &snapCoreResp)

	// Parse payment request from metadata
	paymentRequest := &paymentModel.PaymentRequest{
		PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
		Qris:          &paymentModel.PaymentMetadataQris{},
	}
	paymentRequestB, _ := json.Marshal(metadata)
	json.Unmarshal(paymentRequestB, &paymentRequest.Qris)

	// Create paymentDTO
	paymentDTO := payment.ToDTO()

	var paymentResponse paymentModel.PaymentResponse
	paymentResponse.ToQrisResponse(
		paymentDTO,
		&snapCoreResp,
		paymentRequest,
	)

	if isUnifiedPayment, ok := metadata[constant.IsUnifiedPaymentKey].(bool); ok {
		paymentResponse.IsUnifiedPayment = isUnifiedPayment
	}

	return &paymentResponse
}

func (s *PaymentService) GetSplitRoutingByTransferID(ctx context.Context, paymentID, transferID string) (*splitRoutingPaymentModel.SplitRoutingPaymentDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetSplitRoutingByTransferID")
	defer segment.End()

	// Find Payment
	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Info(ctx, constant.ErrPaymentNotFound.Error(), logger.String("paymentId", paymentID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	if payment.Metadata == nil {
		s.logger.Info(ctx, "payment metadata is empty", logger.String("paymentId", paymentID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentDoesNotHaveSplitRouting)
	}

	paymentMetadata := *payment.Metadata

	// get split routing payment config from metadata
	var (
		splitRoutingConfigs []*splitRoutingPaymentModel.PaymentSplitRoutingConfiguration
		splitRoutingConfig  *splitRoutingPaymentModel.PaymentSplitRoutingConfiguration
	)
	splitRoutingConfigB, _ := json.Marshal(paymentMetadata[constant.SplitRoutingPaymentConfigKey])
	json.Unmarshal(splitRoutingConfigB, &splitRoutingConfigs)

	if len(splitRoutingConfigs) == 0 {
		s.logger.Info(ctx, "split routing config is empty", logger.String("paymentId", paymentID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentDoesNotHaveSplitRouting)
	}

	configIndex := slices.IndexFunc(splitRoutingConfigs, func(c *splitRoutingPaymentModel.PaymentSplitRoutingConfiguration) bool {
		return c.TransferID == transferID
	})

	if configIndex == -1 {
		s.logger.Info(ctx, "split routing by transfer id is not found", logger.String("paymentId", paymentID), logger.String("transferId", transferID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingDestinationNotFound)
	}
	splitRoutingConfig = splitRoutingConfigs[configIndex]

	// Get transfer data
	transfer, err := s.transferSvc.GetById(ctx, splitRoutingConfig.TransferID, payment.MerchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if transfer == nil {
		s.logger.Info(ctx, "transfer to destination is not found", logger.String("paymentId", paymentID), logger.String("transferId", transferID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed)
	}

	if transfer.Status != constant.StatusSuccess {
		s.logger.Info(ctx, "split routing is not processed yet to destination", logger.String("paymentId", paymentID), logger.String("transferId", transferID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed)
	}

	paymentReferenceID := ""
	if payment.ReferenceID != nil {
		paymentReferenceID = *payment.ReferenceID
	}
	return &splitRoutingPaymentModel.SplitRoutingPaymentDetailResponse{
		PaymentID:             paymentID,
		ClientReferenceID:     paymentReferenceID,
		Currency:              transfer.Currency,
		Amount:                transfer.Amount,
		Remarks:               splitRoutingConfig.Remarks,
		SourceMerchantID:      transfer.MerchantID.String(),
		DestinationMerchantID: transfer.RecipientID.String(),
		TransferID:            transfer.UUID.String(),
		CreatedAt:             transfer.CreatedAt,
		UpdatedAt:             transfer.UpdatedAt,
	}, nil
}

func (s *PaymentService) GetDetailByID(ctx context.Context, id string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetDetailByID")
	defer segment.End()

	// Find Payment
	payment, err := s.paymentRepo.GetPaymentById(ctx, id)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Info(ctx, constant.ErrPaymentNotFound.Error(), logger.String("paymentId", id))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	return payment, nil
}

func (s *PaymentService) GetActivePaymentByProcessorReferenceNumber(ctx context.Context, request *paymentModel.GetActivePaymentByProcessorReferenceNumberRequest) (*paymentModel.Payment, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/payment/GetActivePaymentByProcessorReferenceNumber")
	defer span.End()

	payment, err := s.paymentRepo.GetActivePaymentByProcessorReferenceNumber(ctx, request)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	return payment, nil
}

// Helper methods for status history

func (s *PaymentService) getPaymentStatusLabel(status string) string {
	if statusInfo, exists := constant.PaymentStatusHistoryLabelsAndDescriptions[status]; exists {
		if label, ok := statusInfo["label"]; ok {
			return label
		}
	}
	return status
}

func (s *PaymentService) getPaymentStatusDescription(status string) string {
	if statusInfo, exists := constant.PaymentStatusHistoryLabelsAndDescriptions[status]; exists {
		if description, ok := statusInfo["description"]; ok {
			return description
		}
	}
	return ""
}

func (s *PaymentService) getPaymentStatusRecommendation(status string) string {
	if statusInfo, exists := constant.PaymentStatusHistoryLabelsAndDescriptions[status]; exists {
		if recommendation, ok := statusInfo["recommendation"]; ok {
			return recommendation
		}
	}
	return ""
}

func (s *PaymentService) getChargeStatusFromTransaction(accountTransaction *orchestrator_model.AccountTransactionWithUseCase) string {
	if accountTransaction == nil {
		return ""
	}

	chargeStatus := accountTransaction.Status
	if accountTransaction.AdditionalInfo.Valid {
		var additionalInfo map[string]any
		if err := json.Unmarshal(accountTransaction.AdditionalInfo.JSONText, &additionalInfo); err == nil {
			if chargeStatusValue, exists := additionalInfo["chargeStatus"]; exists && chargeStatusValue != nil {
				if chargeStatusStr, ok := chargeStatusValue.(string); ok {
					chargeStatus = chargeStatusStr
				}
			}
		}
	}

	return chargeStatus
}

func (s *PaymentService) shouldShowAuthCaptureBanner(ctx context.Context, payment *paymentModel.Payment, accountTransaction *orchestrator_model.AccountTransactionWithUseCase) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/shouldShowAuthCaptureBanner")
	defer segment.End()

	chargeStatus := s.getChargeStatusFromTransaction(accountTransaction)

	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_CREDIT_CARD && payment.Status == constant.UnifiedPaymentSessionStatusProcessing && chargeStatus == constant.ChargeStatusWaitingForCapture {
		return true
	}

	return false
}

func (s *PaymentService) getAuthCaptureBannerFromConsul(ctx context.Context) *paymentModel.InfoBanner {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/getAuthCaptureBannerFromConsul")
	defer segment.End()

	defaultMessage := constant.PaymentUIAuthCaptureBannerDefaultMessage

	evalCtx := ffcontext.NewEvaluationContext(s.config.Environment)
	evalCtx.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, s.config.Environment)

	message, err := ffclient.StringVariation(
		constant.FeatureFlagKeyPaymentUIAuthCaptureBannerMessage,
		evalCtx,
		defaultMessage,
	)
	if err != nil {
		s.logger.Warn(ctx, "failed to get auth capture banner message from consul, using default",
			logger.Error(err))
		message = defaultMessage
	}

	return &paymentModel.InfoBanner{
		Message: message,
		Type:    "info",
	}
}

func (s *PaymentService) getCardFundedPayoutMetadata(ctx context.Context, payment *paymentModel.Payment) *paymentModel.CardFundedPayoutMetadata {
	if payment == nil {
		return nil
	}

	cardFundedPayout, err := s.disbursementRepo.GetCardFundedPayoutDetail(ctx, &cardFundedPayoutModel.GetPayoutDetailRequest{
		PayoutID:   util.ValueOfPtr(payment.ReferenceID),
		MerchantID: payment.MerchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get card funded payout data", logger.Error(err), logger.String("paymentID", payment.UUID))
		return nil
	}

	if cardFundedPayout == nil {
		s.logger.Warn(ctx, "card funded payout not found", logger.String("paymentID", payment.UUID))
		return nil
	}

	return &paymentModel.CardFundedPayoutMetadata{
		PayoutID:          cardFundedPayout.UUID,
		ReferenceID:       cardFundedPayout.ReferenceID,
		VendorName:        util.ValueOfPtr(cardFundedPayout.MetadataObj.CardFundedDetail).VendorName,
		BankAccountNumber: cardFundedPayout.AccountNumber,
		BankAccountName:   cardFundedPayout.AccountName,
		BankName:          cardFundedPayout.BankName,
		Remarks:           cardFundedPayout.Remarks,
		Amount:            cardFundedPayout.Amount,
		Fee:               cardFundedPayout.Fee,
		TotalAmount:       cardFundedPayout.TotalAmount,
		CreatedAt:         cardFundedPayout.CreatedAt,
	}
}
