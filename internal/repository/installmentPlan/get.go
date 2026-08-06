package installmentPlan

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *InstallmentPlanRepository) GetById(ctx context.Context, planId string) (*installmentPlanModel.InstallmentPlan, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/installmentPlan/GetById")
	defer segment.End()

	var installmentPlan installmentPlanModel.InstallmentPlan
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, installmentPlansTableName)
	query := `
		SELECT 
			uuid, 
			merchant_id, 
			acquirer, 
			settlement_type, 
			installment_type,
			payment_method, 
			title, 
			description, 
			tenor, 
			status, 
			metadata, 
			created_at, 
			updated_at, 
			deleted_at
		FROM ` + installmentPlansTableName + `
		WHERE 
			uuid = ? and deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &installmentPlan, query, planId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding installment plan", logger.Error(err), logger.Any("planId", planId))
		return nil, err
	}
	return &installmentPlan, err
}

func (r *InstallmentPlanRepository) List(ctx context.Context, req *installmentPlanModel.FilterRequest) ([]*installmentPlanModel.InstallmentPlan, int64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/installmentPlan/List")
	defer segment.End()

	var (
		list        = []*installmentPlanModel.InstallmentPlan{}
		whereClause []string
		whereArgs   []interface{}
		errG        = new(errgroup.Group)
		total       int64
	)
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, installmentPlansTableName)
	countQuery := `SELECT COUNT(1) FROM ` + installmentPlansTableName
	query := `
		SELECT uuid, 
			merchant_id, 
			acquirer, 
			settlement_type, 
			installment_type,
			payment_method, 
			title, 
			description, 
			tenor, 
			status, 
			metadata, 
			created_at, 
			updated_at, 
			deleted_at
		FROM ` + installmentPlansTableName

	whereClause = append(whereClause, "deleted_at IS NULL")
	if req.MerchantID != "" {
		whereClause = append(whereClause, "merchant_id = ?")
		whereArgs = append(whereArgs, req.MerchantID)
	}
	if req.Acquirer != "" {
		whereClause = append(whereClause, "acquirer = ?")
		whereArgs = append(whereArgs, req.Acquirer)
	}
	if req.SettlementType != "" {
		whereClause = append(whereClause, "settlement_type = ?")
		whereArgs = append(whereArgs, req.SettlementType)
	}
	if req.Status != "" {
		whereClause = append(whereClause, "status = ?")
		whereArgs = append(whereArgs, req.Status)
	}
	if req.PaymentMethod != "" {
		whereClause = append(whereClause, "payment_method = ?")
		whereArgs = append(whereArgs, req.PaymentMethod)
	}
	if req.Tenor > 0 {
		whereClause = append(whereClause, "tenor = ?")
		whereArgs = append(whereArgs, req.Tenor)
	}
	if req.MidID != "" {
		whereClause = append(whereClause, "metadata->>'$.card.midId' = ?")
		whereArgs = append(whereArgs, req.MidID)
	}
	if len(req.InstallmentIDs) > 0 {
		inQuery, paramArgs, err := sqlx.In("uuid IN (?)", req.InstallmentIDs)
		if err != nil {
			r.logger.Error(ctx, "error when formatting in query for installment ids", logger.Error(err), logger.Any("query", inQuery), logger.Any("req", req))
			return nil, 0, err
		}
		whereClause = append(whereClause, inQuery)
		whereArgs = append(whereArgs, paramArgs...)
	}
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
		countQuery += " WHERE " + strings.Join(whereClause, " AND ")
	}

	countQuery = r.db.Rebind(countQuery)
	countArgs := whereArgs
	errG.Go(func() error {
		err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
		if err != nil {
			r.logger.Error(ctx, "error when count installment plans", logger.Error(err), logger.Any("query", countQuery), logger.Any("req", req))
			return err
		}
		return nil
	})

	query += " ORDER BY created_at DESC"
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query += " LIMIT ? OFFSET ?"
		whereArgs = append(whereArgs, req.PageSize, offset)
	}
	query = r.db.Rebind(query)

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &list, query, whereArgs...)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			r.logger.Error(ctx, "error when finding installment plans", logger.Error(err), logger.Any("query", query), logger.Any("req", req))
			return err
		}
		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, 0, err
	}

	for _, plan := range list {
		if errMetadata := plan.LoadMetadata(); errMetadata != nil {
			r.logger.Error(ctx, "error when load metadata", logger.Error(errMetadata), logger.Any("plan", plan))
			plan.PlanMetadata = &installmentPlanModel.InstallmentPlanMetadata{
				Card: &installmentPlanModel.CardInstallmentMetadata{},
			} // Avoid nil panic in caller functions
		}
	}

	return list, total, nil
}
