package internalPaymentController

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (c *InternalPaymentController) createPaymentViaUnifiedV2(ctx context.Context, merchantID string, snapRequest paymentModel.PaymentRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/createPaymentViaUnifiedV2")
	defer segment.End()

	if c.unifiedPaymentSvc == nil {
		c.logger.Error(ctx, "Unified Payment V2 service is not initialized")
		return nil, pkgErrors.New(response.HttpErrInternal, fmt.Errorf("unified Payment V2 service unavailable"))
	}

	unifiedRequest := c.translateSnapToUnifiedRequest(snapRequest, merchantID)

	unifiedResp, err := c.unifiedPaymentSvc.CreateSession(ctx, unifiedRequest)
	if err != nil {
		c.logger.Error(ctx, "Failed to create payment via Unified Payment V2", logger.Error(err))
		return nil, err
	}

	snapResponse := c.translateUnifiedToSnapResponse(unifiedResp, snapRequest)

	return snapResponse, nil
}

func (c *InternalPaymentController) translateSnapToUnifiedRequest(snapRequest paymentModel.PaymentRequest, merchantID string) *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest {
	unifiedRequest := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		ClientReferenceID: snapRequest.ReferenceID,
		PaymentType:       constant.UnifiedPaymentTypeMultiple,
		Amount: unifiedPaymentModel.Amount{
			Value:    snapRequest.TotalAmount.Value.InexactFloat64(),
			Currency: snapRequest.TotalAmount.Currency,
		},
		Mode:              constant.UnifiedPaymentModeRedirect,
		AutoConfirm:       true,
		PaymentID:         snapRequest.UUID,
		MerchantID:        merchantID,
		CreatedBy:         snapRequest.CreatedBy,
		IsMigratingFromV1: true,
		IsSnap:            true,
	}

	if snapRequest.PaymentMethod != "" {
		unifiedRequest.PaymentMethod = &unifiedPaymentModel.PaymentMethod{
			Type: snapRequest.PaymentMethod,
		}

		unifiedRequest.PaymentMethodOptions = unifiedPaymentModel.PaymentMethodOptions{}

		switch snapRequest.PaymentMethod {
		case constant.ChannelVirtualAccount:
			if snapRequest.VirtualAccount != nil {
				unifiedRequest.PaymentMethodOptions.VirtualAccount = &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
					Channel:               snapRequest.VirtualAccount.Issuer,
					VirtualAccountTrxType: snapRequest.VirtualAccount.VirtualAccountTrxType,
					VirtualAccountName:    snapRequest.VirtualAccount.VirtualAccountName,
					BillDetails:           snapRequest.VirtualAccount.BillDetails,
				}
			}
		case constant.ChannelQris:
			if snapRequest.Qris != nil {
				unifiedRequest.PaymentMethod.Type = "QR"
				unifiedRequest.PaymentMethodOptions.QR = &unifiedPaymentModel.PaymentMethodOptionQR{}

				if unifiedRequest.ClientReferenceID == "" {
					unifiedRequest.ClientReferenceID = fmt.Sprintf("QR%d", time.Now().Unix())
				}
			}
		}
	}

	if snapRequest.Customer.CustomerID != "" {
		unifiedRequest.CustomerID = snapRequest.Customer.CustomerID
		unifiedRequest.CustomerInformation = &unifiedPaymentModel.CustomerInformation{
			Email: snapRequest.Customer.Email,
		}
	}

	unifiedRequest.RedirectUrl = unifiedPaymentModel.RedirectUrl{
		SuccessReturnUrl:    snapRequest.ClientRedirectUrl.SuccessUrl,
		FailureReturnUrl:    snapRequest.ClientRedirectUrl.FailedUrl,
		ExpirationReturnUrl: "",
	}

	unifiedRequest.SplitRoutingConfigurations = snapRequest.SplitRoutingConfigurations

	return unifiedRequest
}

