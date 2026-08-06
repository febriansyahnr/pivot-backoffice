package cardFundedPayoutService

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) ApprovePayout(ctx context.Context, request model.ApprovePayoutRequest) (*model.PayoutActionResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/ApprovePayout")
	defer span.End()

	paymentMethodRequest := paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: request.MerchantID,
		Category:   paymentConst.PAYMENT_METHOD_CATEGORY_PAYMENT,
		Type:       paymentConst.PAYMENT_METHOD_CREDIT_CARD,
	}
	paymentMethod, err := s.paymentMethodRepo.GetActivePaymentMethodByRequest(ctx, &paymentMethodRequest)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if paymentMethod == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotFound)

	} else if !paymentMethod.IsCardPartnerConfigFound() {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrPartnerConfigNotFound)
	}

	request.Payout, err = s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, request.ID)
	if err != nil {
		s.logger.Error(ctx, "Failed to retrieve card-funded payout detail", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if request.Payout == nil || request.Payout.MerchantID != request.MerchantID {
		return nil, pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("payout with ID %s not found", request.ID))

	} else if request.Payout.Status != constant.DisbursementStatusWaiting {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("payout must be in WAITING status; current status is %s", request.Payout.Status))
	}

	cardFundedDetail := request.Payout.GetCardFundedPayoutDetail()
	if cardFundedDetail.Card != nil {
		request.CardID = cardFundedDetail.Card.ID
		request.CardName = cardFundedDetail.Card.CardName
	}

	card, err := s.customerSvc.GetCardFundedPayoutSavedCardDetail(ctx, model.GetSavedCardDetailRequest{
		MerchantID: request.MerchantID,
		CardID:     request.CardID,
	})
	if err != nil {
		return nil, err
	}
	request.CardToken = card.CardToken
	request.VendorID = cardFundedDetail.VendorID
	request.VendorName = cardFundedDetail.VendorName
	request.SettlementMethod = cardFundedDetail.SettlementMethod

	// Setup CFP Processor, use MPGS as default processor limit
	if !paymentMethod.EnableCardFundedPayout() {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("card-funded payout are not allowed"))
	}
	err = request.PrepareCfpProcessor(paymentMethod.MerchantConfigObj.CardFundedPayoutConfig, paymentMethod.MerchantConfigObj.PartnerConfig, s.config.CardFundedPayout.PartnerProcessorLimit.MPGS)
	if err != nil {
		return nil, err
	}

	ctxTx, err := s.disbursementRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to create a new database transaction session", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	isComplete := false
	defer func() {
		if isComplete {
			return
		}
		if rollbackErr := s.disbursementRepo.RollbackTransaction(context.WithoutCancel(ctxTx)); rollbackErr != nil {
			s.logger.Error(ctx, "Failed to rollback changes", logger.Error(rollbackErr))
		}
	}()

	// Payment creation does not use isolated sessions due to multiple sub-processes creating their own sessions, which complicates session management.
	// Therefore, rollback is handled using permanent deletion (hard delete) instead of relying on transaction sessions (begin/rollback).
	authenticationURL, cancelFn, err := s.createPayoutPayment(ctx, request)
	if err != nil {
		return nil, err
	}
	defer func() {
		if isComplete {
			return
		}
		if cancelErr := cancelFn(); cancelErr != nil {
			s.logger.Error(ctx, "Failed to cancel the payment session", logger.Error(cancelErr))
		}
	}()

	now := time.Now().UTC()
	if err := s.disbursementRepo.ApproveInBulk(ctxTx, []string{request.ID}, request.UserID); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	_ = s.recordStatusHistory(ctxTx, request.ID, constant.DisbursementStatusApproved, request.UserID, request.UserName)

	if err := s.disbursementRepo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "Failed to commit changes", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	isComplete = true

	return &model.PayoutActionResponse{
		ID:          request.Payout.UUID,
		VendorID:    cardFundedDetail.VendorID,
		VendorName:  cardFundedDetail.VendorName,
		ReferenceID: request.Payout.ReferenceID,
		FeeAmount:   request.Payout.Fee.InexactFloat64(),
		Amount: commonModel.AmountRequest{
			Currency: request.Payout.Currency,
			Value:    request.Payout.Amount.InexactFloat64(),
		},
		Remarks:           util.ValueOfPtr(request.Payout.Remark),
		SettlementMethod:  cardFundedDetail.SettlementMethod,
		CardID:            request.CardID,
		CardName:          request.CardName,
		AuthenticationUrl: &authenticationURL,
		ApprovedAt:        util.ValueToPtr(now),
	}, nil
}

const defaultExpiryPaymentMinutes = 30

