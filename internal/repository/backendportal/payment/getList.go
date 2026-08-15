package paymentRepository

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/paper-indonesia/pdk/v2/logger"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"golang.org/x/sync/errgroup"
)

func (r *PaymentRepository) GetList(
	ctx context.Context,
	filter *paymentModel.GetListFilterRequest,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		mu   sync.Mutex
		data = make([]*paymentModel.PaymentWithPaymentMethodDTO, 0)
		errG = new(errgroup.Group)
	)

	if filter.Page < 1 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PerPage

	query := `
			SELECT 
				p.uuid, 
				p.reference_id,
				p.merchant_id, 
				p.customer_id, 
				p.payment_method_id, 
				p.processor_reference_number, 
				p.currency, 
				p.amount, 
				p.fee, 
				p.discount, 
				p.total_amount, 
				p.status, 
				p.type,
				p.metadata, 
				p.payment_url,
				p.created_at, 
				p.expired_at,
				p.updated_at, 
				p.deleted_at,
				pm.type as 'payment_method_type',
				pm.name as 'payment_method_name',
				pm.acquirer as 'payment_method_acquirer',
				CASE
					WHEN pm.type IN ('CARD', 'CREDIT_CARD') AND p.metadata->>'$.paymentMethodOptions.card.captureMethod' = 'MANUAL' THEN
						(
							SELECT JSON_ARRAYAGG(
								JSON_OBJECT(
									'id', c.id,
									'paymentId', c.payment_id,
									'processorReferenceId', c.processor_reference_id,
									'status', c.status,
									'releaseRemainingAmount', c.release_remaining_amount,
									'currency', c.currency,
									'amount', c.amount,
									'createdAt', DATE_FORMAT(c.created_at, '%Y-%m-%dT%H:%i:%sZ'),
									'updatedAt', DATE_FORMAT(c.updated_at, '%Y-%m-%dT%H:%i:%sZ')
								)
							)
							FROM payment_captures c
							WHERE c.payment_id = p.uuid
						)
				    ELSE NULL
				END AS payment_captures
			FROM payments p
			LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid`

	// Query condition builder
	conditions, args := buildCondition(filter)
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	querySort := " ORDER BY p.created_at DESC"
	if filter.Sort != "" && filter.SortBy != "" {
		querySort = fmt.Sprintf(" ORDER BY p.%s %s", util.ConvertCamelToSnake(filter.SortBy), filter.Sort)
	}

	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PerPage, offset)
	query += querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get payments list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := `SELECT COUNT(p.uuid) as totalItems FROM payments p
		LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid`

	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ")
	}

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalItems, queryCount, args...)
		if err != nil {
			mu.Lock()
			totalItems = 0
			mu.Unlock()
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(filter.PerPage)))
	meta := commonModel.Meta{
		Page:       int64(filter.Page),
		PerPage:    int64(filter.PerPage),
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return &commonModel.PaginationResponse{
		Data: BuildRespData(data), Meta: meta,
	}, nil
}

func BuildRespData(
	data []*paymentModel.PaymentWithPaymentMethodDTO) []*paymentModel.Payment {
	respData := make([]*paymentModel.Payment, len(data))
	for i, paymentDTO := range data {
		var payment paymentModel.Payment
		payment.PaymentFromPaymentWithPaymentMethodDTO(paymentDTO)

		respData[i] = &payment
	}

	return respData
}

func buildCondition(filter *paymentModel.GetListFilterRequest) (conditions []string, args []interface{}) {
	if !filter.IncludeCardFundedPaymentSession {
		conditions = append(conditions, "p.type IN ('', 'MULTIPLE', 'SINGLE')")
	}

	if filter.MerchantID != "" {
		conditions = append(conditions, "p.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}

	if filter.UUID != "" {
		conditions = append(conditions, "p.uuid = ?")
		args = append(args, filter.UUID)
	}

	if filter.Status != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, strings.ToUpper(filter.Status))
	}

	if filter.ReferenceID != "" {
		conditions = append(conditions, "p.reference_id = ?")
		args = append(args, filter.ReferenceID)
	}

	if filter.PaymentMethod != "" {
		conditions = append(conditions, "pm.type = ?")
		args = append(args, strings.ToUpper(filter.PaymentMethod))
	}

	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "p.created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "p.created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}

	return conditions, args
}
