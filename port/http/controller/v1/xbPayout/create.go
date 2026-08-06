package xbPayoutController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *xbPayoutController) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/CreateSession")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Find merchant
	merchant, err := c.merchantSvc.FindMerchantByID(ctx, user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	} else if merchant == nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantNotFound))
		return
	}

	payload := xbModel.CreatePayoutSessionRequest{
		ReferenceId:  fmt.Sprintf("XB%d", time.Now().UTC().Unix()),
		MerchantId:   user.MerchantId,
		MerchantName: merchant.Name,
		CreatedFrom:  constant.DisbursementCreatedFromMerchantPortal,
		CreatedBy:    user.UUID,
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validateCreateRequest(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Create payout
	payout, err := c.xbPayoutSvc.CreateSession(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, payout)
}

func (c *xbPayoutController) validateCreateRequest(req xbModel.CreatePayoutSessionRequest) error {
	// First, validate the main struct fields
	err := c.validate.StructExcept(req, "SenderData", "BeneficiaryData")
	if err != nil {
		return err
	}

	// Conditionally validate SenderData if SenderID is empty
	if req.SenderID == "" && req.SenderData != nil {
		err = c.validate.Struct(req.SenderData)
		if err != nil {
			return err
		}
	}

	// Conditionally validate BeneficiaryData if BeneficiaryID is empty
	if req.BeneficiaryID == "" && req.BeneficiaryData != nil {
		err = c.validate.Struct(req.BeneficiaryData)
		if err != nil {
			return err
		}
	}

	return nil
}
