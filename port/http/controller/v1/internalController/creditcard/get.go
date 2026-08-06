package creditcard

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
)

// GetPaymentById	godoc
// @Summary			Get credit card payment by uuid
// @Description		Get credit card payment by uuid
// @ID				api-creditcard-get-payment-by-uuid
// @Tags			API - Credit Card
// @Accept			json
// @Produce			json
// @Param			reference_id	path		string	true	"UUID of the credit card payment"
// @Success			200  			{object}	response.OpenApiResponse
// @Failure			500  			{object}	response.OpenApiResponse
// @Router			/api/v1/cards/payment-session/{reference_id} [get]
// @Security		Bearer
func (c *Controller) GetPaymentById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/creditcard/GetPaymentById")
	defer segment.End()

	var (
		err error
		now = time.Now()
	)

	merchantToken, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)

	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	merchantID := merchantToken.MerchantId
	uuid := chi.URLParam(r, "uuid") // :uuid?isNetworkToken=true
	isNetworkToken := r.URL.Query().Get("isNetworkToken")

	if uuid == "" {
		err = fmt.Errorf("uuid required")
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer monitor.WriteAndSend(
		ctx, "creditcard-get-payment-by-id", now, ww, err, func() []string {
			return []string{
				fmt.Sprintf("merchant_id:%s", merchantID),
				fmt.Sprintf("uuid:%s", uuid),
			}
		},
	)

	data, err := c.creditcardSvc.GetPaymentById(ctx, merchantID, uuid)
	if err != nil {
		response.SendOpenApiResponseError(ww, err)
		return
	}

	// get account_transactions id by reference_id from payment
	ledger, err := c.orchestratorSvc.FindByReference(ctx, data.UUID, constant.TypePayment)
	if err != nil {
		response.SendOpenApiResponseError(ww, err)
		return
	}

	var resp interface{}
	if strings.Contains(r.URL.Path, "/open-api") {
		resp, err = data.ToOpenAPICreditcardGetPaymentByUUID()
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenGetOpenAPIGetPaymentDetailByID.Error(),
				logger.Error(err),
				logger.String("uuid", uuid))
			response.SendOpenApiResponseError(ww, pkgErrors.New(response.HttpErrInternal, err))
			return
		}

		if ledger != nil {
			resp.(*creditcardModel.OpenAPIGetCardPaymentByIdResponse).ChargeID = ledger.UUID.String()
		}
	} else {
		resp, err = data.ToInternalCreditcardGetPaymentByUUID()
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenGetInternalGetPaymentDetailByID.Error(),
				logger.Error(err),
				logger.String("uuid", uuid))
			response.SendOpenApiResponseError(ww, pkgErrors.New(response.HttpErrInternal, err))
			return
		}

		// Fetch customer data to populate customer object
		customerResp, err := c.customerSvc.GetCustomerById(ctx, data.CustomerID, merchantID)
		if err != nil {
			c.logger.Error(ctx, "failed to get customer data",
				logger.Error(err),
				logger.String("customerId", data.CustomerID),
				logger.String("merchantId", merchantID))
			// Continue without customer data if customer fetch fails
		}

		// retrieve card config from payment metadata
		unifiedPaymentMetadata := data.ToUnifiedPaymentMetadata()
		// Update response to include customer object instead of customerId
		internalResp := resp.(*creditcardModel.InternalGetCardPaymentByIdResponse)
		if customerResp != nil {
			internalResp.Customer = &creditcardModel.CustomerInfo{
				UUID:         customerResp.UUID,
				MerchantUUID: customerResp.MerchantID,
				FirstName:    customerResp.FirstName,
				LastName:     customerResp.LastName,
				Email:        customerResp.Email,
				Phone:        customerResp.PhoneNumber,
			}
		}

		// Find PaymentMethod by PaymentMethodID
		paymentMethod, err := c.paymentMethodSvc.FindPaymentMethodByIdAndMerchant(ctx, data.PaymentMethodID, data.MerchantID)
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenGetInternalGetPaymentDetailByID.Error(),
				logger.Error(err),
				logger.String("uuid", uuid))
			response.SendOpenApiResponseError(ww, pkgErrors.New(response.HttpErrInternal, err))
			return
		}
		if paymentMethod != nil &&
			paymentMethod.MerchantConfigObj != nil &&
			paymentMethod.MerchantConfigObj.PartnerConfig != nil &&
			paymentMethod.MerchantConfigObj.PartnerConfig.Card != nil &&
			internalResp.Metadata != nil {

			if isNetworkToken == "true" {
				initiator := ""
				if unifiedPaymentMetadata != nil && unifiedPaymentMetadata.PaymentMethodOptions.Card != nil && unifiedPaymentMetadata.PaymentMethodOptions.Card.CardOnFile != nil {
					initiator = unifiedPaymentMetadata.PaymentMethodOptions.Card.CardOnFile.Initiator
				}
				networkTokenProcessingConfig := paymentMethod.MerchantConfigObj.PartnerConfig.Card.GetNetworkTokenPartnerConfig(initiator)

				if internalResp.Metadata.ProcessingConfig == nil {
					internalResp.Metadata.ProcessingConfig = &creditcardModel.CreditCardProcessingConfig{}
				}
				if networkTokenProcessingConfig != nil {
					internalResp.Metadata.ProcessingConfig.BankMerchantId = networkTokenProcessingConfig.AcquirerMerchantID
					internalResp.Metadata.ProcessingConfig.MerchantIdTag = networkTokenProcessingConfig.MerchantIDTag
				}
			}

			internalResp.Metadata.CardPartnerConfigs = paymentMethod.MerchantConfigObj.PartnerConfig.Card.GetPartnerConfigByPaymentType(data.GetGroupPaymentType())
		}

		// Adjustment partner config for virtual terminal
		virtualTerminal := unifiedPaymentMetadata.VirtualTerminal
		if virtualTerminal != nil && virtualTerminal.AcquirerMerchantID != "" {
			internalResp.Metadata.CardPartnerConfigs = &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
					{
						TravelAgentCode:    virtualTerminal.TravelAgentCode,
						AcquirerMerchantID: virtualTerminal.AcquirerMerchantID,
						CardTypes:          virtualTerminal.AllowedCardTypes,
						PrincipalAvailable: virtualTerminal.AllowedPrincipal,
						SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
							// Specifically for this feature, flagging AllowForeignCard and AllowBypass3Ds is always true.
							AllowForeignCard: true,
							AllowBypass3Ds:   true,
							// List of allowed BIN numbers.
							AllowedBinNumbers: virtualTerminal.AllowedBinNumbers,
						},
						IsActive: true,
					},
				},
			}
		}

		// If this is a sub-merchant transaction, fetch parent merchant's card partner config for validation rules
		if internalResp.Metadata != nil && internalResp.Metadata.OnBehalf != nil && internalResp.Metadata.OnBehalf.ParentMerchantId != "" {
			parentMerchantID := internalResp.Metadata.OnBehalf.ParentMerchantId
			parentPaymentMethod, err := c.paymentMethodSvc.FindPaymentMethodByIdAndMerchant(ctx, data.PaymentMethodID, parentMerchantID)
			if err != nil {
				c.logger.Warn(ctx, "failed to get parent merchant payment method config",
					logger.Error(err),
					logger.String("parent_merchant_id", parentMerchantID))
			} else if parentPaymentMethod != nil &&
				parentPaymentMethod.MerchantConfigObj != nil &&
				parentPaymentMethod.MerchantConfigObj.PartnerConfig != nil &&
				parentPaymentMethod.MerchantConfigObj.PartnerConfig.Card != nil &&
				len(parentPaymentMethod.MerchantConfigObj.PartnerConfig.Card.Items) > 0 {

				internalResp.Metadata.ParentCardPartnerConfigs = parentPaymentMethod.MerchantConfigObj.PartnerConfig.Card
				c.logger.Info(ctx, "parent validation rules added to ParentCardPartnerConfigs",
					logger.String("parent_merchant_id", parentMerchantID),
					logger.Int("config_count", len(parentPaymentMethod.MerchantConfigObj.PartnerConfig.Card.Items)))
			}
		}

		// Get and set FDS config to response
		fdsConfig, err := c.merchantSvc.GetFDSConfig(ctx, data.MerchantID)
		if err != nil && !strings.Contains(err.Error(), constant.ErrMerchantNotFound.Error()) {
			c.logger.Error(ctx, constant.ErrWhenGetInternalGetPaymentDetailByID.Error(),
				logger.Error(err),
				logger.String("uuid", uuid))
			response.SendOpenApiResponseError(ww, pkgErrors.New(response.HttpErrInternal, err))
			return

		} else if fdsConfig != nil && internalResp.Metadata != nil && fdsConfig.FDSConfig.BypassExternalPaymentCheck != nil {
			internalResp.Metadata.BypassExternalFdsEvaluation = *fdsConfig.FDSConfig.BypassExternalPaymentCheck
		}

		if ledger != nil {
			internalResp.ChargeID = ledger.UUID.String()
		}
	}

	response.SendOpenApiResponseOK(ww, resp)
}

func (c *Controller) GetStoredCardByCustomerID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/creditcard/GetStoredCardByCustomerID")
	defer segment.End()

	var (
		err error
	)

	merchantID := chi.URLParam(r, "merchantId")
	if merchantID == "" {
		err = fmt.Errorf("merchantId required")
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	customerID := chi.URLParam(r, "customerId")
	if customerID == "" {
		err = fmt.Errorf("customerId required")
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	resp, err := c.creditcardSvc.GetStoredCardByCustomerID(ctx, merchantID, customerID)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
