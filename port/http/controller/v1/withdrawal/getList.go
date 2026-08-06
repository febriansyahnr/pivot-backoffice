package withdrawalController

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// GetList		godoc
// @Summary		Endpoint to display withdrawal history list
// @Description	Endpoint to display withdrawal history list
// @ID			withdrawal-get-list
// @Tags		API - Withdrawal
// @Accept		json
// @Produce		json
// @Param 		account	path					string true "Account name for payments or payouts"
// @Success		200  	{object}				response.ApiResponse{data=[]withdrawal.WithdrawalHistoryResponse}
// @Failure		500  	{object}				response.ApiResponse
// @Router		/api/v1/withdrawals/{account} 	[get]
// @Security 	Bearer
func (h *handler) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawal/GetList")
	defer segment.End()

	account := r.PathValue("account")
	if !slices.Contains(pathAccountNames, account) {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrNotFound, constant.ErrInvalidPath))
		return
	}

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	err := error(nil)
	request := &withdrawal.WithdrawalHistoryRequest{
		WithdrawalListRequest: &withdrawal.WithdrawalListRequest{
			StrStartDate: r.URL.Query().Get("startDate"), StrEndDate: r.URL.Query().Get("endDate"),
			Status:      r.URL.Query().Get("status"),
			Id:          r.URL.Query().Get("id"),
			Sort:        r.URL.Query().Get("sort"),
			MerchantId:  user.MerchantId,
			AccountName: constant.TypeDisbursement,
		},
	}
	if err = h.preparationGetList(r, account, request.WithdrawalListRequest); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if request.Page, err = strconv.Atoi(page); err != nil {
			response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead")))
			return
		}
	}
	if perPage := r.URL.Query().Get("perPage"); perPage != "" {
		if request.PerPage, err = strconv.Atoi(perPage); err != nil {
			response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead")))
			return
		}
	}

	if err = h.validator.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if request.Page == 0 {
		request.Page = 1
	}
	if request.PerPage == 0 {
		request.PerPage = 10

	} else if request.PerPage > 100 {
		request.PerPage = 100
	}

	h.logger.Info(ctx, "Withdrawal history list", logger.Any("request", request))

	if list, err := h.service.GetList(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
	}
}
