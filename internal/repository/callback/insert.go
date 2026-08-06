package callbackRepository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Create a Callback Master
func (r *CallbackRepository) CreateCallbackMaster(
	ctx context.Context,
	callbackMaster *callbackModel.CallbackMaster,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/CreateCallbackMaster")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callback_masters")

	// Force value createdAt & updatedAt
	callbackMaster.CreatedAt = time.Now().UTC()
	callbackMaster.UpdatedAt = time.Now().UTC()

	query := `
		INSERT INTO callback_masters (uuid, name, description, created_at, updated_at)
		VALUES (:uuid, :name, :description, :created_at, :updated_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, callbackMaster)
	if err != nil {
		r.logger.Error(ctx, "error when inserting callback masters", logger.Error(err))
		return err
	}

	if !affected {
		err = errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting callback masters", logger.Error(err))
		return err
	}

	return nil
}

// Create a Callback
func (r *CallbackRepository) CreateCallback(
	ctx context.Context,
	callback *callbackModel.Callback,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/CreateCallback")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callbacks")

	// Force value createdAt & updatedAt
	callback.CreatedAt = time.Now().UTC()
	callback.UpdatedAt = time.Now().UTC()

	query := `
		INSERT INTO callbacks (uuid, callback_master_id, merchant_id, base_url, url, description, created_at, updated_at)
		VALUES (:uuid, :callback_master_id, :merchant_id, :base_url, :url, :description, :created_at, :updated_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, callback)
	if err != nil {
		r.logger.Error(ctx, "error when inserting callbacks", logger.Error(err))
		return err
	}

	if !affected {
		err = errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting callbacks", logger.Error(err))
		return err
	}

	return nil
}

// Create a Callback logs
func (r *CallbackRepository) CreateCallbackLog(ctx context.Context, callbackLog *callbackModel.CallbackLog) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/CreateCallbackLog")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callback_logs")

	query := `
		INSERT INTO callback_logs (uuid, callback_id, event, request, response, status, reference_id, created_at, updated_at, metadata, retry)
		VALUES (:uuid, :callback_id, :event, :request, :response, :status, :reference_id, :created_at, :updated_at, :metadata, :retry)`

	if callbackLog.Metadata != nil {
		callbackLog.RawMetadata.Valid = true
		callbackLog.RawMetadata.JSONText, _ = json.Marshal(callbackLog.Metadata)
	}

	affected, err := r.db.NamedExecContext(ctx, query, callbackLog)
	if err != nil {
		r.logger.Error(ctx, "error when inserting callback logs", logger.Error(err))
		return err
	}

	if !affected {
		err = errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting callback logs", logger.Error(err))
		return err
	}

	return nil
}
