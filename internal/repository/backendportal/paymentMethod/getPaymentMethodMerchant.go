package paymentMethodRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (r *PaymentMethodRepository) GetListPaymentMethodMerchant(
	ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) ([]*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/GetListPaymentMethodMerchant")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		conditions           []string
		args                 []interface{}
		derivedID, isDerived = ctx.Value(constant.CtxDerivedMerchantID).(string)
	)
	data := make([]*paymentModel.PaymentMethodWithPivot, 0)

	// QUERY CONDITION
	query := `SELECT 
    	` + SelectPaymentMethodWithPivotStr + `
	FROM payment_methods p
	LEFT JOIN payment_method_merchant pmm 
		ON p.uuid = pmm.payment_method_id `

	// force append merchant ID
	args = append(args, filter.MerchantID)
	if filter.MerchantID != "" {
		query = query + " AND pmm.merchant_id = ?"
		args = append(args, filter.MerchantID)
		// contain '?' for merchant_id in the query select statement
	}

	if filter.Category != "" {
		conditions = append(conditions, "p.category = ?")
		args = append(args, filter.Category)
	}
	if filter.Type != "" {
		conditions = append(conditions, "p.type = ?")
		args = append(args, filter.Type)
	}
	if filter.Subtype != "" {
		conditions = append(conditions, "p.sub_type = ?")
		args = append(args, filter.Subtype)
	}
	if filter.Acquirer != "" {
		conditions = append(conditions, "p.acquirer = ?")
		args = append(args, filter.Acquirer)
	}
	if filter.Status != "" {
		if filter.Status == constant.PaymentMethodGeneralStatusActive {
			conditions = append(conditions, "(pmm.is_active = ? AND pmm.activation_status = ?)")
			args = append(args, true, constant.PaymentMethodActivationStatusApproved)
		} else {
			conditions = append(conditions, "pmm.is_active = ? OR pmm.activation_status != ? OR pmm.activation_status IS NULL")
			args = append(args, false, constant.PaymentMethodActivationStatusApproved)
		}
	}
	if filter.InstallmentPlan.InstallmentPlanID != "" {
		conditions = append(conditions, "? MEMBER OF (pmm.config->>'$.partnerConfig.installment.installmentPlanIds')")
		args = append(args, filter.InstallmentPlan.InstallmentPlanID)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	querySort := " ORDER BY p.updated_at DESC"
	query += querySort
	// END OF QUERY CONDITION

	if err := r.db.SelectContext(ctx, &data, query, args...); err != nil {
		r.logger.Error(ctx, "error when get payment method merchant list", logger.Error(err))

		return nil, err
	}

	for _, d := range data {
		d.IsDerivedMerchant = util.ValueToPtr(isDerived && derivedID != "")

		if d.RequiredDocuments.Valid {
			_ = json.Unmarshal(d.RequiredDocuments.JSONText, &d.RequiredDocumentObjects)
		}

		if d.MerchantConfig.Valid {
			_ = json.Unmarshal(d.MerchantConfig.JSONText, &d.MerchantConfigObj)
		}

		d.UnmarshalConfigObj()
	}

	return data, nil
}

func (r *PaymentMethodRepository) FindPaymentMethodByIdAndMerchant(ctx context.Context, paymentMethodId, merchantId string) (*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/FindPaymentMethodByIdAndMerchant")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var paymentMethod paymentModel.PaymentMethodWithPivot
	query := `SELECT 
    	` + SelectPaymentMethodWithPivotStr + `
	FROM payment_methods p
	LEFT JOIN payment_method_merchant pmm 
		ON p.uuid = pmm.payment_method_id AND pmm.merchant_id = ?
	WHERE p.uuid = ?`

	if err := r.db.GetContext(ctx, &paymentMethod, query, merchantId, merchantId, paymentMethodId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, fmt.Sprintf("get payment method is not found by id=%s", paymentMethodId), logger.Error(err))
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payment method by id=%s", paymentMethodId), logger.Error(err))
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
