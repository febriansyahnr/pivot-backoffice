package merchantForbiddenUsecase

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	merchantForbiddenUseCaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantForbiddenUsecase"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantForbiddenUsecaseRepository) RegisterForbiddenUsecase(ctx context.Context,
	req *merchantForbiddenUseCaseModel.MerchantForbiddenUseCase) (*merchantForbiddenUseCaseModel.MerchantForbiddenUseCase, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchantforbiddenusecase/RegisterForbiddenUsecase")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO ` + tableName + ` (
			uuid, merchant_id, use_case, created_at, 
			updated_at, deleted_at
		) VALUES (
            :uuid, :merchant_id, :use_case, :created_at, 
            :updated_at, :deleted_at
        )`

	ok, err := r.db.NamedExecContext(ctx, query, req)
	if err != nil {
		r.logger.Error(ctx, "error when register forbidden use case", logger.Error(err), logger.Any("request", req))
		return nil, err
	}

	if !ok {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when register forbidden use case", logger.Error(err), logger.Any("request", req))
		return nil, err
	}

	return req, nil
}
