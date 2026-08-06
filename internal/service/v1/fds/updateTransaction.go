package fdsservice

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *FdsService) UpdateTransaction(ctx context.Context, transactionID string, request *fdscommon.UpdateRequest) (*[]fdscommon.UpdateResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/fds/UpdateTransaction")
	defer span.End()

	// Check if transaction exist
	trx, err := s.accountTransactionsRepository.FindByID(ctx, transactionID)
	if err != nil || trx == nil {
		s.logger.Error(ctx, constant.ErrLedgerDetailNotFound.Error(), logger.Error(err))
		return nil, errors.New(response.HttpErrInternal, constant.ErrGetLedgerRecords)
	}

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceID,
		From:        constant.FDS_PROCESSOR,
		OriginId:    trx.UUID.String(),
		ReferenceId: trx.MerchantID.String(),
	})

	ccMetadata, err := trx.GetCreditcardMetadataFromAdditionalInfo()
	if err != nil || ccMetadata == nil {
		s.logger.Warn(ctx, "failed to get credit card metadata, continuing without card status mapping", logger.Error(err))
	}

	// Check if payment exist
	payment, err := s.paymentRepository.GetPaymentById(ctx, trx.ReferenceID)
	if err != nil || payment == nil {
		s.logger.Error(ctx, constant.ErrPaymentNotFound.Error(), logger.Error(err))
		return nil, errors.New(response.HttpErrInternal, constant.ErrPaymentNotFound)
	}

	// Check reference ID in rule evaluations
	ruleEvals, err := s.ruleEvaluationsRepository.GetByRefID(ctx, trx.UUID.String())
	if err != nil || len(*ruleEvals) == 0 {
		s.logger.Error(ctx, constant.ErrRuleEvaluationsNotFound.Error(), logger.Error(err))
		return nil, errors.New(response.HttpErrInternal, constant.ErrRuleEvaluationsNotFound)
	}

	updatedOn := time.Now().UTC()
	updateRequest := fdscommon.UpdateRequest{
		OrderTotal: &payment.Amount,
		UpdatedOn:  &updatedOn,
		Payment:    &fdscommon.PaymentUpdate{},
	}
	if request != nil {
		// use request from the parameter
		updateRequest = *request
		updateRequest.UpdatedOn = &updatedOn

		// Force payment status from payment data
		updateRequest.Payment.PaymentStatus = payment.Status
	}

	// Get rules for update transaction
	updateResults := []fdscommon.UpdateResponse{}
	for _, ruleEval := range *ruleEvals {
		rule, err := s.fraudRulesRepository.GetByID(ctx, ruleEval.RuleID)
		if err != nil || rule == nil {
			s.logger.Error(ctx, constant.ErrFraudRulesNotFound.Error(), logger.Error(err))
			continue
		}

		repo, ok := s.thirdPartyProcessor[rule.Provider.String]
		if !ok {
			s.logger.Error(ctx, "provider not found", logger.Error(err), logger.String("provider", rule.Provider.String))
			continue
		}

		if strings.Compare(rule.Provider.String, constant.PROVIDER_FRAUD_NET) == 0 {
			// default to failed authorization for now
			// this flow is used when failed in authorization only
			note := "failed authorization"
			trxStatus := MapTransactionStatusToFraudNet(trx.Status)
			paymentStatus := MapPaymentStatusToFraudNetStatus(payment.Status)

			// default when no authorization data
			cardStatus := constant.FRAUD_NET_CARD_STATUS_DECLINE
			if ccMetadata != nil && ccMetadata.AuthorizationData != nil {
				cardStatus = MapAcquirerResponseCodeToFraudNetCardStatus(ccMetadata.AuthorizationData.AcquirerResponseCode)
			}

			// Use rule evaluations ID as idempotent for fraud net
			updateRequest.OrderID = ruleEval.UUID

			if updateRequest.Status == "" {
				updateRequest.Status = trxStatus
			}

			if updateRequest.Payment.PaymentStatus == "" {
				updateRequest.Payment.PaymentStatus = paymentStatus
			}

			if updateRequest.Payment.CardStatus == nil {
				updateRequest.Payment.CardStatus = &cardStatus
			}

			if updateRequest.Note == nil || *updateRequest.Note == "" {
				updateRequest.Note = &note
			}
		}

		// Build full context and set OrderID for Sokratech
		if rule.Provider.String == constant.PROVIDER_SOKRATECH {
			fullContext, err := s.buildPaymentFullContext(ctx, trx, payment, ccMetadata)
			if err != nil {
				s.logger.Error(ctx, "error when building sokratech full context", logger.Error(err))
				continue
			}
			fullContext.Transaction.OrderID = ruleEval.UUID
			updateRequest.FullContext = fullContext
		}

		resp, err := repo.Update(ctx, &updateRequest)
		if err != nil {
			s.logger.Error(ctx, "error when request update", logger.Error(err), logger.String("provider", rule.Provider.String))
			return nil, err
		}

		if resp == nil || !resp.Success {
			s.logger.Error(ctx, "FDS update failed", logger.String("provider", rule.Provider.String), logger.Any("response", resp))
			return nil, errors.New(response.HttpErrInternal, constant.ErrFdsUpdateTransaction)
		}

		s.logger.Info(ctx, "FDS update results", logger.Any("response", resp))

		updateResults = append(updateResults, *resp)
	}

	// update fdsRiskAssesment based on the request
	if updateRequest.IsFraud != nil {
		updateFds := fdscommon.FdsRiskAssessment{
			IsFraud:          updateRequest.IsFraud,
			ChargebackStatus: "",
		}

		if updateRequest.Payment != nil {
			updateFds.ChargebackStatus = util.ValueOfPtr(updateRequest.Payment.ChargebackStatus)
			updateFds.ChargebackNotes = util.ValueOfPtr(updateRequest.Note)
		}

		err := s.UpdateFdsRiskAssessment(ctx, trx, updateFds)
		if err != nil {
			s.logger.Error(ctx, "error when update fdsRiskAssesment", logger.Error(err))
		}
	}

	return &updateResults, nil
}

