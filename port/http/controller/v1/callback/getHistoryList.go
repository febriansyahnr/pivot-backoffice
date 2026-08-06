package callbackController

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CallbackController) GetCallbackLogList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/callback/GetCallbackLogList")
	defer segment.End()

	var (
		err            error
		page           int64      = 1 // default 1
		perPage        int64      = constant.DefaultPaginationPageSize
		startUpdatedAt *time.Time // default nil
		endUpdatedAt   *time.Time // default nil
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")
	startUpdatedAtStr := r.URL.Query().Get("startUpdatedAt")
	endUpdatedAtStr := r.URL.Query().Get("endUpdatedAt")
	callbackType := r.URL.Query().Get("type")
	event := r.URL.Query().Get("event")
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")

	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}
	if perPageStr != "" {
		perPage, err = strconv.ParseInt(perPageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}
	}
	// Validation and parsing
	if startUpdatedAtStr != "" {
		parsedStartUpdatedAt, err := time.Parse(util.UTCLayout, startUpdatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid startUpdatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		startUpdatedAt = &parsedStartUpdatedAt
	}
	if endUpdatedAtStr != "" {
		parsedEndUpdatedAt, err := time.Parse(util.UTCLayout, endUpdatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid endUpdatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		endUpdatedAt = &parsedEndUpdatedAt
	}

	filter := &callbackModel.GetListCallbackLogFilterRequest{
		MerchantID:     user.MerchantId,
		Type:           callbackType,
		Event:          event,
		StartUpdatedAt: startUpdatedAt,
		EndUpdatedAt:   endUpdatedAt,
		Status:         status,
		Keyword:        keyword,
	}

	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagCallback,
		constant.ActivityUserAccessCallbackHistory,
		filter,
	)

	// Hit Callback Log List
	callback, err := c.callbackSvc.GetCallbackLogList(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, callback.Data, callback.Meta)
}
