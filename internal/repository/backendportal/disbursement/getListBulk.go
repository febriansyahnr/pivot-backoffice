package disbursementRepository

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *DisbursementRepository) GetListBulk(ctx context.Context, filter *disbursementModel.GetBulkDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/disbursement/GetListBulk")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	var (
		conditions []string
		args       []interface{}
		mu         sync.Mutex
		data       = make([]*disbursementModel.BulkDisbursementWithAggregate, 0)
		errG       = new(errgroup.Group)

		BaseQuery = `SELECT 
			bd.uuid, 
			bd.merchant_id,
			bd.file,
			bd.file_failed,
			bd.status,
			COALESCE(c.name, 'System') as created_by,
			bd.created_at, 
			bd.updated_at,
			(SELECT COALESCE(SUM(amount),0) FROM disbursements WHERE bulk_id = bd.uuid) as 'total_amount',
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid) as 'total_trx',	
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid AND status = '` + constant.DisbursementStatusApproved + `') as 'total_approved',
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid AND status = '` + constant.DisbursementStatusRejected + `') as 'total_rejected',
			(
			    SELECT 
					COUNT(d.uuid)
				FROM disbursements d
				INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
				WHERE d.bulk_id = bd.uuid AND d.status = '` + constant.DisbursementStatusApproved + `' AND t.status = '` + constant.StatusSuccess + `'
			) as 'total_success',
			(
			    SELECT 
					COUNT(d.uuid)
				FROM disbursements d
				INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
				WHERE d.bulk_id = bd.uuid AND d.status = '` + constant.DisbursementStatusApproved + `' AND t.status = '` + constant.StatusFailed + `'
			) as 'total_failed',
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid AND reason_type = '` + constant.DisbursementReasonTypeCancelled + `') as 'total_cancelled',
			(SELECT COALESCE(COUNT(d.uuid),0) FROM disbursements d
				JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
				WHERE d.bulk_id = bd.uuid AND d.status = '` + constant.DisbursementStatusApproved + `' AND t.status = '` + constant.StatusPending + `') as 'total_pending'
		FROM bulk_disbursements bd
		LEFT JOIN users c ON c.uuid = bd.created_by`

		RetryListQuery = `SELECT 
			bd.uuid, 
			bd.merchant_id,
			bd.file,
			bd.file_failed,
			bd.file_rejected,
			bd.status,
			COALESCE(c.name, 'System') as created_by,
			bd.created_at, 
			bd.updated_at,
			(SELECT COALESCE(SUM(amount),0) FROM disbursements 
				WHERE bulk_id = bd.uuid 
				AND status = '` + constant.DisbursementStatusApproved + `'
				AND reason_type = '` + constant.DisbursementReasonTypeInsufficientBalance + `'
			) as 'total_amount',
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements
				WHERE bulk_id = bd.uuid 
				AND status = '` + constant.DisbursementStatusApproved + `'
				AND reason_type = '` + constant.DisbursementReasonTypeInsufficientBalance + `'
			) as 'total_trx',	
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid AND status = '` + constant.DisbursementStatusApproved + `') as 'total_approved',
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid AND status = '` + constant.DisbursementStatusRejected + `') as 'total_rejected',
			(
				SELECT 
					COUNT(d.uuid)
					FROM disbursements d
				INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
				WHERE d.bulk_id = bd.uuid AND d.status = '` + constant.DisbursementStatusApproved + `' AND t.status = '` + constant.StatusSuccess + `'
			) as 'total_success',
			(
		    	SELECT 
					COUNT(d.uuid)
				FROM disbursements d
				INNER JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
				WHERE d.bulk_id = bd.uuid AND d.status = '` + constant.DisbursementStatusApproved + `' AND t.status = '` + constant.StatusFailed + `'
			) as 'total_failed',
			(SELECT COALESCE(COUNT(uuid),0) FROM disbursements WHERE bulk_id = bd.uuid AND reason_type = '` + constant.DisbursementReasonTypeCancelled + `') as 'total_cancelled',
			(SELECT COALESCE(COUNT(d.uuid),0) FROM disbursements d
				JOIN account_transactions t ON t.reference_id = d.uuid AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
				WHERE d.bulk_id = bd.uuid AND d.status = '` + constant.DisbursementStatusApproved + `' AND t.status = '` + constant.StatusPending + `') as 'total_pending'
		FROM bulk_disbursements bd
		LEFT JOIN users c ON c.uuid = bd.created_by`
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := BaseQuery
	if filter.Status == constant.BulkDisbursementStatusPending {
		query = RetryListQuery
	}

	// Query condition builder
	if filter.MerchantID != "" {
		conditions = append(conditions, "bd.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "bd.status LIKE ?")
		args = append(args, "%"+filter.Status+"%")
	}
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "bd.created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "bd.created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}
	if filter.ReferenceID != "" {
		conditions = append(conditions, "bd.uuid = ?")
		args = append(args, filter.ReferenceID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	querySort := " ORDER BY bd.created_at DESC"
	if filter.Sort != "" && filter.SortBy != "" {
		querySort = fmt.Sprintf(" ORDER BY bd.%s %s", util.ConvertCamelToSnake(filter.SortBy), filter.Sort)
	}

	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.pdkLogger.Error(ctx, "error when get bulk disbursement list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := "SELECT COUNT(uuid) as totalItems FROM bulk_disbursements bd"
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

	totalPages := int64(math.Ceil(float64(totalItems) / float64(perPage)))
	meta := commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}
