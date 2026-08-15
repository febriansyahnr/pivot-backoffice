package userLoggedInDeviceRepository

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *UserLoggedInDeviceRepository) SetRememberDevice(ctx context.Context, userId, deviceIdentifier, data string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userLoggedInDevice/SetRememberDevice")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_logged_in_devices")

	// First update
	query1 := `UPDATE user_logged_in_devices SET additional_info = NULL WHERE user_id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query1, userId)
	if err != nil {
		r.logger.Error(ctx, "error when updating remember device (step 1)", logger.Error(err))
		return err
	}

	// Second update
	query2 := `UPDATE user_logged_in_devices SET additional_info = ? WHERE user_id = ? AND device_identifier = ?`
	_, err = r.db.ExecContext(ctx, query2, data, userId, deviceIdentifier)
	if err != nil {
		r.logger.Error(ctx, "error when updating remember device (step 2)", logger.Error(err))
		return err
	}
	return nil
}
