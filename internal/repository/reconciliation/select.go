package reconciliation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

// GetAll implements repository.IReconciliationRepository.
func (r *ReconciliationRepository) GetAll(ctx context.Context, filter *reconciliation.ReconciliationFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		mu      sync.Mutex
		data          = make([]*reconciliation.Reconciliation, 0)
		errG          = new(errgroup.Group)
		page    int64 = filter.Page
		perPage int64 = filter.PerPage
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	querySelect := `
	SELECT
			uuid, original_name, file_path, result_file_path, status, transaction_type, reasons, created_by, created_at, updated_at
	FROM %s
	%s
	ORDER BY created_at DESC
	%s
	`

	queryWhere := filter.Query()

	// Query builder
	queryLimitOffset := fmt.Sprintf("LIMIT %d OFFSET %d", perPage, offset)
	querySelect = fmt.Sprintf(querySelect, tableName, queryWhere, queryLimitOffset)

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, querySelect)
		if err != nil {
			r.logger.Error(ctx, "error when get recon list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := fmt.Sprintf(`SELECT count(A.uuid) as totalItems FROM( %s ) AS A`, querySelect)

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

// GetByUUID implements repository.IReconciliationRepository.
func (r *ReconciliationRepository) GetByUUID(ctx context.Context, uuid string) (*reconciliation.Reconciliation, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/reconciliation/GetByUUID")
	defer segment.End()

	var recon reconciliation.Reconciliation

	query := fmt.Sprintf(`
		SELECT
			uuid, file_path, result_file_path, status, transaction_type, reasons, created_by, created_at, updated_at
		FROM %s
		WHERE uuid = ?
	`, tableName)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	if err := r.db.GetContext(ctx, &recon, query, uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "recon not found", logger.String("uuid", uuid))
			return nil, constant.ErrDataNotFound
		}

		r.logger.Error(ctx, "error when finding recon", logger.Error(err))
		return &recon, err
	}

	return &recon, nil
}
