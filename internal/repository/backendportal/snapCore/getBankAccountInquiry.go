package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) GetBankAccountInquiry(
	ctx context.Context, request snapCoreModel.InquiryAccountRequest) (*snapCoreModel.InquiryAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/GetBankAccountInquiry")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/bank-account/inquiry", r.config.SnapCoreConfig.BaseUrl)

	r.logger.Info(ctx, "GetBankAccountInquiry Request", logger.String("url", url), logger.Any("request", request))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
			constant.HeaderXMerchantID:         request.MerchantID,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request bank account inquiry", logger.Error(err))
		return nil, err
	}

	if statusCode >= 400 {
		var resp snapCoreModel.InquiryAccountResponse
		if jsonErr := json.Unmarshal(response, &resp); jsonErr == nil {
			errMsg := resp.Message
			if resp.Error != nil {
				if errMap, ok := resp.Error.(map[string]interface{}); ok {
					if msg, exists := errMap["message"]; exists {
						errMsg = fmt.Sprintf("%v", msg)
					}
				} else {
					errB, _ := json.Marshal(resp.Error)
					errMsg = string(errB)
				}
			}

			errType := mapSnapCoreResponseCodeToErrorType(resp.Code)
			if errType == "" {
				errType, _ = mapPartnerHTTPStatusToErrorType(statusCode)
			}

			err = pkgErrors.New(errType, errors.New(errMsg))
			r.logger.Error(ctx, fmt.Sprintf("got error %d when do request bank account inquiry, errorCode %s", statusCode, resp.Code), logger.Error(err))
			return &resp.Data, err
		}

		errType, _ := mapPartnerHTTPStatusToErrorType(statusCode)
		err = pkgErrors.New(errType, fmt.Errorf("partner returned HTTP %d", statusCode))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when do request bank account inquiry (body not parseable)", statusCode), logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.InquiryAccountResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read bank account inquiry response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "response from bank account inquiry", logger.String("body", string(response)))
	return &resp.Data, nil
}

// from model version processor routing
func (r *snapCoreRepository) BankAccountInquiry(ctx context.Context, request *routingProcessorModel.InquiryAccountRequest) (
	*routingProcessorModel.InquiryAccountResponseData,
	error,
) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/BankAccountInquiry")
	defer segment.End()

	resp, err := r.GetBankAccountInquiry(ctx, snapCoreModel.InquiryAccountRequest{
		BeneficiaryBankCode:    request.BeneficiaryBankCode,
		BeneficiaryAccountNo:   request.BeneficiaryAccountNo,
		BeneficiaryAccountName: request.BeneficiaryAccountName,
		PartnerReferenceNo:     request.PartnerReferenceNo,
		MerchantID:             request.MerchantID,
		AdditionalInfo:         request.AdditionalInfo,
	})

	if resp == nil {
		return nil, err
	}

	return &routingProcessorModel.InquiryAccountResponseData{
		ResponseCode:           resp.ResponseCode,
		ResponseMessage:        resp.ResponseMessage,
		PartnerReferenceNo:     resp.PartnerReferenceNo,
		BeneficiaryAccountName: resp.BeneficiaryAccountName,
		BeneficiaryAccountNo:   resp.BeneficiaryAccountNo,
		BeneficiaryBankCode:    resp.BeneficiaryBankCode,
		BeneficiaryBankName:    resp.BeneficiaryBankName,
		IsVirtualAccount:       resp.IsVirtualAccount,
	}, err
}
