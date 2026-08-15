package userLoggedInDeviceRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/userLoggedInDevice"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *UserLoggedInDeviceRepository) GetAllByUserID(ctx context.Context, userID string) ([]*userLoggedInDeviceModel.UserLoggedInDevice, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userLoggedInDevice/GetAllByUserID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_logged_in_devices")

	var userLoggedInDevices []*userLoggedInDeviceModel.UserLoggedInDevice

	query := `SELECT 
			uuid, user_id, device_identifier, status, additional_info, created_at, updated_at 
		FROM user_logged_in_devices
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC`

	if err := r.db.SelectContext(ctx, &userLoggedInDevices, query, userID); err != nil {
		r.logger.Error(ctx, "error when get all user_logged_in_devices by userID", logger.Error(err))
		return nil, err
	}

	return userLoggedInDevices, nil
}

func (r *UserLoggedInDeviceRepository) FindByUserAndDevice(ctx context.Context, userID, deviceIdentifier string) (*userLoggedInDeviceModel.UserLoggedInDevice, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/userLoggedInDevice/FindByUserAndDevice")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "user_logged_in_devices")

	var userLoggedInDevices userLoggedInDeviceModel.UserLoggedInDevice

	query := `SELECT 
			uuid, user_id, device_identifier, status, additional_info, created_at, updated_at 
		FROM user_logged_in_devices
		WHERE user_id = ? AND device_identifier = ? AND deleted_at IS NULL
		LIMIT 1`

	if err := r.db.GetContext(ctx, &userLoggedInDevices, query, userID, deviceIdentifier); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "user logged in device not found", logger.String("userId", userID),
				logger.String("deviceIdentifier", deviceIdentifier))
			return nil, nil
		}

		r.logger.Error(ctx, "error when find user_logged_in_devices by userID and deviceIdentifier",
			logger.String("userId", userID), logger.String("deviceIdentifier", deviceIdentifier), logger.Error(err))
		return nil, err
	}

	return &userLoggedInDevices, nil
}
