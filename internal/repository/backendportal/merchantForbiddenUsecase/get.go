package merchantForbiddenUsecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	merchantForbiddenUseCaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantForbiddenUsecase"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantForbiddenUsecaseRepository) GetForbiddenUsecase(ctx context.Context, req *merchantForbiddenUseCaseModel.GetMerchantForbiddenUseCaseRequest) ([]*merchantForbiddenUseCaseModel.MerchantForbiddenUseCase, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchantforbiddenusecase/GetForbiddenUsecase")
	defer segment.End()

	model := []*merchantForbiddenUseCaseModel.MerchantForbiddenUseCase{}
	query := `SELECT 
				uuid, 
				merchant_id, 
				use_case, 
				created_at, 
				updated_at, 
				deleted_at 
			FROM ` + tableName + ` 
			WHERE merchant_id = ? AND use_case = ? AND deleted_at IS NULL`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	if err := r.db.SelectContext(ctx, &model, query, req.MerchantID, req.UseCase); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "unable to find merchant forbidden usecase", logger.Any("request", req))
			return model, nil
		}

		r.logger.Error(ctx, "error when finding merchant forbidden usecase", logger.Error(err), logger.Any("request", req))
		return nil, err
	}

	return model, nil
}
