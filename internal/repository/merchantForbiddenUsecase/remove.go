package merchantForbiddenUsecase

import (
	"context"
	"errors"
	"time"

	merchantForbiddenUseCaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantForbiddenUsecaseRepository) RemoveForbiddenUsecase(ctx context.Context,
	req *merchantForbiddenUseCaseModel.MerchantForbiddenUseCase) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchantforbiddenusecase/RemoveForbiddenUsecase")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE ` + tableName + ` 
		SET deleted_at = ?
		WHERE merchant_id = ? AND use_case = ?`

	affected, err := r.db.ExecContext(ctx, query, time.Now().UTC(), req.MerchantID, req.UseCase)
	if err != nil {
		r.logger.Error(ctx, "error when remove forbidden use case", logger.Error(err), logger.Any("request", req))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when remove forbidden use case", logger.Error(err), logger.Any("request", req))
		return err
	}

	return nil

}
