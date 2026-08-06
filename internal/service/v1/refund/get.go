package refundService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// FindByID retrieves a refund by its ID.
// This function used for processing not present into user interface
//
// It returns the refund if found, or an appropriate error in the following cases:
// - Database error: Returns a wrapped database error
// - Refund not found: Returns an unprocessable content error with ErrRefundNotFound
func (s *RefundService) FindByID(ctx context.Context, refundID string) (*refundModel.Refund, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/FindByID")
	defer span.End()

	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	} else if refund == nil {
		s.logger.Warn(ctx, "[FindRefundByID] refund not found", logger.String("refundID", refundID))
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrRefundNotFound)
	}

	return refund, nil
}

// GetRefundList retrieves a paginated list of refunds based on the provided filter criteria.
// The function accepts a context and a request object containing various filter parameters.
func (s *RefundService) GetRefundList(ctx context.Context, request refundModel.FilterRefundRequest) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetRefunds")
	defer span.End()

	refunds, err := s.refundRepo.GetRefundList(ctx, request)
	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	return refunds, nil
}

// GetRefundDetail retrieves the detailed information of a refund based on the provided filter criteria.
// Unlike a traditional FindById operation which typically looks up a record by its primary key,
// GetRefundDetail uses a filter-based approach through the refundModel.FilterRefundRequest.
// This func returns detailed information of the refund, including its status, amount, and other relevant fields.
// Note: This method does NOT include status history. Use GetRefundDetailWithStatusHistories for that.
func (s *RefundService) GetRefundDetail(ctx context.Context, request refundModel.FilterRefundRequest) (*refundModel.RefundResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetRefundDetail")
	defer span.End()

	result, err := s.refundRepo.GetRefundList(ctx, request)
	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	refunds := result.Data.([]*refundModel.RefundResponse)
	if len(refunds) == 0 {
		s.logger.Warn(ctx, "[GetRefundDetail] refund not found", logger.String("refundID", request.UUID))
		return nil, pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrRefundNotFound)
	}

	return refunds[0], nil
}

// GetRefundDetailWithStatusHistories retrieves detailed refund information including status history and channel destination.
func (s *RefundService) GetRefundDetailWithStatusHistories(ctx context.Context, request refundModel.FilterRefundRequest) (*refundModel.RefundResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetRefundDetailWithStatusHistories")
	defer span.End()

	refund, err := s.refundRepo.GetRefundByID(ctx, request.UUID, request.MerchantID)
	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	if refund == nil {
		s.logger.Warn(ctx, "[GetRefundDetailWithStatusHistories] refund not found", logger.String("refundID", request.UUID))
		return nil, pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrRefundNotFound)
	}

	// Fetch and populate status history
	if s.statusHistoriesRepo != nil {
		statusHistoryList, err := s.statusHistoriesRepo.GetByReference(ctx, constant.TypeRefund, refund.ID)
		if err != nil {
			s.logger.Warn(ctx, "[GetRefundDetailWithStatusHistories] failed to get refund status history", logger.String("refundId", refund.ID), logger.Error(err))
		} else if len(statusHistoryList) > 0 {
			refund.StatusHistory = s.buildRefundStatusHistoryResponse(statusHistoryList)
		}
	}

	return refund, nil
}

// buildRefundStatusHistoryResponse converts status history records to RefundStatusHistoryResponse slice
func (s *RefundService) buildRefundStatusHistoryResponse(histories []*statusHistoryModel.StatusHistory) []unifiedPaymentModel.RefundStatusHistoryResponse {
	result := make([]unifiedPaymentModel.RefundStatusHistoryResponse, 0, len(histories))
	for _, h := range histories {
		entry := unifiedPaymentModel.RefundStatusHistoryResponse{
			Status:    h.Status,
			Timestamp: &h.CreatedAt,
		}
		if h.MetadataObj != nil {
			entry.Label = h.MetadataObj.Label
			entry.Description = h.MetadataObj.Description
			entry.Recommendation = h.MetadataObj.Recommendation
		}
		result = append(result, entry)
	}
	return result
}

func (s *RefundService) GetExistingRefundList(ctx context.Context, request refundModel.GetExistingRefundListRequest) ([]refundModel.RefundResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetExistingRefundList")
	defer span.End()

	return s.refundRepo.ListByPaymentID(ctx, request.PaymentID, refundModel.ListByPaymentIDRequest{
		Status: request.Status,
	})
}

func (s *RefundService) GetTotalRefundedAmount(ctx context.Context, paymentID string) (float64, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetTotalRefundedAmount")
	defer span.End()

	return s.refundRepo.GetTotalRefundedAmount(ctx, paymentID)
}
