package paymentService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledgerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentService) ProcessSplitRoute(ctx context.Context, paymentID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ProcessSplitRoute")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	parentMerchantID := payment.MerchantID
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, parentMerchantID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	if merchant.ParentID.Valid {
		parentMerchantID = merchant.ParentID.String
	}

	paymentSplitRouteConfig := &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{}
	if payment.Metadata != nil {
		paymentMetadata := *payment.Metadata
		raw, _ := json.Marshal(paymentMetadata[constant.SplitRoutingPaymentConfigKey])
		_ = json.Unmarshal(raw, paymentSplitRouteConfig)
	}

	if len(*paymentSplitRouteConfig) > 0 {
		parentMerchantUUID, _ := uuid.Parse(parentMerchantID)
		sourceMerchantUUID, _ := uuid.Parse(payment.MerchantID)

		paymentReferenceID := ""
		if payment.ReferenceID != nil {
			paymentReferenceID = *payment.ReferenceID
		}

		for idx, val := range *paymentSplitRouteConfig {
			if val.TransferID != "" {
				request := &ledgerModel.UpdateLedgerEntryRequest{
					ReferenceID: util.ParseUUID(val.TransferID),
					Usecase:     constant.ReferencePayment,
					Status:      constant.StatusSuccess,
					ReasonType:  "null", ReasonDescription: "null",
					SettlementStatus: constant.StatusSuccess,
					SettlementAt:     time.Now().UTC(),
				}

				if err := s.ledgerSvc.UpdateTransaction(ctx, request); err != nil {
					return err
				}

				if err = s.transferSvc.UpdateTransferStatus(ctx,
					sourceMerchantUUID.String(), val.TransferID, constant.TransferStatusSuccess, nil); err != nil {
					return err
				}

				continue
			}

			calculationAmount := val.FixedAmount
			if val.Type == constant.SplitRoutingPaymentTypePercentage {
				paymentAmount, _ := payment.TotalAmount.Float64()
				calculationAmount = (val.PercentageAmount / 100) * paymentAmount
			}

			newCtx := context.WithValue(ctx, constant.CtxSetBypassBalanceCheckTransaction, true)
			transferRequest := &transfer.TransferRequest{
				SourceMerchantID: sourceMerchantUUID,
				RecipientID:      val.MerchantId,
				ReferenceID:      paymentID,
				TransferType:     constant.MoneyFlowDirect,
				Amount:           calculationAmount,
				Remarks:          fmt.Sprintf("Routing Payment - Ref: %s", paymentReferenceID),
				ParentMerchantID: parentMerchantUUID,
				Usecase:          constant.TypePayment,
			}
			trf, errTrf := s.transferSvc.Transfer(newCtx, transferRequest)
			if errTrf != nil {
				return errTrf
			}

			(*paymentSplitRouteConfig)[idx].TransferID = trf.UUID.String()
		}

		// Update Payment Metadata
		paymentMetadata := *payment.Metadata
		paymentMetadata[constant.SplitRoutingPaymentConfigKey] = paymentSplitRouteConfig
		payment.Metadata = &paymentMetadata
		paymentDTO := payment.ToDTO()
		if errUpd := s.paymentRepo.UpdatePaymentData(ctx, paymentDTO); errUpd != nil {
			return pkgErr.New(response.HttpErrDatabase, errUpd)
		}
	}

	return nil
}