func (s *FdsService) buildPaymentFullContext(ctx context.Context, trx *orchestrator_model.AccountTransactionWithUseCase, payment *paymentModel.Payment, ccMetadata *card.CreditcardMetadata) (*fdscommon.CheckRequest, error) {
	// Check if merchant exist
	merchant, err := s.merchantRepository.FindMerchantByID(ctx, trx.MerchantID.String())
	if err != nil || merchant == nil {
		return nil, errors.New(response.HttpErrInternal, constant.ErrMerchantNotFound)
	}

	// Build metadata from ccMetadata
	metadata := &fdscommon.CheckTransactionRequest{}
	if ccMetadata != nil {
		metadata.FromCcMetadata(ccMetadata)
	}

	// Fetch payment method to get channel_type for MID mapping
	var midNumber, midType, acquiringName *string
	if payment.PaymentMethodID != "" && s.paymentMethodRepository != nil {
		paymentMethod, pmErr := s.paymentMethodRepository.FindPaymentMethodByIdAndMerchant(ctx, payment.PaymentMethodID, merchant.UUID)
		if pmErr != nil {
			s.logger.Error(ctx, "failed to get payment method for MID mapping", logger.Error(pmErr))
		}
		if paymentMethod != nil {
			convertedMidType := constant.ChannelTypeToMidType(paymentMethod.ChannelType)
			midType = &convertedMidType
			acquiringName = &paymentMethod.Acquirer
		}
	}

	// Get customer by payment.customerId if exist
	var customer *customerModel.Customer
	if payment.CustomerID != "" {
		customer, err = s.customerRepository.GetCustomerById(ctx, payment.CustomerID, merchant.UUID)
		if err != nil || customer == nil {
			s.logger.Error(ctx, constant.ErrCustomerNotFound.Error(), logger.Error(err))
		}
	}

	paymentDetail := paymentModel.PaymentHistoryDetailResponse{}
	paymentDetail.LoadPaymentV2CustomerOrderInformation(payment, customer)

	// Build CheckRequest
	checkRequest := fdscommon.CheckRequest{
		Account: fdscommon.AccountCheck{
			AccountID: &merchant.UUID,
		},
		Customer: fdscommon.CustomerCheck{
			Address1: &metadata.BillingInformation.Address1,
		},
		Payment: fdscommon.PaymentCheck{
			Type:          util.ValueToPtr(payment.GetGroupPaymentType()),
			CardAccountID: payment.ReferenceID,
			ActualAmt:     &payment.Amount,
			ActualCcy:     &payment.Currency,
			BilledAmt:     &payment.Amount,
			BilledCcy:     &payment.Currency,
			MethodType:    payment.PaymentMethod.Type,
			ThreeDsMethod: metadata.ThreeDsMethod,
		},
		Partner: fdscommon.PartnerCheck{
			Address1:   &merchant.Address,
			Name:       &merchant.ShortName,
			Phone:      &merchant.PICPhone,
			PostalCode: &merchant.PostCode,
			ID:         merchant.UUID,
			Company:    &merchant.Name,
			Email:      &merchant.PICEmail,
			RiskLevel:  merchant.RiskLevel.String,
		},
		Transaction: fdscommon.TransactionCheck{
			OrderCurrency:     &payment.Currency,
			OrderID:           payment.UUID,
			OrderTotal:        &payment.Amount,
			OrderedOn:         &payment.CreatedAt,
			TransactionID:     payment.ProcessorReferenceNumber,
			ID:                payment.UUID,
			ClientReferenceID: util.ValueOfPtr(payment.ReferenceID),
			CreatedAt:         payment.CreatedAt,
			UpdatedAt:         payment.UpdatedAt,
		},
		IB: fdscommon.IntermediaryBankCheck{},
		Custom: &fdscommon.CustomCheck{
			Number:        midNumber,
			Type:          midType,
			AcquiringName: acquiringName,
		},
	}

	if metadata.Device != nil {
		checkRequest.Device = *metadata.Device
	}
	if checkRequest.Device.IPAddress != nil && *checkRequest.Device.IPAddress != "" {
		checkRequest.Device.IPType = util.GetIPVersion(*checkRequest.Device.IPAddress)
	}

	if metadata.CardData != nil {
		binLength := s.GetBinLength()
		trimmedBin := util.TrimLengthRight(metadata.CardData.First8Digit, int(binLength))

		checkRequest.Payment.PaymentID = &metadata.CardData.Fingerprint
		checkRequest.Customer.ID = metadata.CardData.Fingerprint
		checkRequest.Transaction.UserID = &metadata.CardData.Fingerprint
		checkRequest.Payment.Bin = &trimmedBin
		checkRequest.Payment.Last4 = metadata.CardData.Last4Digit
		checkRequest.Payment.First8 = metadata.CardData.First8Digit
		checkRequest.Payment.Fingerprint = metadata.CardData.Fingerprint
		checkRequest.Payment.MaskedCardNumber = util.TrimLengthRight(metadata.CardData.First8Digit, 6) + "xxxxxx" + metadata.CardData.Last4Digit
		checkRequest.Payment.CardBrand = metadata.CardData.CardBrand
		checkRequest.Payment.CardCountryCode = metadata.CardData.CountryCode
		checkRequest.Payment.CardType = metadata.CardData.CardType
		checkRequest.Payment.CardIssuing = metadata.CardData.CardIssuing
	}

	if metadata.AuthenticationData != nil {
		checkRequest.Payment.ThreeDsEci = &metadata.AuthenticationData.EciCode
		checkRequest.Payment.ThreeDsXid = &metadata.AuthenticationData.XID
	}

	if metadata.AuthorizationData != nil {
		checkRequest.IB.ID = &metadata.AuthorizationData.AcquirerTransactionID
		checkRequest.Payment.CvvResultCode = &metadata.AuthorizationData.CvvResult
		checkRequest.Payment.AuthCode = &metadata.AuthorizationData.ApprovalCode
		checkRequest.Payment.AuthResCode = &metadata.AuthorizationData.AcquirerResponseCode
	}

	if payment.CustomerID != "" {
		checkRequest.Customer.ID = payment.CustomerID
		if paymentDetail.CustomerInfo != nil && customer != nil {
			checkRequest.Customer.Email = &paymentDetail.CustomerInfo.Email
			checkRequest.Customer.FirstName = &customer.FirstName
			checkRequest.Customer.LastName = &customer.LastName
			checkRequest.Customer.Phone = &customer.PhoneNumber
		}
	}

	if paymentDetail.OrderInfo != nil && len(paymentDetail.OrderInfo.ProductDetails) > 0 {
		orderType := paymentDetail.OrderInfo.ProductDetails[0].Type
		orderIsDigital := orderType == constant.ProductDetailTypeDigital
		checkRequest.Transaction.OrderIsDigital = &orderIsDigital
	}

	return &checkRequest, nil
}

