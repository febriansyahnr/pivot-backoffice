package userLoggedInDeviceRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/userLoggedInDevice"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *UserLoggedInDeviceRepository) Create(ctx context.Context, data *userLoggedInDeviceModel.UserLoggedInDevice) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userLoggedInDevice/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_logged_in_devices")

	query := `
		INSERT INTO user_logged_in_devices (
		   uuid, user_id, device_identifier, status, additional_info, created_at, updated_at
		)
		VALUES (
		        :uuid, :user_id, :device_identifier, :status, :additional_info, :created_at, :updated_at
		)
	`

	affected, err := r.db.NamedExecContext(ctx, query, data)
	if err != nil {
		r.logger.Error(ctx, "error when inserting user_logged_in_devices", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting user_logged_in_devices", logger.Error(constant.ErrNoRowsAffected))
		return constant.ErrNoRowsAffected
	}

	return nil
}
