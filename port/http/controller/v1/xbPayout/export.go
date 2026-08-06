package xbPayoutController

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ExportToExcel godoc
// @Summary			Export XB payout list endpoint
// @Description		Export XB payout list endpoint
// @ID				api-export-xb-payout-list
// @Tags			API - XB Payout
// @Accept			json
// @Produce			json
// @Success			200  		{object}	response.ApiResponse{data=xbModel.ExportXbPayoutResponse}
// @Failure			500  		{object}	response.ApiResponse
// @Router			/api/v1/xb/payout/export [post]
// @Security		Bearer
func (c *xbPayoutController) ExportToExcel(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/ExportToExcel")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var filter xbModel.ExportXbPayoutRequest
	if err = json.NewDecoder(r.Body).Decode(&filter); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	filter.MerchantID = user.MerchantId

	if filter.StartAt != nil {
		*filter.StartAt, err = util.TimeToUTC(util.ValueOfPtr(filter.StartAt), r.Header.Get(constant.HeaderTimeZoneKey))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
	}

	if filter.EndAt != nil {
		*filter.EndAt, err = util.TimeToUTC(util.ValueOfPtr(filter.EndAt), r.Header.Get(constant.HeaderTimeZoneKey))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
	}

	ctx = context.WithValue(ctx, constant.CtxTimeZone, r.Header.Get(constant.HeaderTimeZoneKey))
	resp, err := c.xbPayoutSvc.ExportToExcel(ctx, &filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
