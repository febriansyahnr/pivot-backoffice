package merchant

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) UpdateFeeTieringConfig(ctx context.Context, request *merchant.FeeTieringRequest) (*merchant.FeeTieringResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateFeeTieringConfig")
	defer segment.End()

	if request == nil || len(request.Configs) == 0 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("make sure the data sent is not empty"))
	}

	// Default to MONTHLY_ASSESSED for backward compatibility
	if request.Model == "" {
		request.Model = constant.MonthlyAssessedTieringModel
	}

	merchantFee, err := s.repo.GetMerchantFeeByID(ctx, request.FeeId)
	if err != nil {
		s.logger.Error(ctx, "Get merchant fee by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if merchantFee == nil || merchantFee.MerchantID != request.MerchantId {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	sort.Slice(request.Configs, func(i, j int) bool {
		return request.Configs[i].Tier < request.Configs[j].Tier
	})

	if request.Configs[0].Min != 0 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("minimum value of tier 1 must be 0"))
	}
	length, exampleAmount := len(request.Configs), 10_000.00
	for i := 0; i < length; i++ {
		if request.Configs[i].Tier != i+1 {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrFeeTieringSequence)
		}

		if err := request.Configs[i].Validate(*merchantFee); err != nil {
			return nil, err
		}

		if i == 0 {
			continue

		} else if request.Configs[i-1].Max+1 != request.Configs[i].Min {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrFeeTieringRange)
		}

		feeAmountBefore, _ := s.feeSvc.CalculateFee(ctx, &feeModel.GetFeeRequest{ReferenceAmount: exampleAmount}, &feeModel.FeeMetadataObject{
			AmountType:    request.Configs[i-1].AmountType,
			Amount:        request.Configs[i-1].Amount,
			MaxFeeAmount:  request.Configs[i-1].MaxFeeAmount,
			Percentage:    request.Configs[i-1].Percentage,
			TaxType:       request.Configs[i-1].TaxType,
			TaxPercentage: request.Configs[i-1].TaxPercentage,
		})
		feeAmountAfter, _ := s.feeSvc.CalculateFee(ctx, &feeModel.GetFeeRequest{ReferenceAmount: exampleAmount}, &feeModel.FeeMetadataObject{
			AmountType:    request.Configs[i].AmountType,
			Amount:        request.Configs[i].Amount,
			MaxFeeAmount:  request.Configs[i].MaxFeeAmount,
			Percentage:    request.Configs[i].Percentage,
			TaxType:       request.Configs[i].TaxType,
			TaxPercentage: request.Configs[i].TaxPercentage,
		})
		if feeAmountAfter >= feeAmountBefore {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("fee amount tier %d is greater than or equal tier %d", request.Configs[i].Tier, request.Configs[i-1].Tier))
		}
		if i+1 == length {
			request.Configs[i].Max = 999_999_999_999_999
		}
	}

	// LADDER tiering resolves at transaction time, so AppliedTier is not allowed
	if request.Model == constant.LadderTieringModel && request.AppliedTier > 0 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("appliedTier is not supported for LADDER tiering model"))
	}

	if request.AppliedTier > 0 {
		ok := slices.ContainsFunc(request.Configs, func(r merchant.FeeTieringConfig) (ok bool) {
			if ok = (r.Tier == request.AppliedTier); ok {
				request.AppliedFee = &r
			}
			return
		})
		if !ok {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("tier to be applied was not found"))
		}
	}

	if err := s.repo.UpdateFeeTieringConfig(ctx, request); err != nil {
		s.logger.Error(ctx, "Update fee tiering config", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	// Invalidate cached fee data so subsequent transactions use the updated tiering config
	if merchantFee.Reference != constant.ReferencePayment {
		keys := []string{fmt.Sprintf(constant.NonPaymentFeeConfigsFmt, request.MerchantId, strings.ToLower(merchantFee.Reference))}

		if merchantFee.Reference == constant.ReferenceDisbursement {
			if merchantFee.Channel != nil {
				keys = append(keys, fmt.Sprintf(constant.CacheKeyFmtPayoutTransactionFee, merchantFee.MerchantID, strings.ToLower(*merchantFee.Channel)))
			} else {
				payoutKeys, _ := s.redis.Keys(ctx, fmt.Sprintf(constant.CacheKeyFmtPayoutTransactionFee, merchantFee.MerchantID, "*")).Result()
				keys = append(keys, payoutKeys...)
			}
		}

		if err = s.redis.Del(ctx, keys...).Err(); err != nil {
			s.logger.Error(ctx, "delete merchant fee config from cache", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, err)
		}
	}
	return &merchant.FeeTieringResponse{
		MerchantId:    merchantFee.MerchantID,
		Reference:     merchantFee.Reference,
		PaymentMethod: merchantFee.PaymentMethod,
		DeductionType: merchantFee.DeductionType,
		Model:         request.Model,
		Type:          request.Type,
		AppliedTier:   request.AppliedTier,
		Configs:       request.Configs,
	}, nil
}
