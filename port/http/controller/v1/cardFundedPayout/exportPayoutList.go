package cardFundedPayoutController

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ExportPayoutList godoc
// @Summary		Export card funded payout list endpoint
// @Description	Export card funded payout list to Excel file
// @ID			api-export-card-funded-payout-list
// @Tags		API - Card Funded Payout
// @Accept		json
// @Produce		json
// @Param       startDate			query     	string  false  "filter by start date (ISO 8601)"
// @Param       endDate				query     	string  false  "filter by end date (ISO 8601)"
// @Param       transactionStatus	query     	string  false  "filter by transaction status (processing/success/failed)"
// @Param       approval			query     	string  false  "filter by approval status (approved/rejected/waiting)"
// @Param       searchId			query     	string  false  "search by payout ID or reference ID"
// @Param       sort				query     	string  false  "sort direction (ASC/DESC)"
// @Param       sortBy				query     	string  false  "field to sort by"
// @Success		200  				{object}	response.ApiResponse{data=cardFundedPayoutModel.ExportPayoutListResponse}
// @Failure		500  				{object}	response.ApiResponse
// @Router		/api/v1/card-funded-payouts/export [post]
// @Security	Bearer
func (c *handler) ExportPayoutList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/ExportPayoutList")
	defer span.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var filter cardFundedPayoutModel.FilterGetPayoutList
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Set merchant ID from token
	filter.MerchantID = user.MerchantId

	// Process date filters with timezone
	if filter.StartCreatedAt != nil {
		var err error
		*filter.StartCreatedAt, err = util.TimeToUTC(util.ValueOfPtr(filter.StartCreatedAt), r.Header.Get(constant.HeaderTimeZoneKey))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
	}

	if filter.EndCreatedAt != nil {
		var err error
		*filter.EndCreatedAt, err = util.TimeToUTC(util.ValueOfPtr(filter.EndCreatedAt), r.Header.Get(constant.HeaderTimeZoneKey))
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
	}

	// Add timezone to context for Excel generation
	ctx = context.WithValue(ctx, constant.CtxTimeZone, r.Header.Get(constant.HeaderTimeZoneKey))

	resp, err := c.cardFundedPayoutService.ExportPayoutList(ctx, &filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
