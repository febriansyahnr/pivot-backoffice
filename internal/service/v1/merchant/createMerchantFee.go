package merchant

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) CreateMerchantFee(ctx context.Context, request *merchant.NewMerchantFeeRequest) (*merchant.MerchantFeeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/CreateMerchantFee")
	defer segment.End()

	merchantDetail, err := s.repo.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to find merchant", logger.Error(err))
		return nil, err

	} else if merchantDetail == nil {
		s.logger.Warn(ctx, "merchant not found", logger.String("MerchantID", request.MerchantID))
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)

	} else if merchantDetail.ParentID.Valid && merchantDetail.KYCStatus.String != constant.KYCStatusApproved {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("merchant fees can't be set to non kyc sub-merchant"))

	} else if slices.Contains([]string{constant.ReferencePlatformActivity, constant.ReferencePlatformTransfer}, request.Reference) && merchantDetail.ParentID.Valid {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("this reference is not intended for sub-merchant"))
	}

	merchantFee, err := merchant.NewMerchantFee(request)
	if err != nil {
		return nil, err
	}
	if request.Reference == constant.ReferencePayment && request.PaymentMethod != constant.ChannelCreditCard && request.Channel != "" {
		exists, err := s.paymentMethod.GetPaymentMethodByCategoryTypeAndAcquirer(ctx, request.Reference, request.PaymentMethod, request.Channel)
		if err != nil {
			return nil, pkgErrs.New(response.HttpErrDatabase, err)

		} else if exists == nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("channel not found in payment methods list"))
		}
	}

	getRequest := &merchant.GetMerchantFeeRequest{
		MerchantID:    request.MerchantID,
		Reference:     request.Reference,
		ReferenceType: request.ReferenceType,
		PaymentMethod: request.PaymentMethod,
		Channel:       request.Channel,
	}
	if merchantFee.Channel != nil && *merchantFee.Channel != "" {
		getRequest.Channel = *merchantFee.Channel
	}
	if merchantFee.SettlementModel != nil && *merchantFee.SettlementModel != "" {
		getRequest.SettlementModel = *merchantFee.SettlementModel
	}
	if merchantFee.SettlementMethod != nil && *merchantFee.SettlementMethod != "" {
		getRequest.SettlementMethod = *merchantFee.SettlementMethod
	}
	if existedFee, _ := s.repo.GetMerchantFeeByRequest(ctx, getRequest); existedFee != nil {
		return nil, pkgErrs.New(
			response.HttpErrConflict, fmt.Errorf("the merchant fee already exists. please use the update feature. merchant fee id %s", existedFee.UUID),
		)
	}

	if err = s.repo.CreateMerchantFee(ctx, merchantFee); err != nil {
		s.logger.Error(ctx, "failed to create merchant fee", logger.Error(err))
		return nil, err
	}
	if merchantFee.Reference == constant.ReferenceDisbursement && merchantFee.Channel != nil {
		_ = s.redis.Del(ctx, fmt.Sprintf(constant.CacheKeyFmtPayoutTransactionFee, merchantFee.MerchantID, strings.ToLower(*merchantFee.Channel)))
	}
	return merchantFee.ToResponse(), nil
}
