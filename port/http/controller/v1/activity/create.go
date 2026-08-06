package activityController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Create		godoc
// @Summary		Record user activity
// @Description	Record user activity
// @ID			api-users-activities
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body	activityModel.CreateActivityReq true "JSON Body for Record user activity"
// @Success		201  			{object}	response.ApiResponse{data=activityModel.CreateActivityResp}
// @Failure		500  			{object}	response.ApiResponse
// @Router		/api/v1/users/activities [post]
// @Security	Bearer
func (c *activity) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/activity/Create")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := activityModel.CreateActivityReq{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	data := request.Record(user.MerchantId, user.UUID, r)
	if err := c.activitySvc.Create(ctx, data); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseCreated(w, activityModel.CreateActivityResp{ID: data.ID})
	}
}