func (c *InternalPaymentController) translateUnifiedToSnapResponse(unifiedResp *unifiedPaymentModel.UnifiedPaymentSessionResponse, snapRequest paymentModel.PaymentRequest) *paymentModel.PaymentResponse {

	status := unifiedResp.Status
	if status == constant.StatusActive {
		status = constant.StatusPending
	}

	snapResponse := &paymentModel.PaymentResponse{
		UUID:        unifiedResp.ID,
		ReferenceID: unifiedResp.ClientReferenceID,
		Status:      status,
		Customer: &paymentModel.PaymentRequestCustomer{
			CustomerID: snapRequest.Customer.CustomerID,
			Name:       snapRequest.Customer.Name,
			Email:      snapRequest.Customer.Email,
			Phone:      snapRequest.Customer.Phone,
		},
		PaymentMethod:    snapRequest.PaymentMethod,
		PaymentType:      constant.UnifiedPaymentTypeMultiple,
		IsUnifiedPayment: false,
		TotalAmount: &paymentModel.Amount{
			Value:    decimal.NewFromFloat(unifiedResp.Amount.Value),
			Currency: unifiedResp.Amount.Currency,
		},
	}

	if snapRequest.VirtualAccount != nil && unifiedResp.PaymentMethod != nil &&
		unifiedResp.PaymentMethod.VAPaymentMethodDetail != nil {

		vaDetail := unifiedResp.PaymentMethod.VAPaymentMethodDetail
		metadata := make(map[string]any)
		if c.config != nil && c.config.Environment == constant.EnvironmentStaging && unifiedResp.ID != "" {
			metadata[constant.PaymentSimulatorKey] = fmt.Sprintf(
				c.config.MerchantPortalConfig.PaymentSimulationPatternURL,
				base64.StdEncoding.EncodeToString([]byte(unifiedResp.ID)),
			)
		}

		snapResponse.VirtualAccount = &paymentModel.PaymentVirtualAccountResponse{
			Issuer:                snapRequest.VirtualAccount.Issuer,
			VirtualAccountTrxType: snapRequest.VirtualAccount.VirtualAccountTrxType,
			VirtualAccountNumber:  vaDetail.VirtualAccountNumber,
			VirtualAccountName:    vaDetail.VirtualAccountName,
			MinAmount:             snapRequest.VirtualAccount.MinAmount,
			MaxAmount:             snapRequest.VirtualAccount.MaxAmount,
			ExpiredDate:           snapRequest.VirtualAccount.ExpiredDate,
			Metadata:              metadata,
		}
	}

	if snapRequest.Qris != nil && unifiedResp.PaymentMethod != nil &&
		unifiedResp.PaymentMethod.QrPaymentMethodDetail != nil {

		qrDetail := unifiedResp.PaymentMethod.QrPaymentMethodDetail
		qrMetadata := make(map[string]any)
		if c.config != nil && c.config.Environment == constant.EnvironmentStaging && unifiedResp.ID != "" {
			qrMetadata[constant.PaymentSimulatorKey] = fmt.Sprintf(
				c.config.MerchantPortalConfig.PaymentSimulationPatternURL,
				base64.StdEncoding.EncodeToString([]byte(unifiedResp.ID)),
			)
		}

		snapResponse.Qris = &paymentModel.PaymentQrisResponse{
			QrContent:    qrDetail.QrContent,
			QrUrl:        qrDetail.QrUrl,
			MerchantName: qrDetail.MerchantName,
			QrType:       qrDetail.QrType,
			Metadata:     qrMetadata,
		}
	}

	if snapRequest.PaymentItems != nil {
		items := make([]paymentModel.PaymentResponseItem, len(*snapRequest.PaymentItems))
		for i, item := range *snapRequest.PaymentItems {
			items[i] = paymentModel.PaymentResponseItem{
				Name:        item.Name,
				Description: item.Description,
				Qty:         item.Qty,
				Amount:      item.Amount,
			}
		}
		snapResponse.PaymentItems = &items
	}

	return snapResponse
}
