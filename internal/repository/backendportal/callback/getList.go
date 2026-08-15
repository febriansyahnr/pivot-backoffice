package callbackRepository

import (
	"context"
	"fmt"
	"math"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pdk/v2/logger"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *CallbackRepository) GetList(ctx context.Context, filter *callbackModel.GetListCallbackFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "callbacks")

	var (
		mu   sync.Mutex
		data = make([]callbackModel.Callback, 0)
		errG = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT uuid, callback_master_id, merchant_id, url, description, created_at, updated_at FROM callbacks WHERE 1 = 1`
	queryCondition := ""
	if filter.MerchantID != nil {
		queryCondition += fmt.Sprintf(" AND merchant_id = '%s'", *filter.MerchantID)
	}
	querySort := " ORDER BY created_at DESC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += queryCondition + querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		rows, err := r.db.QueryContext(ctx, query)
		if err != nil {
			r.logger.Error(ctx, "error when get list callback", logger.Error(err))
			return err
		}
		defer rows.Close()

		// Iterate over the result set
		for rows.Next() {
			var callback callbackModel.Callback

			// Scan the row data into variables
			if err = rows.Scan(
				&callback.UUID,
				&callback.CallbackMasterID,
				&callback.MerchantID,
				&callback.URL,
				&callback.Description,
				&callback.CreatedAt,
				&callback.UpdatedAt,
			); err != nil {
				r.logger.Error(ctx, "error when scan list callback", logger.Error(err))
				return err
			}

			mu.Lock()
			data = append(data, callback)
			mu.Unlock()
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := "SELECT COUNT(uuid) as totalItems FROM callbacks WHERE 1 = 1"
	queryCount += queryCondition

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalItems, queryCount)
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
