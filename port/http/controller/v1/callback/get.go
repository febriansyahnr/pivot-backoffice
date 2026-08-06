package callbackController

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CallbackController) GetCallbackList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/callback/GetCallbackList")
	defer segment.End()

	var (
		err  error
		page int64 = 1 // default 1
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}

	var perPage int64 = constant.DefaultPaginationPageSize
	if c.config != nil {
		perPage = c.config.AppConfig.PaginationPerPage
	}

	filter := &callbackModel.GetListCallbackFilterRequest{
		MerchantID: &user.MerchantId,
	}

	// Hit Register Callback
	callback, err := c.callbackSvc.GetList(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, callback.Data, callback.Meta)
}
