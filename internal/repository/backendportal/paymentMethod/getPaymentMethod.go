package paymentMethodRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PaymentMethodRepository) GetPaymentMethodById(
	ctx context.Context, paymentMethodId string) (*paymentModel.PaymentMethod, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/GetPaymentMethodByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var paymentMethod paymentModel.PaymentMethod
	query := `SELECT ` + SelectStrPaymentMethod + ` FROM payment_methods WHERE uuid = ?`

	if err := r.db.GetContext(ctx, &paymentMethod, query, paymentMethodId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, fmt.Sprintf("get payment method is not found by id=%s", paymentMethodId), logger.Error(err))
			return nil, constant.ErrPaymentMethodNotFound
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payment_methods by id=%s", paymentMethodId), logger.Error(err))
		return nil, err
	}

	paymentMethod.UnmarshalConfigObj()

	return &paymentMethod, nil
}

func (r *PaymentMethodRepository) GetActivePaymentMethodByRequest(ctx context.Context, request *paymentModel.GetPaymentMethodFilterRequest) (*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/GetActivePaymentMethodByRequest")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var paymentMethod paymentModel.PaymentMethodWithPivot
	query := `SELECT ` + SelectPaymentMethodWithPivotStr + `
				FROM payment_method_merchant pmm
				LEFT JOIN payment_methods p ON pmm.payment_method_id = p.uuid
				WHERE pmm.merchant_id = ? AND p.category = ? AND p.type = ? AND pmm.is_active = true AND pmm.activation_status = 'APPROVED'`

	if request.Acquirer != "" {
		query = query + fmt.Sprintf(" AND p.acquirer = '%s'", request.Acquirer)
	}

	if err := r.db.GetContext(ctx, &paymentMethod, query, request.MerchantID, request.MerchantID, request.Category, request.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "get payment method is not found", logger.Any("request", request), logger.Error(err))
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding payment method by request", logger.Any("request", request), logger.Error(err))
		return nil, err
	}

	paymentMethod.MerchantConfigObj = &paymentModel.PaymentMethodMerchantConfigObject{}
	if paymentMethod.MerchantConfig.Valid {
		_ = json.Unmarshal(paymentMethod.MerchantConfig.JSONText, &paymentMethod.MerchantConfigObj)
	}

	if paymentMethod.RequiredDocuments.Valid {
		_ = json.Unmarshal(paymentMethod.RequiredDocuments.JSONText, &paymentMethod.RequiredDocumentObjects)
	}

	paymentMethod.UnmarshalConfigObj()

	return &paymentMethod, nil
}

func (r *PaymentMethodRepository) GetAllPaymentMethodByCategory(
	ctx context.Context, category string) ([]*paymentModel.PaymentMethod, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/GetAllPaymentMethodByCategory")
	defer segment.End()

	var paymentMethods []*paymentModel.PaymentMethod

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `SELECT ` + SelectStrPaymentMethod + ` FROM payment_methods WHERE category = ?`

	if err := r.db.SelectContext(ctx, &paymentMethods, query, category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, fmt.Sprintf("get payment method is not found by category=%s", category), logger.Error(err))
			return nil, constant.ErrPaymentMethodNotFound
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payment_methods by category=%s", category), logger.Error(err))
		return nil, err
	}

	return paymentMethods, nil
}

func (r *PaymentMethodRepository) GetPaymentMethodByType(
	ctx context.Context, tipe string) ([]*paymentModel.PaymentMethod, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentMethodByCategory")
	defer segment.End()

	var paymentMethods []*paymentModel.PaymentMethod

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `SELECT ` + SelectStrPaymentMethod + ` FROM payment_methods WHERE type = ?`

	if err := r.db.SelectContext(ctx, &paymentMethods, query, tipe); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, fmt.Sprintf("get payment method is not found by type=%s", tipe), logger.Error(err))
			return nil, constant.ErrPaymentMethodNotFound
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payment_methods by type=%s", tipe), logger.Error(err))
		return nil, err
	}

	return paymentMethods, nil
}

func (r *PaymentMethodRepository) GetPaymentMethodByCategoryTypeAndAcquirer(ctx context.Context, category, typ, acquirer string) (*paymentModel.PaymentMethod, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/paymentMethod/GetPaymentMethodByCategoryAndFilters")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
		` + SelectStrPaymentMethod + `
	FROM payment_methods WHERE category = ? AND type = ? AND acquirer = ? AND deleted_at IS NULL;`

	result := paymentModel.PaymentMethod{}
	if err := r.db.GetContext(ctx, &result, rawQuery, category, typ, acquirer); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	result.UnmarshalConfigObj()
	return &result, nil
}
