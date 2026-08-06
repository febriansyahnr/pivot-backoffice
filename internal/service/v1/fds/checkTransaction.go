package fdsservice

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *FdsService) CheckTransaction(ctx context.Context, transactionID string, request *fdscommon.CheckTransactionRequest) (*fdscommon.CheckTransactionResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/fds/CheckTransaction")
	defer span.End()

	resp := fdscommon.CheckTransactionResponse{
		Status: constant.FDS_STATUS_NOT_EVALUATED,
		Score:  decimal.NewFromInt(0),
	}

	// Check if transaction exist
	trx, err := s.accountTransactionsRepository.FindByID(ctx, transactionID)
	if err != nil || trx == nil {
		s.logger.Error(ctx, constant.ErrLedgerDetailNotFound.Error(), logger.Error(err))
		return &resp, errors.New(response.HttpErrInternal, constant.ErrGetLedgerRecords)
	}

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceID,
		From:        constant.FDS_PROCESSOR,
		OriginId:    trx.UUID.String(),
		ReferenceId: trx.MerchantID.String(),
	})

	metadata := request
	if request == nil {
		// Initialize metadata if it's nil
		metadata = &fdscommon.CheckTransactionRequest{}
		ccMetadata, err := trx.GetCreditcardMetadataFromAdditionalInfo()
		if err != nil || ccMetadata == nil {
			s.logger.Error(ctx, "failed to get metadata", logger.Error(err))
			return &resp, errors.New(response.HttpErrInternal, err)
		}

		metadata.FromCcMetadata(ccMetadata)
	}

	// Check if payment exist
	payment, err := s.paymentRepository.GetPaymentById(ctx, trx.ReferenceID)
	if err != nil || payment == nil {
		s.logger.Error(ctx, constant.ErrPaymentNotFound.Error(), logger.Error(err))
		return &resp, errors.New(response.HttpErrInternal, constant.ErrPaymentNotFound)
	}

	// Check if merchant exist
	merchant, err := s.merchantRepository.FindMerchantByID(ctx, trx.MerchantID.String())
	if err != nil || merchant == nil {
		s.logger.Error(ctx, constant.ErrMerchantNotFound.Error(), logger.Error(err))
		return &resp, errors.New(response.HttpErrInternal, constant.ErrMerchantNotFound)
	}

	// Fetch payment method to get channel_type for MID mapping
	var midNumber, midType, acquiringName *string
	if metadata.MidNumber != nil {
		midNumber = metadata.MidNumber
	}
	if metadata.AcquiringName != nil {
		acquiringName = metadata.AcquiringName
	}

	if payment.PaymentMethodID != "" && s.paymentMethodRepository != nil {
		paymentMethod, pmErr := s.paymentMethodRepository.FindPaymentMethodByIdAndMerchant(ctx, payment.PaymentMethodID, merchant.UUID)
		if pmErr != nil {
			s.logger.Error(ctx, "failed to get payment method for MID mapping", logger.Error(pmErr))
		}
		if paymentMethod != nil {
			convertedMidType := constant.ChannelTypeToMidType(paymentMethod.ChannelType)
			midType = &convertedMidType

			// assign if acquirer is empty
			if acquiringName == nil {
				acquiringName = &paymentMethod.Acquirer
			}
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

	// link the fingerprintID into payment session
	// need to confirm about static payment, because users can access it several times
	// recomendation to store it into charge
	if util.ValueOfPtr(metadata.Device).FingerprintID != nil {
		err = s.paymentRepository.UpdatePaymentMetadataById(ctx, payment.UUID, paymentModel.UpdatePaymentMetadataRequest{
			FingerprintID: *metadata.Device.FingerprintID,
		})
		if err != nil {
			s.logger.Error(ctx, "failed to update payment metadata", logger.Error(err), logger.String("paymentID", payment.UUID))
			return nil, err
		}
	}

	// Get rules related to the reference type
	rules, _, err := s.fraudRulesRepository.List(ctx, &fraudrulesmodel.FraudRulesQuery{
		ReferenceType: trx.Channel,
	})
	if err != nil || len(rules) == 0 {
		s.logger.Error(ctx, constant.ErrFraudRulesNotFound.Error(), logger.Error(err))
		return &resp, errors.New(response.HttpErrInternal, constant.ErrFraudRulesNotFound)
	}

	checkRequest := fdscommon.CheckRequest{
		Account: fdscommon.AccountCheck{
			// use merchant UUID for account ID
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

		// payment ID use card fingerprint (CardFingerprint field from CardDataRequest)
		checkRequest.Payment.PaymentID = &metadata.CardData.Fingerprint
		// customer ID will use card fingerprint as default
		// if customer info exist it will use customer id
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

	// Adjust customer_id logic based on acceptance criteria:
	// IF customer_id exists in merchant's create payment session request, then use customer_id
	// ELSE, use card_id (CardFingerprint)
	if payment.CustomerID != "" {
		// Use customer_id from payment session request
		checkRequest.Customer.ID = payment.CustomerID
		// Set additional customer info if available
		if paymentDetail.CustomerInfo != nil && customer != nil {
			checkRequest.Customer.Email = &paymentDetail.CustomerInfo.Email
			checkRequest.Customer.FirstName = &customer.FirstName
			checkRequest.Customer.LastName = &customer.LastName
			checkRequest.Customer.Phone = &customer.PhoneNumber
		}
	} else if metadata.BillingInformation.Email != "" {
		checkRequest.Customer.Email = &metadata.BillingInformation.Email
		checkRequest.Customer.FirstName = &metadata.BillingInformation.GivenName
		checkRequest.Customer.LastName = &metadata.BillingInformation.Surname
	}
	// If no customer_id in request, card_id (CardFingerprint) is already set above in the CardData section

	if paymentDetail.OrderInfo != nil && len(paymentDetail.OrderInfo.ProductDetails) > 0 {
		// only get from the first object
		orderType := paymentDetail.OrderInfo.ProductDetails[0].Type
		orderIsDigital := orderType == constant.ProductDetailTypeDigital
		checkRequest.Transaction.OrderIsDigital = &orderIsDigital
	}

	// Eval the transaction (using internal or 3rd party)
	// will call here
	totalScore := decimal.NewFromInt(0)
	evalResults := []fdscommon.EvalResult{}
	for _, rule := range rules {
		// internal rules for later use
		if !rule.Provider.Valid {
			continue
		}

		repo, ok := s.thirdPartyProcessor[rule.Provider.String]
		if !ok {
			s.logger.Error(ctx, "provider not found", logger.Error(err), logger.String("provider", rule.Provider.String))
			continue
		}

		// rule eval ID will serve as idempotent key for the fds processor
		ruleEvalID, _ := uuid.NewV7()
		checkRequest.Transaction.OrderID = ruleEvalID.String()

		resp, err := repo.Check(ctx, &checkRequest)
		if err != nil {
			s.logger.Error(ctx, "error when request check", logger.Error(err), logger.String("provider", rule.Provider.String))
			continue
		}

		evalResult := fdscommon.EvalResult{
			Success: resp.Success,
			Code:    resp.Code,
			Source:  resp.Source,
			Message: resp.Message,
			Data:    &resp.Data,
			Weight:  rule.Weight,
		}

		if !resp.Success {
			evalResults = append(evalResults, evalResult)
			continue
		}

		// insert result to the rule evaluations
		score := rule.Weight.Mul(decimal.NewFromInt(int64(resp.Data.RiskScore)))
		ruleEval := ruleevaluationsmodel.RuleEvaluations{
			UUID:        ruleEvalID.String(),
			ReferenceID: trx.UUID.String(),
			RuleID:      rule.UUID,
			EvaluatedAt: time.Now().UTC(),
			Score:       score,
			Result:      resp.Data.RiskGroup,
			Provider:    rule.Provider.String,
		}

		err = s.ruleEvaluationsRepository.Create(ctx, &ruleEval)
		if err != nil {
			s.logger.Error(ctx, "error when inserting data into rule eval", logger.Error(err))
		}

		totalScore = totalScore.Add(score)
		evalResult.RuleEvaluation = &ruleEval
		evalResults = append(evalResults, evalResult)
	}

	fdsStatus := constant.FDS_STATUS_NOT_EVALUATED
	if len(evalResults) > 0 {
		fdsStatus = constant.FDS_STATUS_PASSED
	}

	scoreThreshold := s.GetScoreThreshold()
	if totalScore.Cmp(decimal.NewFromInt(scoreThreshold)) >= 0 {
		fdsStatus = constant.FDS_STATUS_REJECTED
	}

	resp.Status = fdsStatus
	resp.Score = totalScore
	resp.EvalResults = &evalResults

	// send slack alert if status rejected
	if resp.Status == constant.FDS_STATUS_REJECTED {
		var cardID string
		if metadata.CardData != nil {
			cardID = metadata.CardData.Fingerprint
		}
		err := s.SendFdsSlackAlert(ctx, transactionID, cardID, payment, merchant, resp)
		if err != nil {
			s.logger.Error(ctx, "failed to send slack alert", logger.Error(err))
		}
	}

	if s.cfg.Environment == constant.EnvironmentStaging {
		s.ModifyResponseForCardSimulation(ctx, &resp)
	}

	err = s.saveFdsRiskAssessmentToLedger(ctx, trx, &resp, evalResults)
	if err != nil {
		s.logger.Error(ctx, "failed to save FDS risk assessment to ledger", logger.Error(err))
	}

	return &resp, nil
}
