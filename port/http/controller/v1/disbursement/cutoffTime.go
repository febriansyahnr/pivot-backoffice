package disbursementController

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetCutOffTimeStatus	godoc
// @Summary				Endpoint to retrieve the status of payout cut-off time
// @Description			Endpoint to retrieve the status of payout cut-off time
// @ID					api-disbursement-cut-off-time-status
// @Tags				API - Disbursement
// @Accept				json
// @Produce				json
// @Success				200	{object}	response.ApiResponse{data=disbursementModel.CutOffTimeStatusResponse}
// @Failure				500 {object}	response.ApiResponse
// @Router				/api/v1/disbursements/cut-off-time-status [get]
// @Security			Bearer
func (c *Controller) GetCutOffTimeStatus(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetCutOffTimeStatus")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	now := time.Now().UTC()
	windowConfig := c.config.DisbursementConfig.CutOffTimeWindow

	if isEarlyCheck, _ := strconv.ParseBool(r.URL.Query().Get("earlyCheck")); isEarlyCheck {

		tz := time.FixedZone("GMT", windowConfig.GMT*60*60)
		duration := time.Duration(-windowConfig.BannerShowBeforeMinute) * time.Minute

		t1, _ := time.ParseInLocation(time.DateTime, now.In(tz).Format(time.DateOnly)+" "+windowConfig.StartTime+":00", tz)
		if t1.IsZero() {
			response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid start time window format")))
			return
		}
		t2 := t1.Add(duration)

		if t1.Day() != t2.Day() {
			windowConfig.SameDay = false
		}
		windowConfig.StartTime = t2.Format(time.TimeOnly)[:5]
	}

	exactResult, err := c.disbursementSvc.GetCutOffTimeStatus(ctx, now, user.MerchantId, &windowConfig)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, exactResult)
}
