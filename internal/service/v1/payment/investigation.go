package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) GetInvestigatedPayments(
	ctx context.Context,
	filter *paymentModel.GetInvestigatedPaymentsFilterRequest,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetInvestigatedPayments")
	defer segment.End()

	result, err := s.paymentRepo.GetInvestigatedPayments(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "failed to get investigated payments", logger.Error(err))
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	dtos, ok := result.Data.([]*paymentModel.InvestigatedPaymentDTO)
	if !ok {
		return result, nil
	}

	responses := make([]*paymentModel.InvestigatedPaymentResponse, 0, len(dtos))
	for _, dto := range dtos {
		responses = append(responses, dto.ToResponse())
	}

	result.Data = responses
	return result, nil
}

func (s *PaymentService) UpdateInvestigationStatus(
	ctx context.Context,
	paymentID string,
	request *paymentModel.UpdateInvestigationRequest,
) (*paymentModel.UpdateInvestigationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/UpdateInvestigationStatus")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentID)
	if err != nil {
		s.logger.Error(ctx, "failed to get payment by id", logger.Error(err))
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}
	if payment == nil {
		return nil, pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrPaymentNotFound)
	}

	if payment.ReasonType == nil || *payment.ReasonType == "" {
		return nil, pkgErr.New(httpResponse.HttpErrRequest, constant.ErrInvestigationNotFound)
	}

	if *payment.ReasonType != paymentConst.InvestigationStatusInProcess {
		return nil, pkgErr.New(httpResponse.HttpErrRequest, constant.ErrInvestigationAlreadyFinalized)
	}

	if request.InvestigationStatus != paymentConst.InvestigationStatusSuccess &&
		request.InvestigationStatus != paymentConst.InvestigationStatusFailed {
		return nil, pkgErr.New(httpResponse.HttpErrRequest, constant.ErrInvalidInvestigationStatus)
	}

	if request.Notes != nil && len(*request.Notes) > 200 {
		return nil, pkgErr.New(httpResponse.HttpErrRequest, constant.ErrInvestigationNotesExceedLimit)
	}

	completedAt := time.Now().UTC()

	err = s.paymentRepo.UpdateInvestigationStatus(
		ctx,
		paymentModel.UpdateInvestigationStatusRequest{
			PaymentID:   paymentID,
			Status:      request.InvestigationStatus,
			Notes:       request.Notes,
			CompletedAt: completedAt,
		},
	)
	if err != nil {
		s.logger.Error(ctx, "failed to update investigation status", logger.Error(err))
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	// Record investigation status change in status history
	s.recordInvestigationStatusHistory(ctx, paymentID, request.InvestigationStatus, constant.StatusHistoryActorCRM)

	return &paymentModel.UpdateInvestigationResponse{
		PaymentReferenceID:  paymentID,
		InvestigationStatus: request.InvestigationStatus,
		CompletedAt:         &completedAt,
		LastUpdatedAt:       completedAt,
		Notes:               request.Notes,
	}, nil
}

func (s *PaymentService) GetInvestigationProofOfPayment(
	ctx context.Context, request paymentModel.GetInvestigationProofOfPaymentRequest,
) (*paymentModel.GetInvestigationProofOfPaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetInvestigationProofOfPayment")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to retrieve payment details", logger.Error(err))
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	if payment == nil {
		return nil, pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrPaymentNotFound)
	}
	if payment.Metadata == nil || (*payment.Metadata)["investigationPoP"] == nil {
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, errors.New("payment is not under investigation"))
	}

	investigation := paymentModel.InvestigationPoPMetadata{}
	metadataJSON, _ := json.Marshal((*payment.Metadata)["investigationPoP"])

	_ = json.Unmarshal(metadataJSON, &investigation)
	if investigation.Path == "" {
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, errors.New("uploaded proof of transaction was not found for this payment"))
	}

	now := time.Now().UTC()
	signedURL, err := s.gcs.CreateSignedURL(ctx, investigation.Path, proofOfTransactionSignedURLExpiry)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate signed URL", logger.Error(err))
		return nil, pkgErr.New(httpResponse.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	return &paymentModel.GetInvestigationProofOfPaymentResponse{
		SignedURL:     signedURL,
		ExpiresAt:     now.Add(proofOfTransactionSignedURLExpiry),
		MerchantNotes: investigation.MerchantNotes,
	}, nil
}

func (s *PaymentService) recordInvestigationStatusHistory(ctx context.Context, paymentID, investigationStatus, actor string) {
	statusHistoryType := investigationStatus
	s.RecordPaymentStatusHistory(ctx, paymentID, actor, statusHistoryType)
}
