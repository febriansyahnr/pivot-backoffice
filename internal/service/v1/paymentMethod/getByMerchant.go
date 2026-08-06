package paymentMethodService

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *PaymentMethodService) GetPaymentMethodByMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) ([]*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/GetPaymentMethodByMerchant")
	defer segment.End()

	// Get Payment Method by category
	paymentMethods, err := s.paymentMethodRepo.GetListPaymentMethodMerchant(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method merchant", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// this path used for unified payment
	// we need to validate eligible payment method by the amount and expiry time
	if filter.Payment != nil {
		var validPaymentMethods = []*paymentModel.PaymentMethodWithPivot{}

		for _, paymentMethod := range paymentMethods {
			if s.isValidPaymentMethodByPaymentDetail(ctx, filter.Payment, paymentMethod) {
				validPaymentMethods = append(validPaymentMethods, paymentMethod)
			}
		}
		return validPaymentMethods, nil
	}

	return paymentMethods, nil
}

func (s *PaymentMethodService) GetStaticVAPaymentMethodByMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) ([]*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/GetPaymentMethodByMerchant")
	defer segment.End()

	filter.Category = constant.TypePayment
	filter.Type = paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT
	paymentMethods, err := s.paymentMethodRepo.GetListPaymentMethodMerchant(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method merchant", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Get default VA config from snap core with caching (1 day TTL)
	defaultVAConfigMerchantID := s.config.UnifiedPaymentConfig.VirtualAccountConfig.DefaultVaConfigMerchantId
	cacheKey := fmt.Sprintf(constant.CacheKeyDefaultVAConfig, defaultVAConfigMerchantID, constant.PaymentMethodChannelTypeAggregator)

	var defaultVAConfigs []*snapCoreVAModel.VirtualAccountConfigResponseData

	// Try to get from cache first
	cachedData := s.redis.Get(ctx, cacheKey)
	if cachedData.Err() == nil {
		var cachedConfigs []*snapCoreVAModel.VirtualAccountConfigResponseData
		if err := json.Unmarshal([]byte(cachedData.Val()), &cachedConfigs); err == nil {
			defaultVAConfigs = cachedConfigs
			s.logger.Info(ctx, "using cached VA config from redis")
		}
	}

	// If not in cache, fetch from snap core
	if defaultVAConfigs == nil {
		vaConfigResponse, err := s.snapCoreRepo.GetVirtualAccountConfig(ctx, &snapCoreVAModel.GetVirtualAccountConfigRequest{
			MerchantID:      defaultVAConfigMerchantID,
			IntegrationType: constant.PaymentMethodChannelTypeAggregator,
			Status:          constant.VirtualAccountConfigStatusActive,
		})
		if err != nil {
			s.logger.Error(ctx, "error when get default VA config from snap core", logger.Error(err))
			return nil, pkgErrors.New(response.HttpErrDatabase, err)
		}

		defaultVAConfigs = vaConfigResponse

		// Cache the result for 1 day
		if s.redis != nil && defaultVAConfigs != nil {
			if cachedData, err := json.Marshal(defaultVAConfigs); err == nil {
				ttl := 24 * time.Hour // 1 day TTL
				s.redis.Set(ctx, cacheKey, string(cachedData), ttl)
				s.logger.Info(ctx, "cached VA config to redis for 1 day")
			}
		}
	}

	defaultVARangeStart := s.config.UnifiedPaymentConfig.VirtualAccountConfig.DefaultVaRangeStart
	defaultVARangeEnd := s.config.UnifiedPaymentConfig.VirtualAccountConfig.DefaultVaRangeEnd
	for _, paymentMethod := range paymentMethods {
		if !(paymentMethod.MerchantConfigObj != nil && paymentMethod.MerchantConfigObj.PartnerConfig != nil &&
			paymentMethod.MerchantConfigObj.PartnerConfig.VirtualAccount != nil) {

			vaConfigItems := []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{}
			// Find va config from defaultVAConfigs matched by acquirer, then loop into vaConfigItems
			for _, vaConfig := range defaultVAConfigs {
				if strings.EqualFold(vaConfig.Acquirer, paymentMethod.Acquirer) {
					vaConfigItem := paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
						BINPrefix:  vaConfig.BinPrefix,
						Type:       vaConfig.Type,
						StartRange: vaConfig.MetadataObj.MerchantPrefix.StartRange,
						EndRange:   vaConfig.MetadataObj.MerchantPrefix.EndRange,
					}
					if vaConfigItem.StartRange == "" {
						vaConfigItem.StartRange = fmt.Sprintf("%0*s", vaConfig.BinMin, defaultVARangeStart)
					}
					if vaConfigItem.EndRange == "" {
						vaConfigItem.EndRange = fmt.Sprintf("%0*s", vaConfig.BinMin, defaultVARangeEnd)
					}

					vaConfigItems = append(vaConfigItems, vaConfigItem)
				}
			}

			paymentMethod.MerchantConfigObj = &paymentModel.PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
						Items: vaConfigItems,
					},
				},
			}
		}
	}

	return paymentMethods, nil
}

