package disbursementController

import (
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *Controller) CreateBulk(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/CreateBulk")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}
	defer file.Close()

	totalData, totalAmount := 0, 0.0

	ww := monitor.WrapResponse(w, r)
	defer func() {
		monitor.WriteAndSend(
			ctx, "disbursement-create-bulk", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", user.MerchantId),
					fmt.Sprintf("total:%d", totalData),
					fmt.Sprintf("amount:%.9f", totalAmount),
				}
			},
		)
	}()

	bulkDisbursement, err := c.disbursementSvc.BulkCreate(ctx, &disbursementModel.BulkCreateRequest{
		MerchantId: user.MerchantId,
		File:       file,
		CreatedBy:  user.UUID,
	})
	if err != nil {
		response.SendApiResponseError(ctx, ww, err)

	} else {
		totalData, totalAmount = bulkDisbursement.TotalData, bulkDisbursement.TotalAmount

		response.SendApiResponseOK(ww, bulkDisbursement)
	}
}
