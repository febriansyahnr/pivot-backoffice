package reportingService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	reportingModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) UpsertBalanceHistory(ctx context.Context, request *reportingModel.UpsertBalanceHistoryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reporting/UpsertBalanceHistory")
	defer segment.End()

	payload := request.Event.GetCurrent()

	if request.ShouldExcludeEvent() {
		s.logger.Info(ctx, "Event not processed: data does not meet criteria", logger.String("uuid", payload.UUID), logger.String("reference_id", payload.ReferenceID))
		return nil
	}

	account, err := s.accountRepo.GetByUUID(ctx, util.ParseUUID(payload.AccountID))
	if err != nil {
		s.logger.Error(ctx, "Failed to get account details", logger.Error(err))
		return err
	}

	if account == nil {
		s.logger.Info(ctx, fmt.Sprintf("Account ID %s not found. Balance history report generation skipped for this transaction", payload.AccountID))
		return nil
	} else if account.UserType == constant.UserTypeCustomer {
		s.logger.Info(ctx, "Event skipped because it is a customer wallet transaction")
		return nil
	}

	isHardDelete := request.Event.IsDelete() ||
		(request.Event.IsUpdate() && payload.Status != constant.StatusSuccess)
	if isHardDelete {
		return s.repo.HardDeleteBalanceHistory(ctx, payload.UUID)
	}

	if request.Event.IsUpdate() && payload.DeletedAt != nil {
		return s.repo.SoftDeleteBalanceHistory(ctx, payload.UUID, time.UnixMilli(request.Event.TsMs))
	}

	data := request.ToCreateBalanceHistory()
	previousData := request.Event.GetPrevious()

	isUpdateSettlementBalance := request.Event.IsUpdate() &&
		previousData.Status == constant.StatusSuccess &&
		util.ValueOfPtr(previousData.SettlementStatus) == constant.StatusPending &&
		util.ValueOfPtr(payload.SettlementStatus) == constant.StatusSuccess
	if isUpdateSettlementBalance {
		if err := s.repo.UpdateSettlementBalanceHistory(ctx, data); err == nil {
			return nil
		} else if !errors.Is(err, constant.ErrNoRowsAffected) {
			s.logger.Error(ctx, "Failed when update settlement balance history data", logger.Error(err))
			return err
		}
	}

	if err := s.repo.PrepareAdvancedBalanceHistoryData(ctx, &data); err != nil {
		s.logger.Error(ctx, "Failed when prepare advanced balance history data", logger.Error(err))
		return err
	}
	return s.repo.UpsertBalanceHistory(ctx, data)
}

func (s *service) ListBalanceHistory(ctx context.Context, filters *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reporting/ListBalanceHistory")
	defer segment.End()

	result, err := s.repo.ListBalanceHistory(ctx, filters, page, perPage)
	if err != nil {
		s.logger.Error(ctx, "Failed when list balance history via data reporting", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return result, nil
}