// isValidPaymentMethodByPaymentDetail determines whether the given payment method configuration
// is applicable to the provided payment detail.
func (s *PaymentMethodService) isValidPaymentMethodByPaymentDetail(ctx context.Context, payment *paymentModel.PaymentDetailForPaymentUIResponse, paymentMethodCfg *paymentModel.PaymentMethodWithPivot) bool {
	if payment == nil || paymentMethodCfg == nil {
		return false
	}

	amount, err := decimal.NewFromString(payment.Amount.Value)
	if err != nil {
		s.logger.Error(ctx, "failed to parse payment amount", logger.String("paymentID", payment.UUID), logger.String("amount", payment.Amount.Value), logger.Error(err))
		return false
	}

	// validate payment method for supported expiry time
	// waiting another config update, current situation just sample and the limit is 3 month
	expiryLimit := s.getPaymentMethodExpiryLimit(ctx, paymentMethodCfg)
	if payment.ExpiredAt != nil && !expiryLimit.IsZero() && payment.ExpiredAt.After(expiryLimit) {
		s.logger.Warn(ctx, "payment expiry time not supported by payment method",
			logger.String("name", paymentMethodCfg.PaymentMethod.Name),
			logger.String("paymentID", payment.UUID),
			logger.String("expiry", payment.ExpiredAt.String()),
			logger.String("limit", expiryLimit.String()),
		)
		return false
	}

	// validate payment method for supported amount
	cfg, ok := s.paymentMethodValidationConfig[paymentMethodCfg.PaymentMethod.Type]
	if ok && cfg != nil && (amount.GreaterThan(decimal.NewFromFloat(cfg.MaxAmount)) || amount.LessThan(decimal.NewFromFloat(cfg.MinAmount))) {
		s.logger.Warn(ctx, "payment amount not supported by payment method",
			logger.String("type", paymentMethodCfg.PaymentMethod.Type),
			logger.String("paymentID", payment.UUID),
			logger.String("amount", payment.Amount.Value),
			logger.Float64("maxLimit", cfg.MaxAmount),
			logger.Float64("minLimit", cfg.MinAmount),
		)
		return false
	}

	return true
}

// getPaymentMethodExpiryLimit computes the expiry cutoff time for a payment method by combining
// service-level validation defaults and the payment-method's own expiry configuration.
//
// Precedence:
//  1. paymentMethodCfg.ConfigObj.ExpiryConfig (per-method/per-merchant configuration) — highest precedence.
//  2. s.paymentMethodValidationConfig[paymentMethodCfg.PaymentMethod.Type] (service-level defaults).
//
// For either source to be considered valid, both Duration must be > 0 and Unit must be non-empty.
// If neither source provides a valid expiry configuration, the function returns the zero time (time.Time{}).
//
// The selected expiry configuration is converted to an absolute time via expiryValidationCfg.ToDateTime().
func (s *PaymentMethodService) getPaymentMethodExpiryLimit(_ context.Context, paymentMethodCfg *paymentModel.PaymentMethodWithPivot) time.Time {
	var (
		expiryValidationCfg = paymentModel.PaymentMethodExpiryConfig{}
	)
	cfg, ok := s.paymentMethodValidationConfig[paymentMethodCfg.PaymentMethod.Type]
	if ok && cfg != nil && cfg.MaxExpiryDuration > 0 && cfg.MaxExpiryDurationUnit != "" {
		expiryValidationCfg.Duration = cfg.MaxExpiryDuration
		expiryValidationCfg.Unit = cfg.MaxExpiryDurationUnit
	}

	if util.ValueOfPtr(paymentMethodCfg.ConfigObj).ExpiryConfig.Duration > 0 && util.ValueOfPtr(paymentMethodCfg.ConfigObj).ExpiryConfig.Unit != "" {
		expiryValidationCfg.Duration = paymentMethodCfg.ConfigObj.ExpiryConfig.Duration
		expiryValidationCfg.Unit = paymentMethodCfg.ConfigObj.ExpiryConfig.Unit
	}

	if expiryValidationCfg.Duration == 0 && expiryValidationCfg.Unit == "" {
		return time.Time{}
	}

	return expiryValidationCfg.ToDateTime()
}
