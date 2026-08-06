package transferRepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *transferRepository) GetList(ctx context.Context, req *transfer.GetTransferListRequest, page, perPage int64) ([]*transfer.Transfer, int64, error) {

	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/transfer/GetList")
	defer segment.End()

	var (
		whereClause []string
		whereParams []interface{}
		data        []*transfer.Transfer
		errG        = new(errgroup.Group)
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	queryCount := `SELECT
		COUNT(t.uuid) 
	FROM transfers AS t
	LEFT JOIN payments AS p ON p.uuid = t.reference_id` // Need to be reconfirmed

	query := `
		SELECT 
			t.uuid,
			t.merchant_id,
			t.recipient_id,
			t.reference_id,
			t.currency,
			CASE 
				WHEN t.recipient_id = '` + req.MerchantID + `' THEN t.amount
				ELSE -t.amount
			END as amount,
			CASE 
				WHEN t.recipient_id = '` + req.MerchantID + `' THEN '` + constant.TransferTypeIN + `'
				ELSE '` + constant.TransferTypeOUT + `'
			END as direction,
			t.status,
			t.remarks,
			t.transfer_type,
			t.created_at,
			t.updated_at,
			t.deleted_at
		FROM transfers AS t
		JOIN merchants AS sender ON sender.uuid = t.merchant_id
		JOIN merchants AS recipient ON recipient.uuid = t.recipient_id
		LEFT JOIN payments AS p ON t.reference_id = p.uuid` // Need to be reconfirmed

	whereClause, whereParams = r.GenerateMerchantClauseAndParam(ctx, req)

	whereClause = append(whereClause, "(t.created_at >= ? AND t.created_at <= ?)")
	whereParams = append(whereParams, req.StartDate, req.EndDate)
	if req.Status != "" {
		whereClause = append(whereClause, "t.status = ?")
		whereParams = append(whereParams, req.Status)
	}
	if req.ReferenceID != "" {
		whereClause = append(whereClause, "t.reference_id LIKE ?")
		whereParams = append(whereParams, "%"+req.ReferenceID+"%")
	}

	if req.UUID != "" || req.PaymentID != "" || req.PaymentReferenceID != "" {
		clauses := []string{}
		params := []interface{}{}

		if req.UUID != "" {
			clauses = append(clauses, "t.uuid LIKE ?")
			params = append(params, "%"+req.UUID+"%")
		}

		if req.PaymentID != "" {
			clauses = append(clauses, "p.uuid LIKE ?")
			params = append(params, "%"+req.PaymentID+"%")
		}

		if req.PaymentReferenceID != "" {
			clauses = append(clauses, "p.reference_id LIKE ?")
			params = append(params, "%"+req.PaymentReferenceID+"%")
		}

		whereClause = append(whereClause, fmt.Sprintf("(%s)", strings.Join(clauses, " OR ")))
		whereParams = append(whereParams, params...)
	}

	if len(whereClause) > 0 {
		query = query + ` WHERE ` + strings.Join(whereClause, " AND ")
		queryCount = queryCount + ` WHERE ` + strings.Join(whereClause, " AND ")
	}
	countWhereParams := whereParams

	if req.SortBy == transfer.SortColCreatedAt {
		query = query + ` ORDER BY t.created_at ` + req.SortOrder
	}
	if req.SortBy == transfer.SortColAmount {
		query = query + ` ORDER BY amount ` + req.SortOrder
	}
	if req.SortBy == transfer.SortColRecipient {
		query = query + ` ORDER BY recipient.name ` + req.SortOrder
	}
	if req.SortBy == transfer.SortColSender {
		query = query + ` ORDER BY sender.name ` + req.SortOrder
	}

	if req.SortBy == "" {
		query = query + ` ORDER BY t.created_at DESC`
	}
	query = query + ` LIMIT ? OFFSET ?`
	whereParams = append(whereParams, perPage)
	whereParams = append(whereParams, (page-1)*perPage)

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, whereParams...)
		if err != nil {
			if err == sql.ErrNoRows {
				r.logger.Error(ctx, "listing transfers not found", logger.Any("request", req))
				return nil
			} else {
				r.logger.Error(ctx, "error when get listing transfers", logger.Error(err), logger.Any("query", query), logger.Any("request", req))
				return err
			}
		}

		return nil
	})

	var total int64
	errG.Go(func() error {
		err := r.db.GetContext(ctx, &total, queryCount, countWhereParams...)
		if err != nil {
			r.logger.Error(ctx, "error when get total listing transfers", logger.Error(err), logger.Any("query", queryCount), logger.Any("request", req))
			return err
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

// GenerateMerchantClauseAndParam generates SQL clauses and parameters based on the provided transfer request.
// It constructs different SQL conditions depending on the ParentID and Type fields in the request.
// Because the ParentID field is optional, the function will generate different SQL conditions based on the Type field.
// If the ParentID field is provided, indicate that the data was requested by the parent merchant and want to see the transaction between sub-merchants and the parent.
func (r *transferRepository) GenerateMerchantClauseAndParam(ctx context.Context, req *transfer.GetTransferListRequest) (clauses []string, params []interface{}) {

	if req.ParentID != "" {
		if req.Type == "" {
			query, args, err := sqlx.In("(t.merchant_id IN (?) AND t.recipient_id IN (?))", []string{req.ParentID, req.MerchantID}, []interface{}{req.ParentID, req.MerchantID})
			if err != nil {
				r.logger.Error(ctx, "error when construct contain query", logger.Error(err), logger.Any("req", req))
				return clauses, params
			}

			clauses = append(clauses, query)
			params = append(params, args...)
		}

		if req.Type == constant.TransferTypeIN {
			clauses = append(clauses, "t.recipient_id = ? AND t.merchant_id = ?")
			params = append(params, req.MerchantID, req.ParentID)
		}

		if req.Type == constant.TransferTypeOUT {
			clauses = append(clauses, "t.merchant_id = ? AND t.recipient_id = ?")
			params = append(params, req.MerchantID, req.ParentID)
		}

		return clauses, params
	}

	if req.Type == "" {
		clauses = append(clauses, "(t.merchant_id = ? OR t.recipient_id = ?)")
		params = append(params, req.MerchantID, req.MerchantID)
	}

	if req.Type == constant.TransferTypeIN {
		clauses = append(clauses, "t.recipient_id = ?")
		params = append(params, req.MerchantID)
	}

	if req.Type == constant.TransferTypeOUT {
		clauses = append(clauses, "t.merchant_id = ? ")
		params = append(params, req.MerchantID)
	}

	return clauses, params
}
