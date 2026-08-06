package merchant

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disburmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) SetCustomLimitConfig(ctx context.Context, request merchant.BeneficiaryLimitConfigRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/merchant/SetCustomLimitConfig")
	defer span.End()

	// set merchant-policy limit, if not set beneficiary-account
	if request.BeneficiaryBankCode == "" || request.BeneficiaryAccountNo == "" {
		// Find merchant by ID
		merchant, err := s.repo.FindMerchantByID(ctx, request.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "error to find merchant by ID whe set custom limit config", pdkLogger.Error(err))
			return pkgErr.New(response.HttpErrDatabase, err)
		} else if merchant == nil {
			s.logger.Info(ctx, "merchant not found whe set custom limit config")
			return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
		}

		// Set merchant-policy limit
		merchant.UpdatedAt = time.Now().UTC()
		_ = merchant.UpdateBeneficiaryPayoutLimitRule(request.BeneficiaryPayoutLimitRule)
		if err = s.repo.Update(ctx, merchant); err != nil {
			s.logger.Error(ctx, "error to update custom limit config", pdkLogger.Error(err))
			return pkgErr.New(response.HttpErrDatabase, err)
		}

		return nil
	}

	// get or inquiry beneficiary account
	beneficiary, err := s.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryModel.CheckAccountRequest{
		MerchantID:           request.MerchantID,
		BeneficiaryBankCode:  request.BeneficiaryBankCode,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
	})
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if beneficiary == nil {
		return pkgErr.New(response.HttpStatusErrorUnprocessableContent, constant.ErrDataNotFound)
	}

	// Set up custom limit
	if request.BeneficiaryPayoutLimitRule != nil {
		beneficiary.MetadataObj.BeneficiaryPayoutLimitRule = &disburmentModel.BeneficiaryPayoutLimitRuleConfig{
			Velocity:        request.BeneficiaryPayoutLimitRule.Velocity,
			Timeframe:       request.BeneficiaryPayoutLimitRule.Timeframe,
			AmountThreshold: request.BeneficiaryPayoutLimitRule.AmountThreshold,
		}
	} else {
		beneficiary.MetadataObj.BeneficiaryPayoutLimitRule = nil
	}

	// Parse to request object
	beneficiaryUpdateRequest := beneficiary.ToBeneficiaryAccount()
	beneficiaryUpdateRequest.Metadata = types.NullJSONText{
		Valid: true,
	}
	beneficiaryUpdateRequest.Metadata.JSONText, _ = json.Marshal(beneficiary.MetadataObj)

	// update the metadata for beneficiary account
	if err = s.beneficiaryRepo.Update(ctx, beneficiaryUpdateRequest); err != nil {
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		return pkgErr.New(response.HttpErrDatabase, err)
	}

	return nil
}
