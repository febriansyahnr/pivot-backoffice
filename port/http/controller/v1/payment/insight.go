package payment

import (
	e "errors"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// GetPaymentInsight		godoc
// @Summary		Payment insight Endpoint
// @Description	Payment Insight Endpoint, currently we provide payment balance
// @ID			api-payment-insight
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=paymentModel.PaymentInsightResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/insights [get]
// @Security	Bearer
func (c *PaymentController) GetPaymentInsight(w http.ResponseWriter, r *http.Request) {
	var (
		insightResponse paymentModel.PaymentInsightResponse
		totalBalance    *commonModel.Amount
		totalSuccess    *paymentModel.PaymentInsightItem
		totalPending    *paymentModel.PaymentInsightItem
		totalVoid       *paymentModel.PaymentInsightItem
		ctx             = r.Context()
		eg              = new(errgroup.Group)
		err             error
	)
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payment/GetPaymentInsight")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantID, err := uuid.Parse(user.MerchantId)
	if err != nil {
		err = fmt.Errorf("merchant id not valid")
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	eg.Go(func() error {
		var err error
		totalBalance, err = c.paymentService.GetTotalPaymentBalance(ctx, merchantID)
		if err != nil {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		var err error
		totalSuccess, err = c.paymentService.GetTodayPaymentInsight(ctx, paymentModel.PaymentInsightOption{
			MerchantID: user.MerchantId,
			Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
		})
		if err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		totalPending, err = c.paymentService.GetTodayPaymentInsight(ctx, paymentModel.PaymentInsightOption{
			MerchantID: user.MerchantId,
			Status:     paymentConstant.PAYMENT_STATUS_PENDING,
		})
		if err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		totalVoid, err = c.paymentService.GetTodayPaymentInsight(ctx, paymentModel.PaymentInsightOption{
			MerchantID: user.MerchantId,
			Status:     paymentConstant.PAYMENT_STATUS_VOID,
		})
		if err != nil {
			return err
		}
		return nil
	})

	err = eg.Wait()
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	insightResponse.TotalBalance = totalBalance
	insightResponse.TodayTotalSuccess = totalSuccess
	insightResponse.TodayTotalPending = totalPending
	insightResponse.TodayTotalVoid = totalVoid

	response.SendApiResponseOK(w, insightResponse)
}

func (h *PaymentController) GetPaymentDashboardInsights(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/GetPaymentDashboardInsights")
	defer segment.End()

	userTokenClaim, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "insightStartDate", "insightEndDate"); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	queryParams := r.URL.Query()

	request := paymentModel.GetPaymentDashboardInsightRequest{
		MerchantId: userTokenClaim.MerchantId,
	}
	request.StartDate, _ = time.Parse(time.RFC3339Nano, queryParams.Get("insightStartDate"))
	request.EndDate, _ = time.Parse(time.RFC3339Nano, queryParams.Get("insightEndDate"))

	// Validation of value formats, time ranges, and other rules are already handled in ValidateReportDateRangeFromRequest.
	if request.StartDate.IsZero() || request.EndDate.IsZero() {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, e.New("start date or end date cannot be empty")))
		return
	}

	result, err := h.paymentService.GetPaymentDashboardInsights(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, constant.ErrInternalServerForUser))
		return
	}
	response.SendApiResponseOK(w, result)
}
