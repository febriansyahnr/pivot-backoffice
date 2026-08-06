package merchant

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) GetDepositSetting(ctx context.Context, merchantId string) (result *merchant.DepositSettingResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetDepositSetting")
	defer segment.End()

	if result, err = s.repo.GetDepositSetting(ctx, merchantId); err != nil {
		s.logger.Error(ctx, "Failed when fetch deposit setting", logger.Error(err))
	}
	return
}

func (s *MerchantService) SetAutoWithdrawal(ctx context.Context, request *merchant.AutoWithdrawalSettingRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/SetAutoWithdrawal")
	defer segment.End()

	if exists, err := s.bankAccountRepo.BankAccountHasBeenPrepared(ctx, request.MerchantId); err != nil {
		return err

	} else if !exists {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found"))
	}

	if err = s.repo.SetAutoWithdrawal(ctx, request); err != nil {
		s.logger.Error(ctx, "Failed when set auto withdrawal status (on/off)", logger.Error(err))
	}

	activity := constant.ActivityAutoWithdrawalSetOFF
	if request.Status == "ON" {
		activity = constant.ActivityAutoWithdrawalSetON
	}
	_ = s.rabbitMqExt.PublishActivity(
		ctx, &request.MerchantId, &request.UserId, constant.TagSetAutoWithdrawalStatus, activity, map[string]interface{}{
			"success": err == nil,
		},
	)
	return
}
