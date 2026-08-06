package payment

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/virtual-card/model"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *PaymentController) FilterStaticVaList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/FilterStaticVaList")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var (
		request paymentModel.StaticVaFilterRequest
		err     error
	)
	request.Page = paymentModel.DefaultStaticVaPage
	request.PerPage = paymentModel.DefaultStaticVaPerPage
	request.Sort = paymentModel.DefaultStaticVaSort
	request.SortBy = paymentModel.DefaultStaticVaSortBy

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
	request.BankName = r.URL.Query().Get("bankName")

	request.MerchantID = user.MerchantId
	result, err := c.paymentService.FilterStaticVaList(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *PaymentController) GetStaticVaDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/GetStaticVaDetail")
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

	request := paymentModel.StaticVaDetailRequest{
		PaymentID:  paymentID,
		MerchantID: user.MerchantId,
	}

	result, err := c.paymentService.GetStaticVaDetail(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}

func (c *PaymentController) GetStaticVaTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/GetStaticVaTransactions")
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
		request paymentModel.StaticVaTransactionFilterRequest
		err     error
	)
	request.Page = paymentModel.DefaultStaticVaPage
	request.PerPage = paymentModel.DefaultStaticVaPerPage
	request.Sort = paymentModel.DefaultStaticVaSort
	request.SortBy = paymentModel.DefaultStaticVaTransactionSortBy

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

	result, err := c.paymentService.GetStaticVaTransactions(r.Context(), request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *PaymentController) DeactivateStaticVa(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/payments/DeactivateStaticVa")
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

	var request paymentModel.StaticVaUpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	err := c.paymentService.DeactivateStaticVa(r.Context(), paymentID, user.MerchantId, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]string{"message": "Static VA deactivated successfully"})
}

func (c *PaymentController) GetVARangeList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/payments/GetVARangeList")
	defer span.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantID := user.MerchantId
	merchantMID := ""
	_, err := func() (*model.Merchant, error) {
		merchant, err := c.merchantService.FindMerchantByID(ctx, merchantID)
		if err != nil {
			c.logger.Error(ctx, "error when get merchant", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrDatabase, err)
		} else if merchant == nil {
			c.logger.Error(ctx, "not found merchant", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
		}
		merchantID = merchant.UUID
		merchantMID = merchant.MID.String

		if merchant.ParentID.Valid && merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			merchantID = merchant.ParentID.String

			parentMerchant, err := c.merchantService.FindMerchantByID(ctx, merchantID)
			if err != nil {
				c.logger.Error(ctx, "error when get parent merchant", logger.Error(err))
				return nil, pkgErr.New(response.HttpErrDatabase, err)
			} else if parentMerchant != nil && parentMerchant.MID.Valid {
				merchantMID = parentMerchant.MID.String
			}
		}

		return nil, nil
	}()
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Get Static VA Config
	status := r.URL.Query().Get("status")
	result, err := c.paymentMethodService.GetStaticVAPaymentMethodByMerchant(r.Context(), &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantID,
		Status:     status,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, paymentModel.ParsePaymentMethodsToVARangeResponseList(result, merchantMID))
}

func (c *PaymentController) UpdateVARange(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/payments/UpdateVARange")
	defer span.End()

	var (
		payload *paymentMethodModel.UpdateVAStaticRangeRequest
		err     error
	)

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	paymentMethodID := chi.URLParam(r, "id")
	if err = uuid.Validate(paymentMethodID); err != nil {
		response.SendGeneralResponseError(w, pkgErr.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErr.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgErr.New(response.HttpErrRequest, err))
		return
	}

	merchantID := user.MerchantId
	merchantMID := ""
	_, err = func() (*model.Merchant, error) {
		merchant, err := c.merchantService.FindMerchantByID(ctx, merchantID)
		if err != nil {
			c.logger.Error(ctx, "error when get merchant", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrDatabase, err)
		} else if merchant == nil {
			c.logger.Error(ctx, "not found merchant", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
		}
		merchantID = merchant.UUID
		merchantMID = merchant.MID.String

		if merchant.ParentID.Valid && merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			merchantID = merchant.ParentID.String

			parentMerchant, err := c.merchantService.FindMerchantByID(ctx, merchantID)
			if err != nil {
				c.logger.Error(ctx, "error when get parent merchant", logger.Error(err))
				return nil, pkgErr.New(response.HttpErrDatabase, err)
			} else if parentMerchant != nil && parentMerchant.MID.Valid {
				merchantMID = parentMerchant.MID.String
			}
		}

		return nil, nil
	}()
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	requestItems := []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{}
	if payload.CloseRange != nil {
		// Set for closed static
		requestItems = append(requestItems, paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
			BINPrefix:  payload.CloseRange.BinPrefix,
			StartRange: payload.CloseRange.Start,
			EndRange:   payload.CloseRange.End,
			Type:       paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC,
		})
	}
	if payload.OpenRange != nil {
		// Set for open static
		requestItems = append(requestItems, paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
			BINPrefix:  payload.OpenRange.BinPrefix,
			StartRange: payload.OpenRange.Start,
			EndRange:   payload.OpenRange.End,
			Type:       paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
		})
	}

	request := &paymentMethodModel.SetupPaymentMethodConfigRequest{
		MerchantID:      merchantID,
		PaymentMethodID: paymentMethodID,
		PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
			VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
				Items:       requestItems,
				MerchantID:  merchantID,
				MerchantMID: merchantMID,
			},
		},
	}
	if err = c.paymentMethodService.SetupConfig(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
