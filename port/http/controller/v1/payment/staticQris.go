package payment

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *PaymentController) FilterStaticQrisList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/FilterStaticQrisList")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var (
		request paymentModel.StaticQrisFilterRequest
		err     error
	)
	request.Page = paymentModel.DefaultStaticQrisPage
	request.PerPage = paymentModel.DefaultStaticQrisPerPage
	request.Sort = paymentModel.DefaultStaticQrisSort
	request.SortBy = paymentModel.DefaultStaticQrisSortBy

	if r.URL.Query().Get("page") != "" {
		request.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		request.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		request.StartDate = d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		request.EndDate = d
	}

	if r.URL.Query().Get("sort") != "" {
		request.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		request.SortBy = r.URL.Query().Get("sortBy")
	}

	request.Status = r.URL.Query().Get("status")
	request.ID = r.URL.Query().Get("id")

	request.MerchantID = user.MerchantId
	result, err := c.paymentService.FilterStaticQrisList(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *PaymentController) GetStaticQrisDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/GetStaticQrisDetail")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	paymentID := chi.URLParam(r, "paymentId")
	if paymentID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("payment ID is required")))
		return
	}

	request := paymentModel.StaticQrisDetailRequest{
		PaymentID:  paymentID,
		MerchantID: user.MerchantId,
	}

	result, err := c.paymentService.GetStaticQrisDetail(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}

func (c *PaymentController) GetStaticQrisTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/GetStaticQrisTransactions")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	paymentID := chi.URLParam(r, "paymentId")
	if paymentID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("payment ID is required")))
		return
	}

	var (
		request paymentModel.StaticQrisTransactionFilterRequest
		err     error
	)
	request.Page = paymentModel.DefaultStaticQrisPage
	request.PerPage = paymentModel.DefaultStaticQrisPerPage
	request.Sort = paymentModel.DefaultStaticQrisSort
	request.SortBy = paymentModel.DefaultStaticQrisTransactionSortBy

	if r.URL.Query().Get("page") != "" {
		request.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		request.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		request.StartDate = d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}
		request.EndDate = d
	}

	if r.URL.Query().Get("sort") != "" {
		request.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		request.SortBy = r.URL.Query().Get("sortBy")
	}

	request.Status = r.URL.Query().Get("status")
	request.ID = r.URL.Query().Get("id")

	request.PaymentID = paymentID
	request.MerchantID = user.MerchantId

	result, err := c.paymentService.GetStaticQrisTransactions(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *PaymentController) DeactivateStaticQris(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/DeactivateStaticQris")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	reqEncodedPin := r.Header.Get(constant.HeaderXRequestPIN)
	if reqEncodedPin == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrRequiredPIN))
		return
	}

	pin, _ := base64.StdEncoding.DecodeString(reqEncodedPin)
	if err := c.userService.CheckCurrentPin(ctx, user.UUID, string(pin)); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	paymentID := chi.URLParam(r, "paymentId")
	if paymentID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("payment ID is required")))
		return
	}

	var request paymentModel.StaticQrisUpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	err := c.paymentService.DeactivateStaticQris(r.Context(), paymentID, user.MerchantId, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]string{"message": "Static QRIS deactivated successfully"})
}

func (c *PaymentController) GetMaxActiveStaticQRPerMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/GetMaxActiveStaticQRPerMerchant")
	defer segment.End()

	maxActive := c.paymentService.GetMaxActiveStaticQRPerMerchant()

	response.SendApiResponseOK(w, map[string]int{"maxActiveStaticQRPerMerchant": maxActive})
}
