package subMerchant

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantController) ExportPaymentHistory(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/ExportPaymentHistory")
	defer segment.End()

	parentMerchant, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	subMerchantID := chi.URLParam(r, "subMerchantId")
	if _, err := uuid.Parse(subMerchantID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIncorrectSubMerchant))
		return
	}

	if err := c.validateSubMerchantOwnership(ctx, parentMerchant.MerchantId, subMerchantID); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	request := &paymentModel.PaymentDownloadHistoryRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	request.MerchantId = subMerchantID

	if err := c.validate.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(request, "startDate", "endDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.paymentService.Export(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}

func (c *SubMerchantController) ExportDisbursementHistory(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/ExportDisbursementHistory")
	defer segment.End()

	parentMerchant, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	subMerchantID := chi.URLParam(r, "subMerchantId")
	if _, err := uuid.Parse(subMerchantID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIncorrectSubMerchant))
		return
	}

	if err := c.validateSubMerchantOwnership(ctx, parentMerchant.MerchantId, subMerchantID); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	var req disbursementModel.ExportDisbursementFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	tz := r.Header.Get(constant.HeaderTimeZoneKey)
	filter := disbursementModel.GetDisbursementFilterRequest{
		MerchantID:        subMerchantID,
		Status:            req.Status,
		TransactionStatus: req.TransactionStatus,
		Type:              req.Type,
		Keyword:           req.Keyword,
		Sort:              req.Sort,
		SortBy:            req.SortBy,
	}

	if req.StartCreatedAt != "" {
		t, err := time.ParseInLocation(constant.DateFormat, req.StartCreatedAt, time.UTC)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
		if tz != "" {
			t, err = util.TimeToUTC(t, tz)
			if err != nil {
				response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
				return
			}
		}
		filter.StartCreatedAt = &t
	}

	if req.EndCreatedAt != "" {
		t, err := time.ParseInLocation(constant.DateFormat, req.EndCreatedAt, time.UTC)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		if tz != "" {
			t, err = util.TimeToUTC(t, tz)
			if err != nil {
				response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
				return
			}
		}
		filter.EndCreatedAt = &t
	}

	ctx = context.WithValue(ctx, constant.CtxTimeZone, tz)
	resp, err := c.disbursementSvc.ExportToExcel(ctx, &filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}

func (c *SubMerchantController) validateSubMerchantOwnership(ctx context.Context, parentMerchantID, subMerchantID string) error {
	subMerchant, err := c.merchantSvc.FindMerchantByID(ctx, subMerchantID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidateSubMerchantParent)
	}
	if subMerchant == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrSubMerchantNotFound)
	}
	if subMerchant.ParentID.String != parentMerchantID {
		return pkgErrors.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
	}

	return nil
}
