package callbackRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *CallbackRepository) GetCallbackLogList(ctx context.Context, filter *callbackModel.GetListCallbackLogFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetListCallbackLogFilterRequest")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, TableCallbackLog)

	var (
		mu         sync.Mutex
		conditions []string
		args       []interface{}
		data       = make([]callbackModel.CallbackLogWithMaster, 0)
		errG       = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT
			cl.uuid,
			c.merchant_id,
			cl.callback_id,
			cm.name as 'type',
			cl.event,
			c.base_url,
			c.url,
			cl.request,
			cl.response,
			cl.status,
			cl.retry,
			cl.reference_id,
			cl.created_at,
			cl.updated_at
		FROM callback_logs cl
		LEFT JOIN callbacks c ON cl.callback_id = c.uuid
		LEFT JOIN callback_masters cm ON c.callback_master_id = cm.uuid`

	if filter.MerchantID != "" {
		conditions = append(conditions, "c.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}
	if filter.StartUpdatedAt != nil && filter.EndUpdatedAt != nil {
		conditions = append(conditions, "cl.updated_at > ?")
		args = append(args, filter.StartUpdatedAt)
		conditions = append(conditions, "cl.updated_at < ?")
		args = append(args, filter.EndUpdatedAt)
	}
	if filter.Type != "" {
		conditions = append(conditions, "cm.name = ?")
		args = append(args, filter.Type)
	}
	if filter.Event != "" {
		conditions = append(conditions, "cl.event = ?")
		args = append(args, filter.Event)
	}
	if filter.Status != "" {
		conditions = append(conditions, "cl.status = ?")
		args = append(args, filter.Status)
	}
	// Handle keyword search (searches both UUID and reference_id with OR logic)
	if filter.Keyword != "" {
		conditions = append(conditions, "(cl.uuid LIKE ? OR cl.reference_id LIKE ?)")
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	querySort := " ORDER BY cl.updated_at DESC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when callback log list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := `SELECT count(cl.uuid) as totalItems
		FROM callback_logs cl
		LEFT JOIN callbacks c ON cl.callback_id = c.uuid
		LEFT JOIN callback_masters cm ON c.callback_master_id = cm.uuid`

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

func (r *CallbackRepository) FindMerchantCallbackHistory(ctx context.Context, filter *callbackModel.GetListCallbackLogFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/callback/FindMerchantCallbackHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callbacks,callback_masters,callback_logs")

	mainQuery := `SELECT 
		%s
	FROM callbacks c
	JOIN callback_masters cm ON cm.uuid = c.callback_master_id
	JOIN callback_logs cl ON cl.callback_id = c.uuid 
		AND cl.updated_at > ? AND cl.updated_at < ?
		-- Partition pruning by adding created_at with the expectation that logs will not be resend after 3 months.
		AND cl.created_at >= DATE_SUB(?, INTERVAL 3 MONTH) AND cl.created_at <= ?
	WHERE 
		c.merchant_id = ?`
	aggrColumns := "COUNT(cl.uuid)"
	rowColumns := "cl.uuid, c.merchant_id, cl.callback_id, cm.name as 'type', cl.event, c.base_url, c.url, cl.request, cl.response, cl.status, cl.retry, cl.reference_id, cl.created_at, cl.updated_at"

	conditions := make([]string, 0, 4)
	args := []any{filter.StartUpdatedAt, filter.EndUpdatedAt, filter.StartUpdatedAt, filter.EndUpdatedAt, filter.MerchantID}

	if filter.Type != "" {
		args = append(args, filter.Type)
		conditions = append(conditions, "cm.name = ?")
	}
	if filter.Event != "" {
		args = append(args, filter.Event)
		conditions = append(conditions, "cl.event = ?")
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, "cl.status = ?")
	}
	if filter.Keyword != "" {
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
		conditions = append(conditions, "(cl.uuid LIKE ? OR cl.reference_id LIKE ?)")
	}

	// Merging query conditions
	if len(conditions) > 0 {
		mainQuery += " AND " + strings.Join(conditions, " AND ")
	}

	totalItems := int64(0)
	group, ctx := errgroup.WithContext(ctx)
	callbackHistory := []callbackModel.CallbackLogWithMaster{}

	// Get a list of callback logs
	group.Go(func() error {
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * perPage

		rawQuery := fmt.Sprintf(mainQuery+" ORDER BY cl.updated_at DESC LIMIT %d OFFSET %d", rowColumns, perPage, offset)

		if err := r.db.SelectContext(ctx, &callbackHistory, rawQuery, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("list callback history: %v", err)
		}
		return nil
	})

	// Calculate the count of data based on conditions
	group.Go(func() error {
		rawQuery := fmt.Sprintf(mainQuery, aggrColumns)

		if err := r.db.GetContext(ctx, &totalItems, rawQuery, args...); err != nil {
			return fmt.Errorf("calculate total items: %v", err)
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		r.logger.Error(ctx, "Failed while find merchant callback history", logger.Error(err))
		return nil, err
	}

	return &commonModel.PaginationResponse{
		Data: callbackHistory,
		Meta: commonModel.Meta{
			Page:       page,
			PerPage:    perPage,
			TotalItems: totalItems,
			TotalPages: int64(math.Ceil(float64(totalItems) / float64(perPage))),
		},
	}, nil
}
