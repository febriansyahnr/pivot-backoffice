package unifiedPaymentService

import (
	"context"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) GetPaymentMethodConfig(ctx context.Context, merchantId string) (*unifiedPaymentModel.GetPaymentMethodConfigResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetPaymentMethodConfig")
	defer span.End()

	if derivedMerchantID, _ := ctx.Value(constant.CtxDerivedMerchantID).(string); derivedMerchantID != "" {
		merchantId = derivedMerchantID
	}

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "failed to get merchant data", logger.String("merchantID", merchantId))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrFindParentMerchant)
	}

	if merchant == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	// non-kyc merchant able to get the payment method config directly
	if merchant.KYCStatus.String != constant.KYCStatusApproved {
		merchantId = merchant.ParentID.String
	}

	defaultMaxExpiryVa := ""
	if s.config.UnifiedPaymentConfig.VirtualAccountConfig != nil {
		defaultMaxExpiryVa = FormatMaxExpiryDuration(
			s.config.UnifiedPaymentConfig.VirtualAccountConfig.MaxExpiryDuration,
			s.config.UnifiedPaymentConfig.VirtualAccountConfig.MaxExpiryDurationUnit,
		)
	}

	defaultMaxExpiryQr := ""
	if s.config.UnifiedPaymentConfig.QrConfig != nil {
		defaultMaxExpiryQr = FormatMaxExpiryDuration(
			s.config.UnifiedPaymentConfig.QrConfig.MaxExpiryDuration,
			s.config.UnifiedPaymentConfig.QrConfig.MaxExpiryDurationUnit,
		)
	}

	defaultMaxExpiryCard := ""
	if s.config.UnifiedPaymentConfig.CardConfig != nil {
		defaultMaxExpiryCard = FormatMaxExpiryDuration(
			s.config.UnifiedPaymentConfig.CardConfig.MaxExpiryDuration,
			s.config.UnifiedPaymentConfig.CardConfig.MaxExpiryDurationUnit,
		)
	}

	defaultMaxExpiryEwallet := ""
	if s.config.UnifiedPaymentConfig.EwalletConfig != nil {
		defaultMaxExpiryEwallet = FormatMaxExpiryDuration(
			s.config.UnifiedPaymentConfig.EwalletConfig.MaxExpiryDuration,
			s.config.UnifiedPaymentConfig.EwalletConfig.MaxExpiryDurationUnit,
		)
	}

	resp := &unifiedPaymentModel.GetPaymentMethodConfigResponse{
		Card: &unifiedPaymentModel.GetPaymentMethodConfigCard{
			Enabled:          false,
			AcceptedChannels: make([]string, 0),
			MaximumExpiry:    defaultMaxExpiryCard,
		},
		VirtualAccount: &unifiedPaymentModel.GetPaymentMethodConfigVirtualAccount{
			Enabled:          false,
			AcceptedChannels: make([]string, 0),
			MaximumExpiry:    defaultMaxExpiryVa,
		},
		Qr: &unifiedPaymentModel.GetPaymentMethodConfigQr{
			Enabled:       false,
			MaximumExpiry: defaultMaxExpiryQr,
		},
		Ewallet: &unifiedPaymentModel.GetPaymentMethodConfigEWallet{
			Enabled:       false,
			MaximumExpiry: defaultMaxExpiryEwallet,
		},
		Installment: &unifiedPaymentModel.GetPaymentMethodConfigInstallment{
			Enabled:          false,
			AcceptedChannels: make(map[string][]*unifiedPaymentModel.InstallmentAcceptedChannels),
		},
	}

	var minEwalletDuration int
	var minEwalletUnit string

	// Get payment method activation
	paymentMethods, err := s.paymentMethodRepo.GetListPaymentMethodMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantId,
		Category:   paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
	})
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	for _, paymentMethod := range paymentMethods {
		switch paymentMethod.Type {
		case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
			if !resp.VirtualAccount.Enabled && paymentMethod.IsActive {
				resp.VirtualAccount.Enabled = true
			}

			if paymentMethod.IsActive {
				resp.VirtualAccount.AcceptedChannels = append(resp.VirtualAccount.AcceptedChannels, strings.ToUpper(paymentMethod.Acquirer))
			}

		case paymentConstant.PAYMENT_METHOD_QRIS:
			if !resp.Qr.Enabled && paymentMethod.IsActive {
				resp.Qr.Enabled = true
			}

		case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
			if !resp.Card.Enabled && paymentMethod.IsActive {
				resp.Card.Enabled = true
			}

			if paymentMethod.IsActive && paymentMethod.ConfigObj != nil && paymentMethod.ConfigObj.ExpiryConfig.Unit != "" && paymentMethod.ConfigObj.ExpiryConfig.Duration > 0 {
				resp.Card.MaximumExpiry = FormatMaxExpiryDuration(
					paymentMethod.ConfigObj.ExpiryConfig.Duration,
					paymentMethod.ConfigObj.ExpiryConfig.Unit,
				)
			}

		case paymentConstant.PAYMENT_METHOD_EWALLET:
			if !resp.Ewallet.Enabled && paymentMethod.IsActive {
				resp.Ewallet.Enabled = true
			}

			if paymentMethod.IsActive {
				resp.Ewallet.AcceptedChannels = append(resp.Ewallet.AcceptedChannels, strings.ToUpper(paymentMethod.Acquirer))
			}

			if paymentMethod.IsActive && paymentMethod.ConfigObj != nil && paymentMethod.ConfigObj.ExpiryConfig.Unit != "" && paymentMethod.ConfigObj.ExpiryConfig.Duration > 0 {
				// Set to the minimum value
				if minEwalletDuration == 0 || isLessDuration(paymentMethod.ConfigObj.ExpiryConfig.Duration, paymentMethod.ConfigObj.ExpiryConfig.Unit, minEwalletDuration, minEwalletUnit) {
					minEwalletDuration = paymentMethod.ConfigObj.ExpiryConfig.Duration
					minEwalletUnit = paymentMethod.ConfigObj.ExpiryConfig.Unit
				}
			}

		case paymentConstant.PAYMENT_METHOD_INSTALLMENT:
			var installmentPlanIds []string
			if paymentMethod.IsActive &&
				paymentMethod.MerchantConfigObj != nil &&
				paymentMethod.MerchantConfigObj.PartnerConfig != nil &&
				paymentMethod.MerchantConfigObj.PartnerConfig.Installment != nil &&
				len(paymentMethod.MerchantConfigObj.PartnerConfig.Installment.InstallmentPlanIDs) > 0 {

				resp.Installment.Enabled = true
				installmentPlanIds = append(installmentPlanIds, paymentMethod.MerchantConfigObj.PartnerConfig.Installment.InstallmentPlanIDs...)
				resp.Installment.AcceptedChannels[paymentMethod.Acquirer] = []*unifiedPaymentModel.InstallmentAcceptedChannels{}
			}

			if len(installmentPlanIds) > 0 {
				plans, _, err := s.installmentPlanSvc.List(ctx, &installmentPlanModel.FilterRequest{
					InstallmentIDs: installmentPlanIds,
					Acquirer:       paymentMethod.Acquirer,
					Status:         constant.InstallmentPlanStatusActive,
				})
				if err != nil {
					return nil, err
				}
				for _, plan := range plans {
					resp.Installment.AcceptedChannels[paymentMethod.Acquirer] = append(resp.Installment.AcceptedChannels[paymentMethod.Acquirer], &unifiedPaymentModel.InstallmentAcceptedChannels{
						ID:          plan.UUID,
						ProgramName: plan.Title,
						AllowedBins: plan.PlanMetadata.Card.AllowedBins,
						Tenure:      plan.GetStringTenor(),
						Interest:    plan.PlanMetadata.Card.Interest,
						MinimumAmount: commonModel.Amount2{
							Currency: constant.CurrencyIDR,
							Value:    plan.PlanMetadata.Card.MinimumAmount,
						},
						MaximumAmount: commonModel.Amount2{
							Currency: constant.CurrencyIDR,
							Value:    plan.PlanMetadata.Card.MaximumAmount,
						},
					})
				}
			}
		}
	}

	// Set minimum expiry for ewallet if we have configurations
	if minEwalletDuration > 0 {
		resp.Ewallet.MaximumExpiry = FormatMaxExpiryDuration(minEwalletDuration, minEwalletUnit)
	}

	if s.config.UnifiedPaymentConfig.CardConfig != nil {
		if s.config.UnifiedPaymentConfig.CardConfig.MinAmount != nil {
			resp.Card.MinimumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.CardConfig.MinAmount,
			}
		}
		if s.config.UnifiedPaymentConfig.CardConfig.MaxAmount != nil {
			resp.Card.MaximumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.CardConfig.MaxAmount,
			}
		}

		resp.Card.AcceptedChannels = s.config.UnifiedPaymentConfig.CardConfig.AcceptedChannels
	}

	if s.config.UnifiedPaymentConfig.VirtualAccountConfig != nil {
		if s.config.UnifiedPaymentConfig.VirtualAccountConfig.MinAmount != nil {
			resp.VirtualAccount.MinimumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.VirtualAccountConfig.MinAmount,
			}
		}
		if s.config.UnifiedPaymentConfig.VirtualAccountConfig.MaxAmount != nil {
			resp.VirtualAccount.MaximumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.VirtualAccountConfig.MaxAmount,
			}
		}
	}

	if s.config.UnifiedPaymentConfig.QrConfig != nil {
		if s.config.UnifiedPaymentConfig.QrConfig.MinAmount != nil {
			resp.Qr.MinimumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.QrConfig.MinAmount,
			}
		}
		if s.config.UnifiedPaymentConfig.QrConfig.MaxAmount != nil {
			resp.Qr.MaximumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.QrConfig.MaxAmount,
			}
		}
	}

	if s.config.UnifiedPaymentConfig.EwalletConfig != nil {
		if s.config.UnifiedPaymentConfig.EwalletConfig.MinAmount != nil {
			resp.Ewallet.MinimumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.EwalletConfig.MinAmount,
			}
		}
		if s.config.UnifiedPaymentConfig.EwalletConfig.MaxAmount != nil {
			resp.Ewallet.MaximumAmount = &unifiedPaymentModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    *s.config.UnifiedPaymentConfig.EwalletConfig.MaxAmount,
			}
		}
	}

	return resp, nil
}

func FormatMaxExpiryDuration(duration int, unit string) string {
	return fmt.Sprintf("%d %s", duration, unit)
}

func isLessDuration(duration1 int, unit1 string, duration2 int, unit2 string) bool {
	return durationToSeconds(duration1, unit1) < durationToSeconds(duration2, unit2)
}

func durationToSeconds(duration int, unit string) int64 {
	switch strings.ToUpper(unit) {
	case "SECONDS":
		return int64(duration)
	case "MINUTES":
		return int64(duration * 60)
	case "HOURS":
		return int64(duration * 3600)
	case "DAYS":
		return int64(duration * 86400)
	default:
		return 0
	}
}
