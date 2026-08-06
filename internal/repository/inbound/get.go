package inboundRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

const (
	selectStr = `id, ip, client, method, url, headers, body, status_code, response_time_ms, response_body, metadata, snap_compatibility, created_at, updated_at,
    	COALESCE(reference_id, '') as reference_id, COALESCE(trace_id, '') as trace_id, COALESCE(origin_id, '') as origin_id`
)

func (r *repository) GetList(
	ctx context.Context,
	filter *inboundModel.GetInboundFilterRequest,
) (*commonModel.PaginationResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/repository/inbound/GetList")
	defer span.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "inbound")

	var (
		conditions []string
		args       []interface{}
		mu         sync.Mutex
		data       = make([]*inboundModel.Inbound, 0)
		errG       = new(errgroup.Group)
	)

	if filter.Page < 1 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PerPage

	// QUERY CONDITION
	query := `SELECT ` + selectStr + ` FROM inbound`

	// Query condition builder
	if filter.MerchantID != "" {
		conditions = append(conditions, "reference_id = ?")
		args = append(args, filter.MerchantID)
	}
	if filter.OriginID != "" {
		conditions = append(conditions, "origin_id = ?")
		args = append(args, filter.OriginID)
	}
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}
	if filter.Status != "" {
		switch filter.Status {
		case "SUCCESS":
			conditions = append(conditions, "status_code >= 200 AND status_code < 300")
		case "REDIRECT":
			conditions = append(conditions, "status_code >= 300 AND status_code < 400")
		case "FAILED":
			conditions = append(conditions, "status_code >= 400 AND status_code < 500")
		case "ERROR":
			conditions = append(conditions, "status_code >= 500")
		}
	}
	if filter.Method != "" {
		conditions = append(conditions, "method = ?")
		args = append(args, filter.Method)
	}
	if filter.Product != "" {
		conditions = append(conditions, "client->>'$.feature' = ?")
		args = append(args, filter.Product)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	querySort := " ORDER BY created_at DESC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PerPage, offset)
	query += querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := "SELECT COUNT(id) as totalItems FROM inbound"
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
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	responses := make([]*inboundModel.InboundResponse, len(data))
	for i, datum := range data {
		responses[i] = datum.ToResponse()
	}

	return &commonModel.PaginationResponse{
		Data: responses,
		Meta: meta,
	}, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*inboundModel.InboundResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/repository/inbound/GetByID")
	defer span.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "inbound")

	var data inboundModel.Inbound

	query := `SELECT 
			` + selectStr + `
		FROM inbound
		WHERE id = ?`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return data.ToResponse(), nil
}
