package disbursementController

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ExportToExcel	godoc
// @Summary			Export disbursement list endpoint
// @Description		Export disbursement list endpoint
// @ID				api-export-disbursement-list
// @Tags			API - Disbursement
// @Accept			json
// @Produce			json
// @Success			200  		{object}	response.ApiResponse{data=[]disbursementModel.Disbursement,meta=commonModel.Meta}
// @Failure			500  		{object}	response.ApiResponse
// @Router			/api/v1/disbursements/export [post]
// @Security		Bearer
func (c *Controller) ExportToExcel(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/ExportToExcel")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var filter disbursementModel.GetDisbursementFilterRequest
	if err = json.NewDecoder(r.Body).Decode(&filter); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	filter.MerchantID = user.MerchantId
	if filter.StartCreatedAt != nil {
		*filter.StartCreatedAt, err = util.TimeToUTC(util.ValueOfPtr(filter.StartCreatedAt), r.Header.Get(constant.HeaderTimeZoneKey))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
			return
		}
	}

	if filter.EndCreatedAt != nil {
		*filter.EndCreatedAt, err = util.TimeToUTC(util.ValueOfPtr(filter.EndCreatedAt), r.Header.Get(constant.HeaderTimeZoneKey))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
			return
		}
	}

	ctx = context.WithValue(ctx, constant.CtxTimeZone, r.Header.Get(constant.HeaderTimeZoneKey))
	resp, err := c.disbursementSvc.ExportToExcel(ctx, &filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
