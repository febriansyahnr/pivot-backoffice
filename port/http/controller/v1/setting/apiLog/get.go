package apiLog

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/apiLog/GetList")
	defer span.End()

	var (
		err            error
		page           int64      = 1 // default 1
		perPage        int64      = 10
		startCreatedAt *time.Time // default nil
		endCreatedAt   *time.Time // default nil
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")
	startCreatedAtStr := r.URL.Query().Get("startCreatedAt")
	endCreatedAtStr := r.URL.Query().Get("endCreatedAt")
	originID := r.URL.Query().Get("keyword")
	status := r.URL.Query().Get("status")
	method := r.URL.Query().Get("method")
	product := r.URL.Query().Get("product")

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
	if startCreatedAtStr != "" {
		parsedStartUpdatedAt, err := time.Parse(util.UTCLayout, startCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid startUpdatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		startCreatedAt = &parsedStartUpdatedAt
	}
	if endCreatedAtStr != "" {
		parsedEndUpdatedAt, err := time.Parse(util.UTCLayout, endCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid endUpdatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		endCreatedAt = &parsedEndUpdatedAt
	}

	filter := &inboundModel.GetInboundFilterRequest{
		MerchantID:     user.MerchantId,
		OriginID:       originID,
		StartCreatedAt: startCreatedAt,
		EndCreatedAt:   endCreatedAt,
		Page:           page,
		PerPage:        perPage,
		Status:         status,
		Method:         method,
		Product:        product,
	}

	data, err := h.inboundSvc.GetList(ctx, filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, data.Data, data.Meta)
}

func (h *handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/apiLog/GetByID")
	defer span.End()

	// Get User Info from jwt token
	_, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	data, err := h.inboundSvc.GetByID(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, data)
}

func (h *handler) GetSnapVersionByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/apiLog/GetSnapVersionByID")
	defer span.End()

	// Get User Info from jwt token
	_, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	data, err := h.inboundSvc.GetSnapVersionByID(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, data)
}
