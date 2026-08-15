package merchant

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UpdateKYC updates the KYC status of a merchant in the database.
// It takes a context and a payload containing the merchant's KYC update request.
// The function returns an error if the update fails or if the merchant is not found.
func (r *MerchantRepository) UpdateKYC(ctx context.Context, payload merchant.UpdateMerchantKYCRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/UpdateKYC")
	defer segment.End()

	ok, err := r.db.ExecContext(ctx, `
		UPDATE merchants
		SET kyc_status = ?, 
			status = ?,
			mid = ?,
			updated_at = ?
		WHERE uuid = ?
	`, payload.KYCStatus, payload.MerchantStatus, payload.MID, time.Now().UTC(), payload.MerchantID)
	if err != nil {
		return pkgErrs.New(response.HttpErrInternal, err)
	}

	if !ok {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	return nil
}
