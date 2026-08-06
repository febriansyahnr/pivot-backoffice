package paymentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *PaymentRepository) GetVCCTerminalList(
	ctx context.Context,
	filter *paymentModel.GetVCCTerminalListFilterRequest,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetVCCTerminalList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		queryResult          = make([]*paymentModel.VccTerminalItem, 0)
		errG                 = new(errgroup.Group)
		totalRecord    int64 = 0
		sortBy               = ""
		whereStatement       = ""
	)
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.PerPage < 1 {
		filter.PerPage = 10
	}

	if filter.Sort == "" {
		filter.Sort = "ASC"
	}

	if filter.SortBy == "" {
		filter.SortBy = "chargeDate"
	}

	switch filter.SortBy {
	case "createdAt", "chargeDate":
		sortBy = "created_at"
	}

	offset := (filter.Page - 1) * filter.PerPage

	selectedFields := `att.uuid as charge_id, p.reference_id, p.amount, p.currency, p.status, p.created_at,
	metadata->>'$.virtualTerminal.batchId' as bulk_id,
	metadata->>'$.virtualTerminal.travelAgentName' as travel_agent,
	metadata->>'$.virtualTerminal.bookingId' as booking_id`

	queryTemplate := `
	SELECT %s 
	FROM payments p
	JOIN account_transactions att on att.reference_id = p.uuid 
	  AND att.type = 'PAYMENT' 
	  AND att.created_at BETWEEN ? AND DATE_ADD(?, INTERVAL 1 MINUTE)
	`

	whereClause, args := buildVCCTerminalListWhereClause(filter)
	if len(whereClause) > 0 {
		whereStatement = " WHERE " + strings.Join(whereClause, " AND ")
	}
	args = append([]any{filter.ChargeStartDate, filter.ChargeEndDate}, args...)

	errG.Go(func() error {
		query := fmt.Sprintf(queryTemplate, "COUNT(p.uuid)")
		query = query + whereStatement
		err := r.db.GetContext(ctx, &totalRecord, query, args...)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})

	errG.Go(func() error {
		sortStatement := fmt.Sprintf(" ORDER BY %s %s", sortBy, filter.Sort)
		limitStatement := " LIMIT ? OFFSET ?"
		args2 := append(args, filter.PerPage, offset)

		query := fmt.Sprintf(queryTemplate, selectedFields)
		query = query + whereStatement + sortStatement + limitStatement
		err := r.db.SelectContext(ctx, &queryResult, query, args2...)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})

	err := errG.Wait()
	if err != nil {
		r.logger.Error(ctx, "error when get vcc terminal list", logger.Error(err))
		return nil, err
	}

	// mapping amount
	for i, res := range queryResult {
		queryResult[i].Amount = paymentModel.Amount{
			Value:    res.ChargeAmount,
			Currency: res.ChargeCurrency,
		}
	}

	totalPages := int64(math.Ceil(float64(totalRecord) / float64(filter.PerPage)))
	result := &commonModel.PaginationResponse{
		Data: queryResult,
		Meta: commonModel.Meta{
			Page:       int64(filter.Page),
			PerPage:    int64(filter.PerPage),
			TotalItems: totalRecord,
			TotalPages: totalPages,
		},
	}

	return result, nil
}

func buildVCCTerminalListWhereClause(filter *paymentModel.GetVCCTerminalListFilterRequest) ([]string, []any) {
	var whereClause []string
	var args []any

	whereClause = append(whereClause, "p.type = ?")
	args = append(args, "VIRTUAL_TERMINAL")

	whereClause = append(whereClause, "p.merchant_id = ?")
	args = append(args, filter.MerchantID)

	if !filter.ChargeStartDate.IsZero() && !filter.ChargeEndDate.IsZero() {
		whereClause = append(whereClause, "p.created_at BETWEEN ? AND ?")
		args = append(args, filter.ChargeStartDate, filter.ChargeEndDate)
	}

	if filter.ChargeID != "" {
		whereClause = append(whereClause, "att.uuid = ?")
		args = append(args, filter.ChargeID)
	}

	if filter.ReferenceID != "" {
		whereClause = append(whereClause, "p.reference_id = ?")
		args = append(args, filter.ReferenceID)
	}

	if filter.Status != "" {
		whereClause = append(whereClause, "p.status = ?")
		args = append(args, filter.Status)
	}

	return whereClause, args
}
