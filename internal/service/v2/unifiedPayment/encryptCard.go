package unifiedPaymentService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func (s *UnifiedPaymentService) EncryptCard(ctx context.Context, request *unifiedPaymentModel.EncryptCardRequest) (*unifiedPaymentModel.EncryptedCardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/EncryptCard")
	defer segment.End()

	ffContext := ffcontext.NewEvaluationContext(s.config.Environment)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, request.MerchantID)
	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagKeyUnifiedPaymentCardEncryptionWhitelistedMerchant, ffContext, false)
	if !enabled {
		return nil, pkgErrors.New(response.HttpErrForbidden, constant.ErrCardEncryptionIsNotEnabledForMerchant)
	}

	data, err := s.creditCardProcessorRepo.EncryptCardData(ctx, request.ToCreditcardRequestModel())
	if err != nil {
		return nil, err
	}

	return unifiedPaymentModel.GetUnifiedEncryptedCardResponse(data), nil
}

func (s *UnifiedPaymentService) GetEncryptedCard(ctx context.Context, merchantId, cardId string) (*unifiedPaymentModel.EncryptedCardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetEncryptedCard")
	defer segment.End()

	ffContext := ffcontext.NewEvaluationContext(s.config.Environment)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)
	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagKeyUnifiedPaymentCardEncryptionWhitelistedMerchant, ffContext, false)
	if !enabled {
		return nil, pkgErrors.New(response.HttpErrForbidden, constant.ErrCardEncryptionIsNotEnabledForMerchant)
	}

	data, err := s.creditCardProcessorRepo.GetEncryptedCardData(ctx, merchantId, cardId)
	if err != nil {
		return nil, err
	}

	return unifiedPaymentModel.GetUnifiedEncryptedCardResponse(data), nil
}
