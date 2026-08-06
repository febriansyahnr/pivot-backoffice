package feeService

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

var (
	defDeductionDay                  = int16(31)
	defPlatformTransactionPercentage = float64(0.05)
	defPlatformTransactionMaxAmount  = float64(20000)
)

func (s *FeeService) GetFeeCalculationAndDetail(ctx context.Context, request *feeModel.GetFeeRequest) (fee float64, detail *feeModel.FeeMetadataObject, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/GetFeeCalculationAndDetail")
	defer segment.End()

	var (
		feeAmount   = defaultFeeAmount(request.Reference)
		taxAmount   = 0.0
		merchantFee *merchantModel.MerchantFee
	)

	defaultFeeDetail := &feeModel.FeeMetadataObject{
		Type:            request.Reference,
		ReferenceType:   request.ReferenceType,
		Method:          "",
		Channel:         request.Channel,
		DeductionType:   constant.MerchantFeeDeductionTypeDirect,
		TrxAmount:       request.ReferenceAmount,
		AmountType:      constant.MerchantFeeAmountType,
		Amount:          feeAmount,
		TaxType:         constant.MerchantTaxTypeNonPKP,
		TaxAmount:       taxAmount,
		IsDefaultConfig: true,
		FinalAmount:     feeAmount,
	}
	// Default configuration other than fee amount
	if request.Reference == constant.ReferencePlatformActivity {
		defaultFeeDetail.DeductionDay = &defDeductionDay
		defaultFeeDetail.DeductionType = constant.MerchantFeeDeductionTypeAutomated
	}

	switch request.Reference {
	case constant.TypePayment:
		merchantFee, err = s.getPaymentMerchantFee(ctx, request)

		defaultFeeDetail.Method = request.PaymentMethod
		if merchantFee == nil {
			feeAmountFromDefault, amountType, amount, percentage := defaultFeeAmountForPaymentUsecase(request)

			feeAmount = feeAmountFromDefault
			defaultFeeDetail.AmountType = amountType
			defaultFeeDetail.Amount = amount
			defaultFeeDetail.Percentage = percentage
			defaultFeeDetail.FinalAmount = feeAmount
		}

	case constant.ReferencePaymentFundedPayout:
		merchantFee, err = s.getPaymentFundedPayoutMerchantFee(ctx, request)

		defaultFeeDetail.Method = request.PaymentMethod
		if merchantFee == nil {
			feeAmountFromDefault, amountType, amount, percentage := defaultFeeAmountForPaymentFundedPayoutUsecase(request.SettlementMethod, request.ReferenceAmount)

			feeAmount = feeAmountFromDefault
			defaultFeeDetail.AmountType = amountType
			defaultFeeDetail.Amount = amount
			defaultFeeDetail.Percentage = percentage
			defaultFeeDetail.FinalAmount = feeAmount
		}

	case constant.ReferencePlatformTransaction:
		merchantFee, err = s.getDefaultMerchantFee(ctx, request)
		if merchantFee == nil {
			defaultFeeDetail.DeductionType = constant.MerchantFeeDeductionTypeDirect
			defaultFeeDetail.AmountType = constant.MerchantFeePercentageType
			defaultFeeDetail.Percentage = defPlatformTransactionPercentage
			defaultFeeDetail.MaxFeeAmount = &defPlatformTransactionMaxAmount

			feeAmount, _ = s.CalculateFee(ctx, request, defaultFeeDetail)
			defaultFeeDetail.FinalAmount = feeAmount
		}
	case constant.ReferenceWallet:
		merchantFee, err = s.merchantRepo.GetMerchantFeeByRequest(ctx, &merchantModel.GetMerchantFeeRequest{
			MerchantID:    request.MerchantID,
			Reference:     constant.TypeWallet,
			ReferenceType: request.ReferenceType,
		})

	case constant.TypeRefund:
		merchantFee, err = s.getRefundMerchantFee(ctx, request)

		defaultFeeDetail.Method = request.PaymentMethod
		if merchantFee == nil {
			feeForRefund := 0.0
			if request.ReferenceType == constant.RefundDestinationTypeAccount {
				feeForRefund = 2000.0
			}

			feeAmount = feeForRefund
			defaultFeeDetail.AmountType = constant.MerchantFeeAmountType
			defaultFeeDetail.Amount = feeForRefund
			defaultFeeDetail.Percentage = 0
			defaultFeeDetail.FinalAmount = feeForRefund
		}

	case constant.ReferenceDisbursement:
		merchantFee, err = s.getPayoutMerchantFee(ctx, request)
	case constant.ReferenceDisbursementVA:
		merchantFee, err = s.getPayoutVAMerchantFee(ctx, request)
	case constant.TypeXB:
		merchantFee, err = s.getMerchantFeeXB(ctx, request)
	case constant.ReferenceTopUp:
		merchantFee, err = s.merchantRepo.DetermineTopupFeeByMerchantIdMethodAndChannel(ctx, request.MerchantID, request.PaymentMethod, request.Channel)
	default:
		merchantFee, err = s.getDefaultMerchantFee(ctx, request)
	}

	// Handle error and empty merchant fee
	if err != nil {
		return feeAmount, nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if merchantFee == nil {
		return feeAmount, defaultFeeDetail, nil
	}

	// LADDER tiering: resolve tier at transaction time
	if merchantFee.TieringModel != nil && *merchantFee.TieringModel == constant.LadderTieringModel {
		if result := s.resolveLadderTier(ctx, merchantFee, request); result != nil {
			merchantFee.AmountType = result.Tier.AmountType
			merchantFee.Amount = result.Tier.Amount
			merchantFee.Percentage = result.Tier.Percentage
			merchantFee.MaxFeeAmount = result.Tier.MaxFeeAmount
			merchantFee.TaxType = result.Tier.TaxType
			merchantFee.TaxPercentage = result.Tier.TaxPercentage

			defaultFeeDetail.LadderCounterKey = result.RedisKey
			defaultFeeDetail.LadderCounterIncrement = result.Increment
		}
	}

	// Build feeDetail from merchantFee
	feeDetail := defaultFeeDetail
	feeDetail.DeductionType = merchantFee.DeductionType
	feeDetail.AmountType = merchantFee.AmountType
	feeDetail.Amount = merchantFee.Amount
	feeDetail.Percentage = merchantFee.Percentage
	feeDetail.TaxType = merchantFee.TaxType
	feeDetail.TaxPercentage = merchantFee.TaxPercentage
	if merchantFee.MaxFeeAmount != nil {
		feeDetail.MaxFeeAmount = merchantFee.MaxFeeAmount
	}
	if merchantFee.DeductionDay != nil {
		feeDetail.DeductionDay = merchantFee.DeductionDay
	}
	if merchantFee.DeductionLastDate != nil {
		feeDetail.DeductionLastDate = merchantFee.DeductionLastDate
	}
	feeDetail.IsDefaultConfig = false

	// Calculate fee
	feeAmount, taxAmount = s.CalculateFee(ctx, request, feeDetail)
	if feeAmount < 0.0 {
		return feeAmount, nil, pkgErrors.New(response.HttpErrRequest, errors.New("fee amount cannot be negative"))
	}
	feeDetail.TaxAmount = taxAmount
	feeDetail.FinalAmount = feeAmount

	return feeAmount, feeDetail, nil
}

func (s *FeeService) GetTransactionFeeOnBehalf(ctx context.Context, request *feeModel.GetTrxFeeOnBehalfRequest) (*feeModel.TrxFeeOnBehalfMetadata, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/GetTransactionFeeOnBehalf")
	defer segment.End()

	config, err := s.merchantRepo.GetTransactionFeeOnBehalf(
		ctx, request.MerchantId, request.SubMerchantId, request.Reference, request.PaymentMethod, request.ReferenceType,
	)
	if err != nil {
		return nil, err

	} else if config == nil {
		return &feeModel.TrxFeeOnBehalfMetadata{
			Type: constant.FeeOnBehalfTypeNotSet, AmountType: "AMOUNT",
		}, nil
	}
	feeAmount, _ := s.CalculateFee(
		ctx,
		&feeModel.GetFeeRequest{
			ReferenceAmount: request.TransactionAmount,
		},
		&feeModel.FeeMetadataObject{
			AmountType: config.AmountType,
			Amount:     config.Amount,
			Percentage: config.Percentage,
		},
	)
	return &feeModel.TrxFeeOnBehalfMetadata{
		Reference:   config.Reference,
		Type:        config.Type,
		AmountType:  config.AmountType,
		Amount:      config.Amount,
		Percentage:  config.Percentage,
		FinalAmount: feeAmount,
	}, nil
}

func (s *FeeService) getDefaultMerchantFee(ctx context.Context, request *feeModel.GetFeeRequest) (result *merchantModel.MerchantFee, err error) {

	key := fmt.Sprintf(
		constant.NonPaymentFeeConfigsFmt, request.MerchantID, strings.ToLower(request.Reference),
	)
	result = &merchantModel.MerchantFee{}

	if err = s.redis.Get(ctx, key).Scan(result); err == nil {
		return

	} else if !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "getting merchant fee config from cache", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if result, err = s.merchantRepo.GetMerchantFeeByMerchantIDAndType(ctx, request.MerchantID, request.Reference); err != nil {
		return

	} else if result != nil {
		_ = s.redis.Set(ctx, key, result, 15*time.Minute).Err()
	}
	return
}

func (s *FeeService) getPaymentMerchantFee(ctx context.Context, request *feeModel.GetFeeRequest) (*merchantModel.MerchantFee, error) {
	return s.merchantRepo.DeterminePaymentFeeByMerchantIdMethodAndChannel(ctx, request)
}

func (s *FeeService) getPaymentFundedPayoutMerchantFee(ctx context.Context, request *feeModel.GetFeeRequest) (*merchantModel.MerchantFee, error) {
	return s.merchantRepo.DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel(ctx, request.MerchantID, request.PaymentMethod, request.Channel, request.SettlementMethod)
}

func (s *FeeService) getPayoutMerchantFee(ctx context.Context, request *feeModel.GetFeeRequest) (result *merchantModel.MerchantFee, err error) {

	key := fmt.Sprintf(
		constant.CacheKeyFmtPayoutTransactionFee, request.MerchantID, strings.ToLower(request.Channel),
	)
	result = &merchantModel.MerchantFee{}

	if err = s.redis.Get(ctx, key).Scan(result); err == nil {
		return
	}

	if result, err = s.merchantRepo.DeterminePayoutFeeByMerchantIdAndChannel(ctx, request.MerchantID, request.Channel, constant.ReferenceDisbursement); err != nil {
		s.logger.Error(ctx, "Failed while determine payout fee by merchant and channel", logger.Error(err))
		return
	}

	if result != nil {
		_ = s.redis.Set(ctx, key, result, 15*time.Minute)
	}
	return
}

func (s *FeeService) getPayoutVAMerchantFee(ctx context.Context, request *feeModel.GetFeeRequest) (result *merchantModel.MerchantFee, err error) {
	key := fmt.Sprintf(
		constant.CacheKeyFmtPayoutTransactionFee, request.MerchantID, strings.ToLower("va-"+request.Channel),
	)
	result = &merchantModel.MerchantFee{}

	if err = s.redis.Get(ctx, key).Scan(result); err == nil {
		return
	}

	// Try DISBURSEMENT_VA config first
	result, err = s.merchantRepo.DeterminePayoutFeeByMerchantIdAndChannel(ctx, request.MerchantID, request.Channel, constant.ReferenceDisbursementVA)
	if err != nil {
		s.logger.Error(ctx, "Failed while determine VA payout fee by merchant and channel", logger.Error(err))
		return
	}

	// Fall back to regular DISBURSEMENT fee if no VA-specific config
	if result == nil {
		result, err = s.merchantRepo.DeterminePayoutFeeByMerchantIdAndChannel(ctx, request.MerchantID, request.Channel, constant.ReferenceDisbursement)
		if err != nil {
			s.logger.Error(ctx, "Failed while getting fallback payout fee for VA", logger.Error(err))
			return
		}
	}

	if result != nil {
		_ = s.redis.Set(ctx, key, result, 15*time.Minute)
	}
	return
}

func (s *FeeService) getRefundMerchantFee(ctx context.Context, request *feeModel.GetFeeRequest) (*merchantModel.MerchantFee, error) {
	return s.merchantRepo.DetermineRefundFeeByMerchantIdAndReferenceType(ctx, request.MerchantID, request.ReferenceType)
}

func defaultFeeAmount(reference string) float64 {
	switch reference {
	case constant.TypeDisbursement:
		return 4_000

	case constant.ReferenceDisbursementVA:
		return 4_000

	case constant.ReferenceAccountInquiry:
		return 450

	case constant.TypePayment:
		return 4_000

	case constant.TypeXB:
		return 0

	case constant.ReferencePlatformActivity:
		return 10_000

	case constant.ReferencePlatformTransfer:
		return 0

	case constant.ReferencePlatformTransaction:
		return 0

	case constant.ReferenceWallet:
		return 0
	case constant.ReferenceTopUp:
		return 0
	case constant.ReferencePaymentFundedPayout:
		return 0

	default:
		return 4_000
	}
}

func defaultFeeAmountForPaymentUsecase(request *feeModel.GetFeeRequest) (feeAmount float64, amountType string, amount, percentage float64) {
	switch request.PaymentMethod {
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		amountType = constant.MerchantFeeAmountType
		amount = 4_000
		percentage = 0

	case paymentConstant.PAYMENT_METHOD_QRIS:
		amountType = constant.MerchantFeePercentageType
		amount = 0
		percentage = 0.7

	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
		amountType = constant.MerchantFeeAmountPercentageType

		defaultFee, ok := config.GetCreditCardReferences().DefaultFee.CustomChannel[strings.ToLower(request.Channel)]
		if !ok {
			defaultFee = config.GetCreditCardReferences().DefaultFee.OtherChannel
		}
		amount = defaultFee.Amount
		percentage = defaultFee.Percentage

	case paymentConstant.PAYMENT_METHOD_VIRTUAL_TERMINAL:
		def := config.GetCreditCardReferences().DefaultVirtualTerminalFee
		amountType = def.Type
		amount = def.Amount
		percentage = def.Percentage

	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD_SPLIT_PAYMENT:
		def := config.GetCreditCardReferences().DefaultSplitPaymentFee
		amountType = def.Type
		amount = def.Amount
		percentage = def.Percentage

	case paymentConstant.PAYMENT_METHOD_EWALLET:
		feeDefault := config.GetPaymentEWalletFeeDefault(request.Channel)
		amountType, amount, percentage = feeDefault.Type, feeDefault.Amount, feeDefault.Percentage

	case paymentConstant.PAYMENT_METHOD_INSTALLMENT:
		amountType = constant.MerchantFeeAmountPercentageType
		channelFeeConfig, exist := config.GetInstallmentDefaultChannelFeeConfig()[request.Channel]
		if !exist {
			channelFeeConfig = config.GetInstallmentDefaultFeeConfig()
		}
		amount = channelFeeConfig.Amount
		percentage = channelFeeConfig.Percentage

	default:
		return 4_000, constant.MerchantFeeAmountType, 4_000, 0
	}

	feeAmount = ((percentage / 100) * request.ReferenceAmount) + amount
	return feeAmount, amountType, amount, percentage
}

func defaultFeeAmountForPaymentFundedPayoutUsecase(settlementType string, referenceAmount float64) (feeAmount float64, amountType string, amount, percentage float64) {
	defMap := config.GetCreditCardReferences().DefaultCardFundedPayout
	def, ok := defMap[strings.ToLower(settlementType)]
	if !ok {
		def = defMap[strings.ToLower(constant.PaymentSettlementMethodStandard)]
	}

	amountType = def.Type
	amount = def.Amount
	percentage = def.Percentage

	feeAmount = ((percentage / 100) * referenceAmount) + amount
	return feeAmount, amountType, amount, percentage
}

func (s *FeeService) getMerchantFeeXB(ctx context.Context, request *feeModel.GetFeeRequest) (result *merchantModel.MerchantFee, err error) {
	reference := request.Reference + "-" + request.Channel
	key := fmt.Sprintf(
		constant.CacheKeyFmtPayoutTransactionFee, request.MerchantID, strings.ToLower(reference),
	)
	result = &merchantModel.MerchantFee{}

	if err = s.redis.Get(ctx, key).Scan(result); err == nil {
		return

	} else if !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "getting merchant fee config from cache", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	result, err = s.merchantRepo.GetMerchantFeeXB(ctx, &merchantModel.MerchantFeeXBQuery{
		MerchantID: request.MerchantID,
		Reference:  request.Reference,
		Channel:    request.Channel,
	})

	if err != nil {
		return
	} else if result != nil {
		_ = s.redis.Set(ctx, key, result, 15*time.Minute).Err()
	}
	if result == nil {
		// set default fee base on channel
		defaultFee := 0.0
		switch strings.ToUpper(request.Channel) {
		case constant.XBRoutingCodeLocal:
			defaultFee = s.config.XbCoreProcessorConfig.DefaultLocalFee
		case constant.XBRoutingCodeSwift:
			defaultFee = s.config.XbCoreProcessorConfig.DefaultSwiftFee
		}
		result = &merchantModel.MerchantFee{
			MerchantID: request.MerchantID,
			Reference:  request.Reference,
			Channel:    &request.Channel,
			Percentage: 0.0,
			AmountType: constant.MerchantFeeAmountType,
			Amount:     defaultFee,
			TaxType:    constant.MerchantTaxTypeNonPKP,
		}
	}

	return
}

func (s *FeeService) GetXbFeeConfigs(ctx context.Context, merchantID string) (*merchantModel.XbFeeConfigResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/GetXbFeeConfigs")
	defer segment.End()

	localFee, err := s.getMerchantFeeXB(ctx, &feeModel.GetFeeRequest{
		MerchantID: merchantID,
		Reference:  constant.ReferenceXB,
		Channel:    constant.XBRoutingCodeLocal,
	})
	if err != nil {
		return nil, err
	}

	swiftFee, err := s.getMerchantFeeXB(ctx, &feeModel.GetFeeRequest{
		MerchantID: merchantID,
		Reference:  constant.ReferenceXB,
		Channel:    constant.XBRoutingCodeSwift,
	})
	if err != nil {
		return nil, err
	}

	resp := &merchantModel.XbFeeConfigResponse{}
	if localFee != nil {
		resp.Local = localFee.ToResponse()
	}
	if swiftFee != nil {
		resp.Swift = swiftFee.ToResponse()
	}
	return resp, nil
}
