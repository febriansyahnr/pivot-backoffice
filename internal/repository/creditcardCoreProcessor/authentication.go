package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *creditcardCoreProcessorRepository) CreateEncryptedCardAuthenticationLink(ctx context.Context, request *creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest) (*creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/CreateEncryptedCardAuthenticationLink")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/internal/credit-cards/encrypt-card/authentication", r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "Encrypt card authentication", logger.String("url", url))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			"X-MERCHANT-ID":                    request.MerchantID,
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
			constant.HeaderXEncryptHashKey:     request.EncryptedEncryptionKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request encrypt card authentication", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.BaseEncryptedCardAuthenticationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when unmarshal encrypt card authentication response", logger.Error(err))
		return nil, err
	}
	r.logger.Info(ctx, "Encrypt Card Authentication", logger.ByteString("response", response))

	if resp.Message != "" && statusCode != http.StatusOK {
		r.logger.Error(ctx, "error when encrypt card authentication", logger.Any("response", resp), logger.String("error_code", resp.Code))
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Message))
		return nil, err
	}
	// need to attention about the err resp message
	// "invalid card number" & ""
	// the value used to mapping the error information in pkg/util/response/http_response.go
	if statusCode >= 400 {
		switch resp.Message {
		case "decryption key used does not match",
			"failed while decrypting card encryption key",
			"failed while decrypting card details",
			"failed to parse card details in JSON format":
			err = pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrFailedToDecryptCardPayment)
		case "invalid card number format", "invalid card number":
			err = pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInvalidCardNumber)
		case "billing information is required for foreign card":
			err = pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrForeignCardBillingInformationMissing)
		case "foreign cards are not allowed for this merchant":
			err = pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrForeignCardNotAllowed)
		case "the card principal is not supported for this merchant":
			err = pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrUnsupportedCardPrincipal)
		case "invalid network token number",
			"invalid network token cryptogram",
			"invalid network token requestor id",
			"network token card on file with customer initiator missing cryptogram",
			"network token card on file with merchant initiator contain cryptogram",
			"network token card on file with merchant initiator contain token requestor id",
			"network token card on file with customer initiator missing token requestor id",
			"unable to specify both card and network token number",
			"not allowed to specify network token number and cvc":
			err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(resp.Message))
		case "there is ongoing network token transaction for same payment id":
			err = pkgErrors.New(httpResponse.HttpErrUnprocessableContent, errors.New(resp.Message))
		default:
			err = pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInvalidCardInformation)
		}

		if strings.Contains(resp.Message, constant.ErrMsgForeignCardBillingInformationValidation) {
			attributeDetailStr := strings.Replace(resp.Message, constant.ErrMsgForeignCardBillingInformationValidation, "", 1)
			if strings.Contains(attributeDetailStr, "First Name") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.givenName is required"))
			} else if strings.Contains(attributeDetailStr, "Last Name") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.surname is required"))
			} else if strings.Contains(attributeDetailStr, "Address Line") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.addressLine is required"))
			} else if strings.Contains(attributeDetailStr, "City") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.city is required"))
			} else if strings.Contains(attributeDetailStr, "State/Province") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.provinceState is required"))
			} else if strings.Contains(attributeDetailStr, "Postal Code") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.postalCode is required"))
			} else if strings.Contains(attributeDetailStr, "Country") {
				err = pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.country is required"))
			}
		}

		r.logger.Error(ctx, "got error when request encrypted card", logger.Error(err), logger.String("reference_id", request.ClientTransactionID))

		return nil, err
	}

	return &resp.Data, nil

}

func (r *creditcardCoreProcessorRepository) Authentication(ctx context.Context, request creditcardCoreProcessorModel.AuthenticationRequest) (_ *creditcardCoreProcessorModel.AuthenticationResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/creditcardCoreProcessor/Authentication")
	defer segment.End()

	url := r.config.CreditcardCoreProcessorConfig.BaseUrl + "/api/v1/credit-card/authentication"
	headers := map[string]string{
		constant.HeaderContentType:     "text/plain",
		constant.HeaderXIdempotencyKey: request.PaymentID,
		constant.HeaderXMerchantID:     request.MerchantID,
	}
	defer func() {
		if err == nil || strings.Contains(err.Error(), "unmarshalling json") {
			return
		}
		r.logger.Info(ctx, "Failed to send authentication request to card processor", logger.String("url", url), logger.Any("headers", headers), logger.String("body", request.EncryptedPayload), logger.Error(err))
	}()
	responseBody, statusCode, err := r.httpRequest.POST(ctx, url, []byte(request.EncryptedPayload), headers)
	if err != nil {
		r.logger.Error(ctx, "Failed while performing authentication request (encrypted card web view flow)", logger.Error(err))
		return nil, err
	}

	if statusCode >= 400 {
		r.logger.Error(ctx, fmt.Sprintf("Failed while performing authentication request (encrypted card web view flow), received http code %d", statusCode), logger.ByteString("responseBody", responseBody))
		if statusCode >= 500 {
			return nil, fmt.Errorf("partner response: %s", responseBody)
		}
		return nil, pkgErrors.NewNonRetryableError(fmt.Errorf("partner response: %s", responseBody))
	}

	response := creditcardCoreProcessorModel.Response[creditcardCoreProcessorModel.AuthenticationResponse]{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		r.logger.Error(ctx, "Failed while unmarshalling authentication response", logger.Error(err))
		return nil, pkgErrors.NewNonRetryableError(fmt.Errorf("unmarshalling json: %w", err))
	}
	return response.Data, nil
}
