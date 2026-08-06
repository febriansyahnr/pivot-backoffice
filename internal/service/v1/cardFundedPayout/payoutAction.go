package cardFundedPayoutService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *service) CreatePayout(ctx context.Context, request model.CreatePayoutRequest) (*model.PayoutActionResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/CreatePayout")
	defer span.End()

	if s.disbursementRepo.CountByMerchantAndReference(ctx, request.MerchantID, request.ReferenceID) > 0 {
		return nil, pkgErrs.New(response.HttpErrConflict, constant.ErrReferenceIdExist)
	}

	vendorDetail, err := s.vendorSvc.Detail(ctx, request.VendorID)
	if err != nil {
		return nil, err

	} else if vendorDetail.MerchantID != request.MerchantID {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrVendorNotFound)

	} else if vendorDetail.Status != constant.StatusActive {
		return nil, pkgErrs.New(response.HttpErrForbidden, constant.ErrVendorNotActive)
	}

	cardRequest := model.GetSavedCardDetailRequest{
		CardID:     request.CardID,
		MerchantID: request.MerchantID,
	}
	cardDetail, err := s.customerSvc.GetCardFundedPayoutSavedCardDetail(ctx, cardRequest)
	if err != nil {
		return nil, err
	}

	feeRequest := &feeModel.GetFeeRequest{
		MerchantID:       request.MerchantID,
		Reference:        constant.ReferencePaymentFundedPayout,
		PaymentMethod:    paymentConst.PAYMENT_METHOD_CREDIT_CARD,
		Channel:          cardDetail.CardOrigin + "_" + cardDetail.PaymentChannel,
		ReferenceAmount:  request.Amount.Value,
		SettlementMethod: request.SettlementMethod,
	}
	_, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, feeRequest)
	if err != nil {
		return nil, err
	}
	feeDetail.FinalAmount = decimal.NewFromFloat(feeDetail.FinalAmount).Round(0).InexactFloat64()

	disbursement := request.ToCreateDisbursement(vendorDetail, cardDetail, feeDetail)
	if err := s.disbursementRepo.Insert(ctx, &disbursement); err != nil {
		s.logger.Error(ctx, "Failed to create disbursement data", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	_ = s.recordStatusHistory(ctx, disbursement.UUID, disbursement.Status, request.UserID, "")

	return &model.PayoutActionResponse{
		ID:               disbursement.UUID,
		VendorID:         vendorDetail.UUID,
		VendorName:       vendorDetail.Name,
		ReferenceID:      request.ReferenceID,
		FeeAmount:        disbursement.Fee.InexactFloat64(),
		Amount:           request.Amount,
		Remarks:          request.Remarks,
		SettlementMethod: request.SettlementMethod,
		CardID:           cardDetail.ID,
		CardName:         cardDetail.CardName,
		CreatedAt:        util.ValueToPtr(time.Now().UTC()),
	}, nil
}

func (s *service) RejectPayout(ctx context.Context, request model.RejectPayoutRequest) (*model.PayoutActionResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/RejectPayout")
	defer span.End()

	payout, err := s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, request.ID)
	if err != nil {
		s.logger.Error(ctx, "Failed to retrieve card-funded payout detail", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if payout == nil || payout.MerchantID != request.MerchantID {
		return nil, pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("payout with ID %s not found", request.ID))

	} else if payout.Status != constant.DisbursementStatusWaiting {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("payout must be in WAITING status; current status is %s", payout.Status))
	}

	now := time.Now().UTC()
	if err := s.disbursementRepo.Reject(ctx, payout.UUID, constant.DisbursementReasonTypeOther, request.Reason, request.UserID); err != nil {
		s.logger.Error(ctx, "Failed when reject card-funded payout", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	_ = s.recordStatusHistory(ctx, payout.UUID, constant.DisbursementStatusRejected, request.UserID, request.UserName)

	cardFundedDetail := payout.GetCardFundedPayoutDetail()

	var cardID, cardName string
	if cardFundedDetail.Card != nil {
		cardID = cardFundedDetail.Card.ID
		cardName = cardFundedDetail.Card.CardName
	}
	return &model.PayoutActionResponse{
		ID:          payout.UUID,
		VendorID:    cardFundedDetail.VendorID,
		VendorName:  cardFundedDetail.VendorName,
		ReferenceID: payout.ReferenceID,
		FeeAmount:   payout.Fee.InexactFloat64(),
		Amount: commonModel.AmountRequest{
			Currency: payout.Currency,
			Value:    payout.Amount.InexactFloat64(),
		},
		Remarks:          util.ValueOfPtr(payout.Remark),
		SettlementMethod: cardFundedDetail.SettlementMethod,
		CardID:           cardID,
		CardName:         cardName,
		RejectReason:     &request.Reason,
		RejectedAt:       util.ValueToPtr(now),
	}, nil
}