func (s *FdsService) UpdateFdsRiskAssessment(ctx context.Context, trx *orchestrator_model.AccountTransactionWithUseCase, updateFds fdscommon.FdsRiskAssessment) error {
	// either get from existing or create if not exist
	additionalInfo := make(map[string]any)
	if trx.AdditionalInfo.Valid && len(trx.AdditionalInfo.JSONText) > 0 {
		if err := json.Unmarshal(trx.AdditionalInfo.JSONText, &additionalInfo); err != nil {
			s.logger.Error(ctx, "failed to unmarshal existing additional_info", logger.Error(err))
			// additionalInfo is already initialized above, so no need to reassign
		}
	}

	var existingFds *fdscommon.FdsRiskAssessment
	if fdsData, ok := additionalInfo[constant.FdsRiskAssesment]; ok {
		fdsBytes, err := json.Marshal(fdsData)
		if err != nil {
			s.logger.Error(ctx, "failed to marshal existing FDS data", logger.Error(err))
		} else {
			existingFds = &fdscommon.FdsRiskAssessment{}
			if err := json.Unmarshal(fdsBytes, existingFds); err != nil {
				s.logger.Error(ctx, "failed to unmarshal existing FDS data", logger.Error(err))
				existingFds = nil
			}
		}
	}

	if existingFds == nil {
		existingFds = &fdscommon.FdsRiskAssessment{}
	}

	existingFds.Update(&updateFds)

	additionalInfo[constant.FdsRiskAssesment] = existingFds
	additionalInfoBytes, err := json.Marshal(additionalInfo)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal additional_info", logger.Error(err))
		return err
	}

	nullJSONText := types.NullJSONText{
		JSONText: additionalInfoBytes,
		Valid:    true,
	}

	err = s.accountTransactionsRepository.UpdateAdditionalInfoByID(ctx, trx.UUID.String(), nullJSONText)
	if err != nil {
		s.logger.Error(ctx, "failed to update additional_info", logger.Error(err))
		return err
	}

	return nil
}
