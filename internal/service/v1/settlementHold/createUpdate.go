package settlementHoldService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *settlementHoldService) CreateUpdate(ctx context.Context, request *settlementHold.CreateUpdateSettlementHoldRequest) (*settlementHold.CreateUpdateSettlementHoldResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlementHold/CreateUpdate")
	defer segment.End()

	var (
		newHistory = &settlementHold.SettlementHoldHistory{}
		err        error
	)
	payment, err := s.paymentSvc.GetDetailByID(ctx, request.PaymentID)
	if err != nil {
		s.logger.Error(ctx, "failed validate payment id", logger.Error(err), logger.String("paymentId", request.PaymentID))
		return nil, err
	}
	request.MerchantID = payment.MerchantID

	settlementHoldRecord, err := s.repo.GetByPaymentID(ctx, payment.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when get settlement hold by payment id", logger.Error(err), logger.String("paymentId", request.PaymentID))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrValidateSettlementHold)
	}

	s.logger.Info(ctx, "trigger action settlement hold", logger.String("action", request.Action), logger.String("paymentId", request.PaymentID))
	err = s.settlementSvc.ProcessSettlementHoldOrRelease(ctx, &settlementModel.ProcessHoldReleaseSettlementRequest{
		ReferenceID:    request.PaymentID,
		Action:         request.Action,
		LastActionTime: time.Now().UTC(),
	})
	if err != nil {
		s.logger.Error(ctx, "error process settlement hold", logger.Error(err))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrProcessSettlementHoldRelease)
	}

	if settlementHoldRecord != nil {
		newHistory = settlementHoldRecord.Update(request)
		err = s.repo.Update(ctx, settlementHoldRecord, newHistory)
	} else {
		settlementHoldRecord, newHistory = settlementHold.New(request)
		err = s.repo.Create(ctx, settlementHoldRecord, newHistory)
	}
	if err != nil {
		s.logger.Error(ctx, "error stored settlement hold updates", logger.Error(err))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrStoredSettlementHold)
	}

	return &settlementHold.CreateUpdateSettlementHoldResponse{
		UUID:       settlementHoldRecord.UUID,
		MerchantID: request.MerchantID,
		PaymentID:  request.PaymentID,
		Status:     request.Action,
		Reason:     request.Reason,
		CreatedBy:  settlementHoldRecord.CreatedBy,
		CreatedAt:  settlementHoldRecord.CreatedAt,
		UpdatedBy:  request.CreatedBy,
		UpdatedAt:  settlementHoldRecord.UpdatedAt,
	}, nil
}
