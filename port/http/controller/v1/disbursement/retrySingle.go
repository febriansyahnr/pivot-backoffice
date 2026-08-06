package disbursementController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// RetrySingle		godoc
// @Summary			Retry single disbursement
// @Description		Retry single disbursement
// @ID				api-retry-single-disbursements
// @Tags			API - Disbursement
// @Accept			json
// @Produce			json
// @Param			Request	body		disbursementModel.RetrySingleRequest true "JSON Body for retry single disbursement"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/disbursements/single/retry [post]
// @Security		Bearer
func (c *Controller) RetrySingle(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		ctx = r.Context()
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/disbursement/RetrySingle")
	defer func() {
		if err != nil {
			segment.End()
			return
		}

		segment.End()
	}()

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = constant.ErrUserNotFound
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	var requestPayload disbursementModel.RetrySingleRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Validate request
	if err = c.validate.Struct(requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Set merchant ID
	requestPayload.MerchantID = user.MerchantId

	err = c.disbursementSvc.RetrySingle(r.Context(), &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
