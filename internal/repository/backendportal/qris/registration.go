package qris

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
)

func (r *qrisRepository) InitRegistration(ctx context.Context, data *qris.Registration) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/InitRegistration")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	rawQuery := `INSERT INTO 
		qr_registrations(
			id, external_id, acquirer, acquirer_merchant_id, merchant_type, acquirer_parent_merchant_id, merchant_name, 
			merchant_short_name, address, status, created_at, created_by, updated_at, business_info, business_document
		) VALUES(
		 	:id, :external_id, :acquirer, :acquirer_merchant_id, :merchant_type, :acquirer_parent_merchant_id, :merchant_name, 
			:merchant_short_name, :address, :status, :created_at, :created_by, :updated_at, :business_info, :business_document
		);`
	_, err := r.db.NamedExecContext(ctx, rawQuery, data)
	return err
}

func (r *qrisRepository) UpdateUploadedDocument(ctx context.Context, id string, data *qris.UpdateDocument) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/UpdateUploadedDocument")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	var (
		rawQuery string
		args     []interface{}
	)

	mediaRaw, _ := json.Marshal(data.Media)

	switch data.Type {
	case "NationalIdentityCard", "BusinessLicense", "TaxIdentification", "BusinessRegistration":

		args = []interface{}{data.Number, string(mediaRaw), id}
		rawQuery = fmt.Sprintf(`UPDATE qr_registrations 
			SET business_info = JSON_SET(business_info, '$.%s', ?),
				business_info = JSON_SET(business_info, '$.%s', CAST(? AS JSON)), updated_at = NOW()
			WHERE id = ?;`, documents[data.Type][0], documents[data.Type][1])

	case "CertificateIncorporation", "CertificateNo40", "CertificateLastAmendment", "CertificateDeedAmendment", "CertificateAmendmentAct", "CertificateEstablishment", "CertificateTaxRegistration", "BusinessEnvironmentPhoto":

		args = []interface{}{string(mediaRaw), id}
		rawQuery = fmt.Sprintf(`UPDATE qr_registrations 
			SET 
				business_document = JSON_SET(business_document, '$.%s', CAST(? AS JSON)), updated_at = NOW()
			WHERE id = ?;`, documents[data.Type][0])

	default:
		return errors.New("document type not found")
	}

	_, err := r.db.ExecContext(ctx, rawQuery, args...)
	return err
}

func (r *qrisRepository) UpdateRegistrationStatus(ctx context.Context, id, status string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/UpdateRegistrationStatus")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	rawQuery := `UPDATE qr_registrations SET status = ?, updated_at = NOW() WHERE id = ?;`

	_, err := r.db.ExecContext(ctx, rawQuery, status, id)
	return err
}

func (r *qrisRepository) FindQrRegistrationForValidationById(ctx context.Context, id string) (result *qris.Registration, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/FindQrRegistrationForValidationById")
	defer segment.End()

	ctx, result = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName), &qris.Registration{}

	rawQuery := `SELECT id, status, created_at, updated_at FROM qr_registrations WHERE id = ?;`

	if err = r.db.GetContext(ctx, result, rawQuery, id); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

func (r *qrisRepository) UpdateCallbackQrRegistration(ctx context.Context, id string, data *qris.RegistrationCallback) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/UpdateCallbackQrRegistration")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	callbackDetailRaw, _ := json.Marshal(&qris.CallbackDetail{
		ApplymentCode: data.ApplymentCode,
		ResultCode:    data.ResultCode,
		AuditDetail:   data.AuditDetail,
	})

	rawQuery := `UPDATE 
		qr_registrations
	SET
		status = ?, acquirer_merchant_id = ?, callback_detail = ?, callback_datetime = ?, updated_at = NOW()
	WHERE id = ?;`

	_, err := r.db.ExecContext(
		ctx, rawQuery, data.Status, data.MerchantId, types.JSONText(callbackDetailRaw), data.Datetime, id,
	)
	return err
}

func (r *qrisRepository) UpdateQrRegistration(ctx context.Context, id string, acquirerMerchantId string, acquirerTerminalId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/UpdateQrRegistration")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	rawQuery := `UPDATE qr_registrations SET acquirer_merchant_id = ?, acquirer_terminal_id = ?, status = ?, updated_at = ? WHERE id = ?;`

	_, err := r.db.ExecContext(ctx, rawQuery, acquirerMerchantId, acquirerTerminalId, constant.SuccessReg, time.Now().UTC(), id)
	return err
}
