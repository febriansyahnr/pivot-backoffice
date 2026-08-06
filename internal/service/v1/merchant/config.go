package merchant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) TransactionConfig(ctx context.Context, merchantId string, config *merchantModel.TransactionConfigs) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/TransactionConfig")
	defer segment.End()

	if err := s.validateTransactionConfig(ctx, merchantId, config); err != nil {
		return err
	}

	_ = s.redis.Del(
		ctx,
		fmt.Sprintf(constant.DisbursementTransactionConfigFmt, merchantId),
		fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt, merchantId, constant.DisbursementDailyLimitMerchant),
		fmt.Sprintf(constant.DailyDisbursementTransactionConfigFmt, merchantId, constant.DisbursementDailyLimitMerchantPlatform),
	)
	return s.repo.UpdateTransactionConfig(ctx, merchantId, config)
}

func (s *MerchantService) GetTransactionConfig(ctx context.Context, merchantId string) (*merchantModel.TransactionConfigResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/TransactionConfig")
	defer segment.End()

	result, err := s.repo.GetTransactionConfig(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Get transaction config", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)

	} else if result == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}
	return result, nil
}

func (s *MerchantService) GetMerchantIdForConfigs(next context.Context, merchantId string, setMerchantCtx bool) (context.Context, *merchantModel.MerchantIdForConfigs, error) {
	ctx, segment := otelTracer.Start(next, "internal/service/v1/merchant/GetMerchantIdForConfigs")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
		return next, nil, pkgErr.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return next, nil, pkgErr.New(response.HttpErrRequest, constant.ErrMerchantNotFound)
	}

	if setMerchantCtx {
		next = context.WithValue(next, constant.CtxMerchantData, merchant)
	}

	configs := &merchantModel.MerchantIdForConfigs{
		MerchantType:              constant.MerchantTypeMerchant,
		MerchantTransactionConfig: merchantId,
	}

	if merchant.ParentID.String != "" {
		configs.MerchantType = constant.MerchantTypeSubMerchant

		next = context.WithValue(next, constant.CtxParentMerchantId, merchant.ParentID.String)

		if merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			configs.MerchantTransactionConfig = merchant.ParentID.String
		}
	}
	return next, configs, nil
}

func (s *MerchantService) FDSConfig(ctx context.Context, merchantID string, config merchantModel.FDSConfigRequest) (*merchantModel.FDSConfigResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/FDSConfig")
	defer segment.End()

	if _, err := s.validateMerchantForConfig(ctx, merchantID, "FDS configuration"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateFDSConfig(ctx, merchantID, config); err != nil && !errors.Is(err, constant.ErrNoRowsAffected) {
		s.logger.Error(ctx, "Failed to update FDS configuration", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return &merchantModel.FDSConfigResponse{FDSConfig: config.FDSConfig}, nil
}

func (s *MerchantService) GetFDSConfig(ctx context.Context, merchantID string) (*merchantModel.GetFDSConfigResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetFDSConfig")
	defer segment.End()

	config, err := s.repo.GetFDSConfig(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch merchant configuration", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if config == nil {
		return nil, pkgErr.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}
	return config, nil
}

func (s *MerchantService) validateTransactionConfig(ctx context.Context, merchantId string, config *merchantModel.TransactionConfigs) error {
	merchant, err := s.repo.FindMerchantByID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Find merchant by id", logger.Error(err))
		return pkgErr.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return pkgErr.New(response.HttpErrRequest, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.Valid && merchant.KYCStatus.String == constant.KYCStatusNotRequired {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("transaction config applies only to platform merchants, merchants, and KYC sub merchants"))

	} else if config.DailyDisbursement == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("daily transaction limit must be set"))
	}

	if err = config.Validate(); err != nil {
		return pkgErr.New(response.HttpErrRequest, err)
	}

	if config.DailyDisbursement.Merchant < config.Disbursement.MaxAmount {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("daily transaction limit (merchant) must be gte to the max transaction"))
	}

	platform, err := s.productRepo.GetMerchantSelectedProductByName(ctx, merchantId, constant.ProductPlatform)
	if err != nil {
		s.logger.Error(ctx, "Get merchant selected product by name", logger.Error(err))
		return pkgErr.New(response.HttpErrDatabase, err)

	} else if platform == nil && config.DailyDisbursement.MerchantPlatform != nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("merchant does not activate platform product"))

	} else if platform != nil && config.DailyDisbursement.MerchantPlatform == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("merchant platform daily transaction limit must be set"))

	}

	if dailyPlatform := config.DailyDisbursement.MerchantPlatform; dailyPlatform != nil && *dailyPlatform < config.Disbursement.MaxAmount {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("daily transaction limit (merchant platform) must be gte to the max transaction"))
	}
	return nil
}

func (s *MerchantService) UpdateSettlementConfig(ctx context.Context, merchantFeeId string, config *merchantModel.SettlementConfig) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateSettlementConfig")
	defer segment.End()

	if config == nil {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid config"))
	}

	// check if merchant fee is existed
	existedFee, err := s.repo.GetMerchantFeeByID(ctx, merchantFeeId)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)

	} else if existedFee == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}
	if existedFee.SettlementMethod != nil && *existedFee.SettlementMethod == constant.SettlementMethodInstant && util.IsValidSettlementTime(config.Type) {
		return pkgErr.New(response.HttpErrUnprocessableContent, errors.New("invalid settlement config type for instant settlement"))
	}

	// Assign config to settlement
	existedFee.SettlementConfigsObj = *config
	existedFee.SettlementConfigs.Valid = true
	existedFee.SettlementConfigs.JSONText, _ = json.Marshal(existedFee.SettlementConfigsObj)

	err = s.repo.UpdateMerchantFee(ctx, existedFee)
	if err != nil && !errors.Is(err, constant.ErrNoRowsAffected) {
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	return nil
}

func (s *MerchantService) PaymentInvestigationConfig(ctx context.Context, merchantID string, config merchantModel.PaymentInvestigationConfigRequest) (*merchantModel.PaymentInvestigationConfigResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/PaymentInvestigationConfig")
	defer segment.End()

	merchant, err := s.validateMerchantForConfig(ctx, merchantID, "Payment investigation config")
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePaymentInvestigationConfig(ctx, merchantID, config); err != nil {
		s.logger.Error(ctx, "Failed to update payment investigation config", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return &merchantModel.PaymentInvestigationConfigResponse{
		MerchantID:          merchantID,
		MerchantName:        merchant.Name,
		Enabled:             config.Enabled,
		PivotPercentageLoss: config.PivotPercentageLoss,
		PivotMaxLoss:        config.PivotMaxLoss,
	}, nil
}

func (s *MerchantService) validateMerchantForConfig(ctx context.Context, merchantID, configName string) (*merchantModel.Merchant, error) {
	merchant, err := s.repo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "Failed to retrieve merchant details by ID", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return nil, pkgErr.New(response.HttpErrRequest, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.Valid && merchant.KYCStatus.String == constant.KYCStatusNotRequired {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("%s can only be used by merchants or KYC sub-merchants", configName))
	}
	return merchant, nil
}