func (s *service) createPayoutPayment(ctx context.Context, request model.ApprovePayoutRequest) (authenticationURL string, cancelFn func() error, err error) {
	var (
		feeConfig      = request.Payout.MetadataObj.FeeDetail
		payoutAmount   = request.Payout.Amount.InexactFloat64()
		processorLimit = request.ProcessorLimit
		responses      = []unifiedPaymentModel.UnifiedPaymentSessionResponse{} // Currently, only the first response is required, but it may evolve in the future.
		redirectURL    = s.config.MerchantPortalConfig.CardFundedPayoutURL + "/detail/" + request.Payout.UUID
	)

	if processorLimit <= 0 {
		processorLimit = request.Payout.TotalAmount.InexactFloat64()
	}
	cancelFn = func() error {
		return s.paymentRepo.HardDeleteCardFundedPayoutPayments(context.WithoutCancel(ctx), request.MerchantID, request.Payout.UUID)
	}
	defer func() {
		if err == nil {
			return
		}
		if cancelErr := cancelFn(); cancelErr != nil {
			s.logger.Error(ctx, "Failed to cancel the payment session", logger.Error(cancelErr))
		}
	}()

	sequence := 1
	count := int(math.Ceil(payoutAmount / processorLimit))
	for payoutAmount > 0 {
		paymentAmount := processorLimit
		if payoutAmount < processorLimit {
			paymentAmount = payoutAmount
		}
		// The fee calculation for card-funded payouts always uses a percentage,
		// and the value is determined at the time the payout request is created.
		feeAmount := decimal.NewFromFloat((feeConfig.Percentage / 100) * paymentAmount).Round(0).InexactFloat64()
		feeConfig.FinalAmount = feeAmount
		feeConfig.TrxAmount = paymentAmount

		expiryAfterMinutes := s.config.CardFundedPayout.ExpiryAfterMinutes
		if expiryAfterMinutes <= 0 {
			expiryAfterMinutes = defaultExpiryPaymentMinutes
		}

		// Initialize a payment request for card-funded transactions.
		paymentRequest := unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
			PaymentID:         util.GenerateUUID().String(),
			ClientReferenceID: request.Payout.UUID,
			PaymentMethod: &unifiedPaymentModel.PaymentMethod{
				Type: constant.UnifiedPaymentMethodCard,
			},
			PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
				Card: &unifiedPaymentModel.PaymentMethodOptionCard{
					ThreeDsMethod: constant.CardThreeDsMethodChallenge,
					ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
						BankMerchantId: request.CitAcquirerMerchantID,
					},
				},
			},
			AutoConfirm: true,
			Mode:        constant.UnifiedPaymentModeRedirect,
			ExpiryAt:    time.Now().Add(time.Duration(expiryAfterMinutes) * time.Minute).UTC(),
			Amount: unifiedPaymentModel.Amount{
				Currency: request.Payout.Currency,
				Value:    paymentAmount,
			},
			MerchantID:  request.MerchantID,
			PaymentType: constant.TypeCardFundedPayout,
			CustomerID:  request.CardID,
			CreatedBy:   request.UserID,
			CreatedFrom: constant.SourceDashboard,
			CardFundedPayout: &unifiedPaymentModel.CardFundedPayout{
				Sequence:         sequence,
				Count:            count,
				SettlementMethod: request.SettlementMethod,
				VendorID:         request.VendorID,
				VendorName:       request.VendorName,
				FeeAmount:        feeAmount,
				FeeConfig:        feeConfig,
				CardID:           request.CardID,
				CardToken:        request.CardToken,
			},
		}
		if sequence == 1 {
			// The first transaction requires an authentication process.
			paymentRequest.PaymentMethod.CardPaymentMethodDetail = &unifiedPaymentModel.CardPaymentMethodDetail{
				CVC:   request.CVC,
				Token: request.CardToken,
			}
			// Back to dashboard after authentication.
			paymentRequest.RedirectUrl = unifiedPaymentModel.RedirectUrl{
				SuccessReturnUrl:    redirectURL,
				FailureReturnUrl:    redirectURL,
				ExpirationReturnUrl: redirectURL,
			}
		} else {
			// Subsequent transactions use a separate MID and require no authentication.
			paymentRequest.CardFundedPayout.FirstPaymentID = responses[0].ID
			paymentRequest.PaymentMethodOptions.Card.ThreeDsMethod = constant.CardThreeDsMethodNever
			paymentRequest.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId = request.MitAcquirerMerchantID
		}

		paymentResponse, err := s.unifiedPaymentSvc.CreateSession(ctx, &paymentRequest)
		if err != nil {
			return "", cancelFn, err
		}

		responses = append(responses, *paymentResponse)
		sequence, payoutAmount = (sequence + 1), (payoutAmount - processorLimit)
	}

	if len(responses) == 0 {
		return "", cancelFn, errors.New("payment transaction cannot be processed")
	}
	return responses[0].PaymentUrl, cancelFn, nil
}
