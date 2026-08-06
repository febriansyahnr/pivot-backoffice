package merchant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) UpdateMerchantFee(ctx context.Context, request *merchant.UpdateMerchantFeeRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateMerchantFee")
	defer segment.End()

	existedFee, err := s.repo.GetMerchantFeeByID(ctx, request.ID)
	if err != nil {
		s.logger.Error(ctx, "failed to find merchant fee data", logger.Error(err))
		return pkgErr.New(response.HttpErrInternal, errors.New("unexpected error"))

	} else if existedFee == nil {
		s.logger.Error(ctx, "unable to find existing merchant fee data", logger.Any("id", request.ID))
		return pkgErr.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	if existedFee, err = existedFee.UpdateMerchantFee(request); err != nil {
		return err
	}

	if err = s.repo.UpdateMerchantFee(ctx, existedFee); err != nil {
		s.logger.Error(ctx, "failed to update merchant fee data", logger.Error(err))
		return pkgErr.New(response.HttpErrInternal, constant.ErrNoRowsAffected)
	}

	if existedFee.Reference != constant.ReferencePayment {

		keys := []string{fmt.Sprintf(constant.NonPaymentFeeConfigsFmt, request.MerchantID, strings.ToLower(existedFee.Reference))}
		if existedFee.Reference == constant.ReferenceDisbursement {
			if existedFee.Channel != nil {
				keys = []string{fmt.Sprintf(constant.CacheKeyFmtPayoutTransactionFee, existedFee.MerchantID, strings.ToLower(*existedFee.Channel))}

			} else {
				keys, _ = s.redis.Keys(ctx, fmt.Sprintf(constant.CacheKeyFmtPayoutTransactionFee, existedFee.MerchantID, "*")).Result()
			}
		}

		if len(keys) > 0 {
			if err = s.redis.Del(ctx, keys...).Err(); err != nil {
				s.logger.Error(ctx, "delete merchant fee config from cache", logger.Error(err))
				return pkgErr.New(response.HttpErrDatabase, err)
			}
		}
	}
	return nil
}
