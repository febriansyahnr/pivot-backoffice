package merchantTopUp

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *merchantTopUpService) FindByMerchantAccountNameAndPaymentMethodId(ctx context.Context, merchantId, accountName, paymentMethodId string) (*model.MerchantTopUp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantTopUp/FindByMerchantAccountNameAndPaymentMethodId")
	defer segment.End()

	merchantReference, err := s.merchantTopUpRepo.GetByMerchantAccountNameAndPaymentMethodId(ctx, merchantId, accountName, paymentMethodId)
	if err != nil {
		s.logger.Error(ctx, "failed to get merchant top up reference by merchant id, account name and payment method id", logger.Error(err))
		return nil, err

	} else if merchantReference == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, errors.New("merchant top up reference with merchant, account name and payment method is not found"))
	}
	return merchantReference, nil
}

func (s *merchantTopUpService) FindByReferenceNumber(ctx context.Context, referenceNumber string) (*model.MerchantTopUp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantTopUp/FindByReferenceNumber")
	defer segment.End()

	merchantReference, err := s.merchantTopUpRepo.GetByReferenceNumber(ctx, referenceNumber)
	if err != nil {
		s.logger.Error(ctx, "failed to get merchant top up reference by reference number", logger.Error(err))
		return nil, err

	} else if merchantReference == nil {
		return nil, constant.ErrMerchantTopUpReferenceNotFound
	}
	return merchantReference, nil
}
